package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/time/rate"
)

const (
	contentSelector                  = ".caas-body,.caas-header"
	minWordLen                       = 3
	toManyRequestsLongSleepDuration  = 4 * time.Minute
	toManyRequestsShortSleepDuration = 30 * time.Second
)

type WordCounterService interface {
	FindTopWords(articlesUrl string, bankOfWordsUrl string, n int, batchSize int) ([]*Pair, error)
}

type WordCounterServiceImpl struct {
	httpClient      HttpClient
	artifactFetcher ArtifactFetcher
	workers         int
	progress        int
	limiter         *rate.Limiter
}

func NewWordCounterService(
	httpClient HttpClient,
	artifactFetcher ArtifactFetcher,
	rps int,
	workers int,
	progress int,
) WordCounterService {
	return &WordCounterServiceImpl{
		httpClient:      httpClient,
		artifactFetcher: artifactFetcher,
		workers:         workers,
		limiter:         rate.NewLimiter(rate.Limit(rps), 1),
		progress:        progress,
	}
}

func (service *WordCounterServiceImpl) FindTopWords(
	articlesUrl string,
	bankOfWordsUrl string,
	n int,
	batchSize int,
) ([]*Pair, error) {
	articles, bankOfWords, err := service.artifactFetcher.Fetch(articlesUrl, bankOfWordsUrl)
	if err != nil {
		return nil, fmt.Errorf("error fetching articles: %w", err)
	}
	initCounter := map[string]int{}
	for _, word := range bankOfWords {
		if isValidWord(word) {
			initCounter[strings.ToLower(word)] = 0
		}
	}
	counter := NewSafeWordCounter(initCounter)
	if batchSize != 0 && batchSize < len(articles) {
		articles = articles[:batchSize]
	}
	service.wordCount(articles, counter)
	return counter.GetTopWords(n), nil
}

func (service *WordCounterServiceImpl) wordCount(articles []string, counter SafeWordCounter) {
	startTime := time.Now()
	var workersWg sync.WaitGroup
	var articlesProcessed int64
	articlesChan := make(chan string, len(articles))
	for i := 0; i < service.workers; i++ {
		workersWg.Add(1)
		go service.worker(articlesChan, &workersWg, counter, &articlesProcessed, len(articles), startTime)
	}
	for _, article := range articles {
		articlesChan <- article
	}
	close(articlesChan)
	workersWg.Wait()
}

func (service *WordCounterServiceImpl) worker(
	articlesChan <-chan string,
	wg *sync.WaitGroup,
	counter SafeWordCounter,
	articlesProcessed *int64,
	articlesTotal int,
	startTime time.Time,
) {
	defer wg.Done()
	for articleUrl := range articlesChan {
		articleUrl = articleUrl[:len(articleUrl)-1] + ".html" // avoid redirect
		var err error
		var resp *http.Response
		for attempt := 0; ; attempt += 1 {
			if service.limiter.Wait(context.Background()) != nil {
				log.Printf("failed to aquire limiter token. Sleep for a second")
				time.Sleep(time.Second)
			}
			resp, err = service.httpClient.Get(articleUrl)
			if err == nil && resp.StatusCode == 999 { // 999 - status code for too many requests
				var sleepDuration time.Duration
				if attempt == 0 {
					sleepDuration = toManyRequestsLongSleepDuration
				} else {
					sleepDuration = toManyRequestsShortSleepDuration
				}
				log.Printf("faced too many requests error sleep for %v. attempt %v \n", sleepDuration, attempt+1)
				time.Sleep(sleepDuration)
				continue
			}
			break
		}
		if err != nil || resp.StatusCode != 200 {
			log.Printf("failed to fetch %s: status code %v, error %v\n", articleUrl, resp.StatusCode, err)
			continue
		}
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("failed to parse html %s: error %v\n", articleUrl, err)
			continue
		}
		doc.Find(contentSelector).Each(func(i int, s *goquery.Selection) { counter.CountWords(s.Text()) })
		service.logProgress(articlesProcessed, articlesTotal, startTime)
	}
}

func (service *WordCounterServiceImpl) logProgress(progress *int64, total int, startTime time.Time) {
	if service.progress != 0 {
		atomic.AddInt64(progress, 1)
		s := atomic.LoadInt64(progress)
		if int(s)%service.progress == 0 {
			log.Printf("progress [%v / %v]: %v", s, total, time.Since(startTime))
		}
	}
}

func isValidWord(s string) bool {
	if len(s) < minWordLen {
		return false
	}
	for _, char := range s {
		if !unicode.IsLetter(char) {
			return false
		}
	}
	return true
}
