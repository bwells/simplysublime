package main

import "testing"

var results []string

func BenchmarkMatchAll(b *testing.B) {
	corpus := loadFile("files")
	pattern := "cont test acc"

	for n := 0; n < b.N; n++ {
		results = matchAll(corpus, pattern)
	}
}

func BenchmarkMatchAllN(b *testing.B) {
	corpus := loadFile("files")
	pattern := "cont test acc"

	for n := 0; n < b.N; n++ {
		matchAllN(corpus, pattern)
	}
}
