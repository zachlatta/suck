package main

import "sync"

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
