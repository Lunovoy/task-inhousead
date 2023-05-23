package main

import (
	"fmt"
	"math"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type Site struct {
	Name          string
	Availability  bool
	ResponseTime  time.Duration
	LastCheckedAt time.Time
}

type SiteChecker struct {
	Sites       []*Site
	Concurrency int
	RequestLogs map[string]int
}

func (sc *SiteChecker) checkAvailability(site *Site, wg *sync.WaitGroup) {
	defer wg.Done()

	startTime := time.Now()

	response, err := http.Get(site.Name)
	if err != nil {
		site.Availability = false
		site.ResponseTime = math.MaxInt64
	} else {
		site.Availability = true
		site.ResponseTime = time.Since(startTime)
		response.Body.Close()
	}

	site.LastCheckedAt = time.Now()
}

func (sc *SiteChecker) checkAllSites() {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, sc.Concurrency)

	for _, site := range sc.Sites {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(site *Site) {
			defer func() {
				<-semaphore
			}()

			sc.checkAvailability(site, &wg)
		}(site)
	}

	wg.Wait()
}

func (sc *SiteChecker) getRequestLogs() map[string]int {
	requestLogsCopy := make(map[string]int)
	for k, v := range sc.RequestLogs {
		requestLogsCopy[k] = v
	}

	return requestLogsCopy
}

func (sc *SiteChecker) updateRequestLogs(endpoint string) {
	sc.RequestLogs[endpoint]++
}

func main() {

	port := 8080

	siteChecker := &SiteChecker{
		Sites: []*Site{
			{Name: "https://google.com"},
			{Name: "https://youtube.com"},
			{Name: "https://facebook.com"},
			{Name: "https://wikipedia.org"},
			{Name: "https://qq.com"},
			{Name: "https://taobao.com"},
			{Name: "https://yahoo.com"},
			{Name: "https://tmall.com"},
			{Name: "https://amazon.com"},
			{Name: "https://google.co.in"},
			{Name: "https://twitter.com"},
			{Name: "https://sohu.com"},
			{Name: "https://jd.com"},
			{Name: "https://live.com"},
			{Name: "https://instagram.com"},
			{Name: "https://sina.com.cn"},
			{Name: "https://weibo.com"},
			{Name: "https://google.co.jp"},
			{Name: "https://reddit.com"},
			{Name: "https://vk.com"},
			{Name: "https://login.tmall.com"},
			{Name: "https://blogspot.com"},
			{Name: "https://yandex.ru"},
			{Name: "https://netflix.com"},
		},
		Concurrency: runtime.NumCPU(),
		RequestLogs: make(map[string]int),
	}

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			siteChecker.checkAllSites()
		}
	}()

	http.HandleFunc("/site", func(w http.ResponseWriter, r *http.Request) {
		siteName := r.URL.Query().Get("name")
		for _, site := range siteChecker.Sites {
			if site.Name == siteName {
				siteChecker.updateRequestLogs("/site")
				fmt.Fprintf(w, "Site: %s, Availability: %v, Response Time: %v", site.Name, site.Availability, site.ResponseTime)
				return
			}
		}
	})

	http.HandleFunc("/min", func(w http.ResponseWriter, r *http.Request) {
		minResponseTime := time.Duration(time.Duration(math.MaxInt64))
		var minSite *Site

		for _, site := range siteChecker.Sites {
			if minSite == nil || site.ResponseTime < minResponseTime {
				minResponseTime = site.ResponseTime
				minSite = site
			}
		}

		if minSite != nil {
			siteChecker.updateRequestLogs("/min")
			fmt.Fprintf(w, "Site with minimum response time: %s, Response Time: %v", minSite.Name, minSite.ResponseTime)
		}
	})

	http.HandleFunc("/max", func(w http.ResponseWriter, r *http.Request) {
		maxResponseTime := time.Duration(0)
		var maxSite *Site

		for _, site := range siteChecker.Sites {
			if site.ResponseTime > maxResponseTime && site.Availability == true {
				maxResponseTime = site.ResponseTime
				maxSite = site
			}
		}

		if maxSite != nil {
			siteChecker.updateRequestLogs("/max")
			fmt.Fprintf(w, "Site with maximum response time: %s, Response Time: %v", maxSite.Name, maxSite.ResponseTime)
		}
	})

	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		requestLogs := siteChecker.getRequestLogs()
		fmt.Fprintf(w, "Request statistics:\n")
		for endpoint, count := range requestLogs {
			fmt.Fprintf(w, "%s: %d\n", endpoint, count)
		}
	})

	fmt.Printf("Listen localhost:%v...\n", port)
	http.ListenAndServe(fmt.Sprintf(":%v", port), nil)

}
