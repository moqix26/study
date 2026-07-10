package main

import "fmt"

func Contains[T comparable](items []T, target T) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func main() {
	fmt.Println(Contains([]int{1, 2, 3}, 2))
	fmt.Println(Contains([]string{"go", "java"}, "go"))
}
