package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type ArtifactFetcher interface {
	Fetch(articlesUrl string, bankOfWordsUrl string) ([]string, []string, error)
}

type ArtifactFetcherImpl struct {
	httpClient           HttpClient
	cacheDir             string
	articlesCacheFile    string
	bankOfWordsCacheFile string
}

func NewArtifactFetcher(
	httpClient HttpClient,
	cacheDir string,
	articlesCacheFile string,
	bankOfWordsCacheFile string,
) ArtifactFetcher {
	return &ArtifactFetcherImpl{
		httpClient:           httpClient,
		cacheDir:             cacheDir,
		articlesCacheFile:    articlesCacheFile,
		bankOfWordsCacheFile: bankOfWordsCacheFile,
	}
}

func (fetcher *ArtifactFetcherImpl) Fetch(articlesUrl string, bankOfWordsUrl string) ([]string, []string, error) {
	articles, err := fetcher.getData(articlesUrl, fetcher.articlesCacheFile)
	if err != nil {
		return nil, nil, err
	}
	bankOfWords, err := fetcher.getData(bankOfWordsUrl, fetcher.bankOfWordsCacheFile)
	if err != nil {
		return nil, nil, err
	}
	return articles, bankOfWords, nil
}

func (fetcher *ArtifactFetcherImpl) getData(downloadURL string, cacheFilePath string) ([]string, error) {
	fullCacheFilePath := fmt.Sprintf("%s/%s", fetcher.cacheDir, cacheFilePath)
	if _, err := os.Stat(fullCacheFilePath); os.IsNotExist(err) {
		if err := os.MkdirAll(fetcher.cacheDir, 0755); err != nil {
			return nil, fmt.Errorf("error creating cache directory: %w", err)
		}
		err := fetcher.fetchData(downloadURL, fullCacheFilePath)
		if err != nil {
			return nil, fmt.Errorf("error fetching data from %s: %w", downloadURL, err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("error checking file %v: %w", fullCacheFilePath, err)
	}

	file, err := os.Open(fullCacheFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file %v: %w", fullCacheFilePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var data []string
	for scanner.Scan() {
		data = append(data, strings.Trim(scanner.Text(), " \n\t"))
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %v: %w", fullCacheFilePath, err)
	}
	return data, nil
}

func (fetcher *ArtifactFetcherImpl) fetchData(downloadURL string, cacheFilePath string) error {
	outFile, err := os.Create(cacheFilePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer outFile.Close()

	resp, err := fetcher.httpClient.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("error making HTTP request %v: %w", downloadURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 response status %v: %s", downloadURL, resp.Status)

	}
	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}
	return nil
}
