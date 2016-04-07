package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
)

func fuzzy(pattern string, str string) (score int, matched_indices []int) {
	const LastWasSeparatorBonus = 5
	const LastWasPathSeparatorBonus = 8
	const MatchedCharBonus = 2
	const SeparatorMatchBonus = 1
	const GapPenelty = 2

	score = 0
	matched_indices = make([]int, 0, 100)
	last_str_was_sep := true
	// sep_matches := 0
	// last_seps := 0
	// path_last_seps := 0

	pattern_idx := 0

	// SEPARATORS = '_', ' ', '/'
	// PATH SEPARATORS = '/', '\'

	str_len := len([]rune(str))
	pattern_len := len([]rune(pattern))

	var last_str_char string = ""

	for str_idx := 0; str_idx < str_len; str_idx++ {
		var pattern_char string = ""
		if pattern_idx < pattern_len {
			pattern_char = string(pattern[pattern_idx])
		}
		str_char := string(str[str_idx])

		str_is_sep := false
		if isSep(str_char) {
			str_is_sep = true
		}

		if pattern_char != "" && strings.ToLower(pattern_char) == strings.ToLower(str_char) {
			matched_indices = append(matched_indices, str_idx)
			score += MatchedCharBonus
			pattern_idx += 1

			if last_str_was_sep {
				if isPathSep(last_str_char) {
					score += LastWasPathSeparatorBonus
					// path_last_seps += 1
				} else {
					score += LastWasSeparatorBonus
					// last_seps += 1
				}
			}
		} else if str_is_sep && isSep(pattern_char) {
			score += SeparatorMatchBonus
			// sep_matches += 1
			pattern_idx += 1
		}

		last_str_was_sep = str_is_sep
		last_str_char = str_char
	}

	run_length := 0
	gaps := 0
	last_idx := -1

	for _, idx := range matched_indices {
		if last_idx < 0 {
			last_idx = idx
			continue
		}

		if idx-last_idx == 1 {
			run_length += 1
		} else {
			score += (run_length + 1) * (run_length + 1)
			run_length = 0
			gaps += 1
		}
	}

	if run_length > 1 {
		score += (run_length + 1) * (run_length + 1)
	}

	score -= gaps * GapPenelty

	return score, matched_indices
}

func isSep(char string) bool {
	if char == "/" || char == " " || char == "_" {
		return true
	} else {
		return false
	}
}

func isPathSep(char string) bool {
	if char == "/" || char == "\\" {
		return true
	} else {
		return false
	}
}

func formatMatch(indices []int, str string) string {

	s := "\033[31;1m"
	e := "\033[0m"

	hunks := make([]string, len(indices)+1)

	last_idx := 0
	i := 0

	for _, str_idx := range indices {
		hunks[i] = str[last_idx:str_idx] + s + string(str[str_idx]) + e
		last_idx = str_idx + 1
		i++
	}
	hunks[i] = str[last_idx:len(str)]

	formatted_str := strings.Join(hunks, "")

	return formatted_str
}

type matchResult struct {
	score   int
	indices []int
	str     string
}

// An matchResultHeap is a min-heap of matchResults
type matchResultHeap []matchResult

func (h matchResultHeap) Len() int           { return len(h) }
func (h matchResultHeap) Less(i, j int) bool { return h[i].score < h[j].score }
func (h matchResultHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *matchResultHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(matchResult))
}

func (h *matchResultHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func matchAll(items []string, pattern string) []string {
	h := &matchResultHeap{}
	heap.Init(h)
	maxScore := -99999999999
	for _, item := range items {
		score, indices := fuzzy(pattern, item)
		r := matchResult{score: score * -1, indices: indices, str: item}
		heap.Push(h, r)

		if r.score > maxScore {
			maxScore = r.score
		}
	}
	// fmt.Println("max score %d", maxScore)

	results := make([]string, 10)
	for i := 0; i < 10; i++ {
		r := heap.Pop(h).(matchResult)
		// fmt.Println("%+v", r)
		results[i] = formatMatch(r.indices, r.str)
	}

	return results
}

func matchAllN(items []string, pattern string) []string {

	ch := make(chan matchResult, 100)
	done := make(chan bool, 1)

	h := &matchResultHeap{}
	heap.Init(h)

	workers := runtime.NumCPU()

	var wg sync.WaitGroup
	wg.Add(workers)

	var final sync.WaitGroup
	final.Add(1)

	producer := func(items []string, c chan<- matchResult) {
		defer wg.Done()
		for _, item := range items {
			score, indices := fuzzy(pattern, item)
			r := matchResult{score: score * -1, indices: indices, str: item}
			c <- r
		}
	}

	chunk_size := len(items) / workers
	for i := 0; i < workers; i++ {
		start := i * chunk_size
		end := (i + 1) * chunk_size
		if end > len(items) {
			end = len(items)
		}
		go producer(items[start:end], ch)
	}

	results := make([]string, 10)

	drain := func() {
		for {
			select {
			case r := <-ch:
				heap.Push(h, r)
			default:
				return
			}
		}
	}

	go func() {
		for {
			select {
			case r := <-ch:
				heap.Push(h, r)
			case <-done:

				// drain the rest of the content of the channel
				drain()

				for i := 0; i < 10; i++ {
					r := heap.Pop(h).(matchResult)
					results[i] = formatMatch(r.indices, r.str)
				}

				final.Done()
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		done <- true
		close(done)
	}()

	final.Wait()

	return results
}

func loadFile(filename string) []string {

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	lines := make([]string, 0, 4000)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return lines
}

func main() {
	pattern := "cont test acc"
	// input := "/erp/controllers/testing/accounts.py"
	// score, indices := fuzzy("cont test acc", input)
	//
	// formatted := formatMatch(indices, input)
	// fmt.Println(score, formatted)

	corpus := loadFile("files")
	// corpus := []string{
	// 	"erp/automated_tests/tests/controllers/test_accounts.py",
	// 	"erp/controllers/testing/accounttests.py",
	// 	"erp/controllers/testing/accounts.py",
	// 	"erp/automated_tests/tests/controllers/test_app.py",
	// 	"erp/controllers/testing/tracktests.py",
	// 	"erp/controllers/testing/tracks.py",
	// 	"erp/controllers/testing/testpackages.py",
	// 	"erp/controllers/testing/paperinvoicerecipients.py",
	// 	"erp/controllers/testing/emailinvoicerecipients.py",
	// 	"erp/automated_tests/tests/controllers/test_transactions.py",
	// }
	results := matchAllN(corpus, pattern)
	// results := matchAll(corpus, pattern)
	for _, result := range results {
		fmt.Println(result)
	}

}
