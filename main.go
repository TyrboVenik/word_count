package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
)

const (
	defaultArticlesUrl          = "https://drive.google.com/uc?id=1TF4RPuj8iFwpa-lyhxG67V8NDlktmTGi&export=download"
	defaultWordBankUrl          = "https://raw.githubusercontent.com/dwyl/english-words/master/words.txt"
	defaultTmpDir               = "tmp"
	defaultArticlesCacheFile    = "articles.txt"
	defaultBankOfWordsCacheFile = "bank_of_words.txt"
	defaultRps                  = 2
	defaultWorkers              = 8
	defaultN                    = 10
	defaultProgress             = 100
)

func parseArgs() *Config {
	var config Config
	flag.IntVar(&config.Rps, "rps", defaultRps, "rps limit")
	flag.IntVar(&config.Workers, "workers", defaultWorkers, "parallel requests limit")
	flag.IntVar(&config.N, "n", defaultN, "amount of top words to be shown")
	flag.IntVar(&config.Progress, "progress", defaultProgress, "show progress every <progress> requests. 0 - no progres")
	flag.IntVar(&config.BatchSize, "batch-size", 0, "process first <batch-size> articles. 0 - process all")
	flag.StringVar(&config.ArticlesUrl, "articles-url", defaultArticlesUrl, "articles file url")
	flag.StringVar(&config.WordBankUrl, "bank-of-words-url", defaultWordBankUrl, "bank of words file url")
	flag.Parse()
	return &config
}

func main() {
	args := parseArgs()
	artifactFetcher := NewArtifactFetcher(
		http.DefaultClient,
		defaultTmpDir,
		defaultArticlesCacheFile,
		defaultBankOfWordsCacheFile,
	)
	wordCounterService := NewWordCounterService(
		http.DefaultClient,
		artifactFetcher,
		args.Rps,
		args.Workers,
		args.Progress,
	)
	top, err := wordCounterService.FindTopWords(args.ArticlesUrl, args.WordBankUrl, args.N, args.BatchSize)
	if err != nil {
		log.Fatal(fmt.Errorf("error processing articles: %w", err))
	}
	prettyJSON, err := json.MarshalIndent(Response{Data: top}, "", "    ")
	if err != nil {
		log.Fatal(fmt.Errorf("error serializing json: %w", err))
	}
	fmt.Println(string(prettyJSON))
}
