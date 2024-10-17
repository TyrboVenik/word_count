package main

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockArtifactsHttpClient struct {
	mock.Mock
}

func (m *MockArtifactsHttpClient) Get(url string) (*http.Response, error) {
	args := m.Called(url)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestArtifactFetcher(t *testing.T) {
	cacheDir := "tmp_test"
	articlesUrl, bankOfWordsUrl := "https://articles", "https://bankOfWords"
	defer os.RemoveAll(cacheDir)

	mockHttpClient := new(MockArtifactsHttpClient)
	mockHttpClient.On("Get", articlesUrl).Return(
		&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("a1\na2\na3\n")),
		},
		nil,
	).Once()
	mockHttpClient.On("Get", bankOfWordsUrl).Return(
		&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("w1\nw2\nw3\n")),
		},
		nil,
	).Once()
	expectedArticles := []string{"a1", "a2", "a3"}
	expectedBankOfWords := []string{"w1", "w2", "w3"}

	fetcher := NewArtifactFetcher(mockHttpClient, cacheDir, "articles.txt", "bank_of_words.txt")
	articles, bankOfWords, err := fetcher.Fetch(articlesUrl, bankOfWordsUrl)

	// no cache
	assert.NoError(t, err)
	assert.Equal(t, expectedArticles, articles, "unexpected result")
	assert.Equal(t, expectedBankOfWords, bankOfWords, "unexpected result")

	// cache
	articles, bankOfWords, err = fetcher.Fetch(articlesUrl, bankOfWordsUrl)
	assert.NoError(t, err)
	assert.Equal(t, expectedArticles, articles, "unexpected result")
	assert.Equal(t, expectedBankOfWords, bankOfWords, "unexpected result")
}
