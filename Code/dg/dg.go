package main

import "fmt"

func f(n int) {
	if n == 0 {
		return
	}
	defer fmt.Print(n % 10)
	f(n / 10)
}

func main() {
	f(124761)
	is := []int{1, 2, 3, 4}
	fs := []func(int){f, f, f, f}
	for i := range is {
		fs[i](is[i])
	}
}
