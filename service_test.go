package main

import (
	"errors"
	"os"
	"strings"
	"testing"

	"io"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockArticlesHttpClient struct {
	mock.Mock
}

func (m *MockArticlesHttpClient) Get(url string) (*http.Response, error) {
	args := m.Called(url)
	return args.Get(0).(*http.Response), args.Error(1)
}

type MockArtifactFetcher struct {
	mock.Mock
}

func (fetcher *MockArtifactFetcher) Fetch(articlesUrl string, bankOfWordsUrl string) ([]string, []string, error) {
	args := fetcher.Called(articlesUrl, bankOfWordsUrl)
	return args.Get(0).([]string), args.Get(1).([]string), args.Error(2)
}

func TestIsValidWord(t *testing.T) {
	testCases := []struct {
		word   string
		expect bool
	}{
		{"hello", true},
		{"Hello", true},
		{"HELLO", true},
		{"Hi", false},
		{"Hi1", false},
		{"hello1", false},
		{"hello there", false},
		{"a-b-c", false},
		{"1-1-1", false},
		{"A.A.A", false},
		{"its's", false},
	}
	for _, tc := range testCases {
		result := isValidWord(tc.word)
		assert.Equal(t, tc.expect, result, "unexpected result")
	}
}

func TestWordCounterService(t *testing.T) {
	htmlExample, err := os.ReadFile("testdata/example.html")
	if err != nil {
		assert.NoError(t, err)
		return
	}

	articles := []string{"https://a1/", "https://a2/", "https://a3/", "https://a4/", "https://a5/"}
	bankOfWords := []string{"a", "security", "the", "bigbigwordverybig"}
	mockArtifactFetcher := new(MockArtifactFetcher)
	mockArtifactFetcher.On("Fetch", "articles", "bank_of_words").Return(articles, bankOfWords, nil)

	mockClient := new(MockArticlesHttpClient)
	mockClient.On("Get", "https://a1.html").Return(
		&http.Response{
			StatusCode: http.StatusOK,
			// article body. bigbigwordverybig: 1
			Body: io.NopCloser(strings.NewReader(
				"<div class=\"caas-body\"><p>hello good there bigbigwordverybig</p></div>",
			)),
		},
		nil,
	)
	mockClient.On("Get", "https://a2.html").Return(
		&http.Response{
			StatusCode: http.StatusOK,
			// article header. the: 1
			Body: io.NopCloser(strings.NewReader(
				"<div class=\"caas-header\"><p>hello good there the</p></div>",
			)),
		},
		nil,
	)
	mockClient.On("Get", "https://a3.html").Return(
		&http.Response{
			StatusCode: http.StatusOK,
			// no valid html classes
			Body: io.NopCloser(strings.NewReader("good? no! good? no!!!")),
		},
		nil,
	)
	mockClient.On("Get", "https://a4.html").Return(
		&http.Response{
			StatusCode: http.StatusOK,
			// article example. security: 4, the: 18
			Body: io.NopCloser(strings.NewReader(string(htmlExample))),
		},
		nil,
	)
	mockClient.On("Get", "https://a5.html").Return(
		&http.Response{
			StatusCode: http.StatusNotFound,
			// not 200. skipped
			Body: io.NopCloser(strings.NewReader(
				"<div class=\"caas-header\"><p>hello good there the</p></div>",
			)),
		},
		nil,
	)
	mockClient.On("Get", "https://a6.html").Return(
		nil,
		// error. skipped
		errors.New("something went wrong"),
	)

	service := NewWordCounterService(mockClient, mockArtifactFetcher, 2, 2, 0)
	top, err := service.FindTopWords("articles", "bank_of_words", 10, 0)

	assert.NoError(t, err)
	expected := []*Pair{{"the", 19}, {"security", 4}, {"bigbigwordverybig", 1}}
	assert.Equal(t, expected, top, "unexpected result")
}
