package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountWords(t *testing.T) {
	testCases := []struct {
		data     string
		initData map[string]int
		n        int
		expect   []*Pair
	}{
		{
			"hello",
			map[string]int{"hello": 0, "good": 0, "bad": 0},
			3,
			[]*Pair{{"hello", 1}},
		},
		{
			"hello hello! hello? ?good! !good. good.com. This is very good and not bad.",
			map[string]int{"hello": 0, "good": 0, "bad": 0},
			3,
			[]*Pair{{"good", 4}, {"hello", 3}, {"bad", 1}},
		},
		{
			"",
			map[string]int{"hello": 0, "good": 0, "bad": 0},
			3,
			[]*Pair{},
		},
		{
			"aaa bbb aaa bbb bbb ccc ddd ddd eee aaa bbb",
			map[string]int{"aaa": 0, "bbb": 0, "ccc": 0},
			2,
			[]*Pair{{"bbb", 4}, {"aaa", 3}},
		},
		{
			"aaa bbb aaa bbb bbb ccc ddd ddd eee aaa bbb",
			map[string]int{},
			2,
			[]*Pair{},
		},
		{
			"aaa bbb aaa bbb bbb ccc ddd ddd eee aaa bbb",
			map[string]int{"aaa": 0, "bbb": 0, "ccc": 0},
			0,
			[]*Pair{},
		},
		{
			"aaa bbb aaa bbb bbb ddd ddd eee aaa bbb",
			map[string]int{"aaa": 0, "bbb": 0, "ccc": 0, "ddd": 0, "eee": 0, "fff": 0},
			10,
			[]*Pair{{"bbb", 4}, {"aaa", 3}, {"ddd", 2}, {"eee", 1}},
		},
	}
	for _, tc := range testCases {
		counter := NewSafeWordCounter(tc.initData)
		counter.CountWords(tc.data)
		assert.Equal(t, tc.expect, counter.GetTopWords(tc.n), "unexpected result")
	}
}
