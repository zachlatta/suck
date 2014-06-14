package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"

	"github.com/jmcvetta/neoism"

	"code.google.com/p/go.net/html"
)

var (
	db           *neoism.Database
	requestsMade = 0
	urlsFound    = 0
	urls         = NewURLMap()
)

var regexpURL = regexp.MustCompile(`(?i)\b((?:[a-z][\w-]+:(?:/{1,3}|[a-z0-9%])|www\d{0,3}[.]|[a-z0-9.\-]+[.][a-z]{2,4}/)(?:[^\s()<>]+|\(([^\s()<>]+|(\([^\s()<>]+\)))*\))+(?:\(([^\s()<>]+|(\([^\s()<>]+\)))*\)|[^\s!()\[\]{};:'".,<>?«»“”‘’]))`)

type result struct {
	url           *url.URL
	referer       *url.URL
	err           error
	statusCode    int
	duration      time.Duration
	contentLength int64
}

type request struct {
	*http.Request
	referer *url.URL
}

type Sucker struct {
	ConcurrencyLevel int
	jobs             chan *request
	results          chan *result
}

func (s *Sucker) Run(firstURL string) {
	s.results = make(chan *result)
	s.run(firstURL)
}

func (s *Sucker) run(firstURL string) {
	var wg sync.WaitGroup
	wg.Add(s.ConcurrencyLevel)

	s.jobs = make(chan *request)

	for i := 0; i < s.ConcurrencyLevel; i++ {
		go func() {
			s.worker(s.jobs)
			wg.Done()
		}()
	}

	for i := 0; i < 4; i++ {
		go func() {
			for res := range s.results {

				referer, e := "", ""

				if res.referer != nil {
					referer = res.referer.String()
				}

				if res.err != nil {
					e = res.err.Error()
				}

				node, err := db.CreateNode(neoism.Props{
					"url":           res.url.String(),
					"referer":       referer,
					"err":           e,
					"statusCode":    res.statusCode,
					"duration":      res.duration,
					"contentLength": res.contentLength,
				})
				if err != nil {
					log.Println(err)
				}
				node.AddLabel("Website")

				fmt.Printf("COUNT: %d, LINKS: %d, REQ RECEIVED: %s\n", requestsMade, urlsFound, res.url.String())
			}
		}()
	}

	req, err := http.NewRequest("GET", firstURL, nil)
	if err != nil {
		panic(err)
	}

	s.jobs <- &request{req, nil}

	wg.Wait()
}

func (s *Sucker) worker(ch chan *request) {
	client := &http.Client{}
	for req := range ch {
		requestsMade++
		start := time.Now()
		resp, err := client.Do(req.Request)
		if err != nil {
			fmt.Println(err)
		}
		code := 0
		size := int64(-1)
		if resp != nil {
			if resp.ContentLength > 0 {
				size = resp.ContentLength
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
			}

			reader := bytes.NewReader(body)

			go func(body io.Reader, base *url.URL) {
				links := links(body)
				for _, link := range links {
					link = base.ResolveReference(link)

					if !urls.Exists(link.String()) {
						urlsFound++
						urls.Add(link.String())
						nextReq, err := http.NewRequest("GET", link.String(), nil)
						if err != nil {
							fmt.Println(err)
						}

						s.jobs <- &request{nextReq, req.URL}
					}
				}
			}(reader, req.URL)

			resp.Body.Close()
		}

		s.results <- &result{
			url:           req.URL,
			referer:       req.referer,
			statusCode:    code,
			duration:      time.Now().Sub(start),
			err:           err,
			contentLength: size,
		}
	}
}

func links(reader io.Reader) []*url.URL {
	links := []*url.URL{}
	z := html.NewTokenizer(reader)
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			return links
		}

		for _, attr := range z.Token().Attr {
			if attr.Key == "href" {
				url, err := url.Parse(attr.Val)
				if err == nil {
					links = append(links, url)
				}
			}
		}
	}
}

func main() {
	firstURL := flag.String("first", "http://wikipedia.org", "first url to scrape")
	dbURL := flag.String("db", "http://localhost:7474/db/data", "url to neo4j")
	flag.Parse()

	var err error
	db, err = neoism.Connect(*dbURL)
	if err != nil {
		log.Fatal(err)
	}

	sucker := Sucker{
		ConcurrencyLevel: 64,
	}

	sucker.Run(*firstURL)
}
