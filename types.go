package main

import (
	"regexp"
	"sort"
	"strings"
	"sync"
)

var wordRegexp = regexp.MustCompile(`[a-zA-Z]+`)

type Pair struct {
	Word   string `json:"word"`
	Amount int    `json:"amount"`
}
type Response struct {
	Data []*Pair `json:"data"`
}

type Config struct {
	Rps         int
	Workers     int
	N           int
	Progress    int
	BatchSize   int
	ArticlesUrl string
	WordBankUrl string
}

type SafeWordCounter interface {
	CountWords(text string)
	GetTopWords(n int) []*Pair
}

type SafeWordCounterImpl struct {
	mu   sync.Mutex
	data map[string]int
}

func NewSafeWordCounter(initData map[string]int) SafeWordCounter {
	return &SafeWordCounterImpl{data: initData}
}

func (counter *SafeWordCounterImpl) CountWords(text string) {
	for _, word := range wordRegexp.FindAllString(text, -1) {
		counter.AddWord(strings.ToLower(word))
	}
}

func (counter *SafeWordCounterImpl) AddWord(key string) {
	counter.mu.Lock()
	defer counter.mu.Unlock()
	if _, exists := counter.data[key]; !exists {
		return
	}
	counter.data[key]++
}

func (counter *SafeWordCounterImpl) GetTopWords(n int) []*Pair {
	counter.mu.Lock()
	defer counter.mu.Unlock()
	var pairs []*Pair
	for key, value := range counter.data {
		pairs = append(pairs, &Pair{Word: key, Amount: value})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Amount > pairs[j].Amount
	})
	if len(pairs) == 0 || pairs[0].Amount == 0 {
		return []*Pair{}
	}
	if len(pairs) < n {
		n = len(pairs)
	}
	lastNotEmpty := n
	for lastNotEmpty > 0 && pairs[lastNotEmpty-1].Amount == 0 {
		lastNotEmpty -= 1
	}
	return pairs[:lastNotEmpty]
}
