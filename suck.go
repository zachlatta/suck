package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"

	"crypto/md5"

	"code.google.com/p/go.net/html"
)

var (
	requestsMade = 0
	urlsFound    = 0
	urls         = NewURLMap()
)

type URLMap struct {
	urlMap map[string]bool
	mutex  *sync.Mutex
}

func NewURLMap() *URLMap {
	return &URLMap{
		urlMap: map[string]bool{},
		mutex:  &sync.Mutex{},
	}
}

func (u *URLMap) Exists(url string) bool {
	u.mutex.Lock()
	exists := u.urlMap[url]
	u.mutex.Unlock()
	return exists
}

func (u *URLMap) Add(url string) {
	u.mutex.Lock()
	u.urlMap[url] = true
	u.mutex.Unlock()
}

var regexpURL = regexp.MustCompile(`(?i)\b((?:[a-z][\w-]+:(?:/{1,3}|[a-z0-9%])|www\d{0,3}[.]|[a-z0-9.\-]+[.][a-z]{2,4}/)(?:[^\s()<>]+|\(([^\s()<>]+|(\([^\s()<>]+\)))*\))+(?:\(([^\s()<>]+|(\([^\s()<>]+\)))*\)|[^\s!()\[\]{};:'".,<>?«»“”‘’]))`)

type result struct {
	err           error
	statusCode    int
	duration      time.Duration
	contentLength int64
}

type Sucker struct {
	ConcurrencyLevel int
	jobs             chan *http.Request
	results          chan *result
}

func (s *Sucker) Run() {
	s.results = make(chan *result, 99999999)
	s.run()
}

func (s *Sucker) run() {
	var wg sync.WaitGroup
	wg.Add(s.ConcurrencyLevel)

	s.jobs = make(chan *http.Request)

	for i := 0; i < s.ConcurrencyLevel; i++ {
		go func() {
			s.worker(s.jobs)
			wg.Done()
		}()
	}

	req, err := http.NewRequest("GET", "http://wikipedia.org", nil)
	if err != nil {
		panic(err)
	}

	s.jobs <- req

	wg.Wait()
}

func (s *Sucker) worker(ch chan *http.Request) {
	client := &http.Client{}
	for req := range ch {
		requestsMade++
		fmt.Printf("COUNT: %d, LINKS: %d, REQ RECEIVED: %s\n", requestsMade, urlsFound, req.URL.String())
		start := time.Now()
		resp, err := client.Do(req)
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

			h := md5.New()
			io.WriteString(h, req.URL.String())

			err = ioutil.WriteFile(fmt.Sprintf("%x", h.Sum(nil)), body, 0644)
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
						req, err = http.NewRequest("GET", link.String(), nil)
						if err != nil {
							fmt.Println(err)
						}

						s.jobs <- req
					}
				}
			}(reader, req.URL)

			resp.Body.Close()
		}

		s.results <- &result{
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
	sucker := Sucker{
		ConcurrencyLevel: 64,
	}

	sucker.Run()
}
