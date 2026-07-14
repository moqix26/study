package main

import "fmt"

func main() {
	const maxExactInteger uint64 = 1 << 53
	id1 := maxExactInteger
	id2 := maxExactInteger + 1

	score1 := float64(id1)
	score2 := float64(id2)

	fmt.Printf("id1=%d score1=%.0f\n", id1, score1)
	fmt.Printf("id2=%d score2=%.0f\n", id2, score2)
	fmt.Printf("scores equal: %v\n", score1 == score2)

	if score1 != score2 {
		panic("expected float64 precision collision above 2^53")
	}

	fmt.Println("Do not store a full 64-bit Snowflake ID as a Redis ZSet score.")
	fmt.Println("Use a safe-range time or logical score, and keep the full ID in the member/cursor.")
}
