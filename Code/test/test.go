package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: wordcount <file>")
		os.Exit(1)
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()

	freq := make(map[string]int)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		for _, w := range strings.Fields(line) {
			w = strings.ToLower(strings.Trim(w, ".,!?\"'"))
			if w == "" {
				continue
			}
			freq[w]++
		}
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for word, count := range freq {
		fmt.Printf("%s: %d\n", word, count)
	}
}
