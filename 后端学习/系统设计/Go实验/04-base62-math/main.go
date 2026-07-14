package main

import (
	"fmt"
	"math"
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func encodeBase62(n uint64) string {
	if n == 0 {
		return "0"
	}
	out := make([]byte, 0, 12)
	for n > 0 {
		out = append(out, alphabet[n%62])
		n /= 62
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return string(out)
}

func decodeBase62(s string) uint64 {
	var n uint64
	for _, c := range []byte(s) {
		idx := -1
		for i := range alphabet {
			if alphabet[i] == c {
				idx = i
				break
			}
		}
		if idx < 0 {
			panic("invalid Base62 character")
		}
		n = n*62 + uint64(idx)
	}
	return n
}

func main() {
	const daily = 10_000_000.0
	const years = 5.0
	const bytesPerRecord = 500.0

	records := daily * 365 * years
	logicalBytes := records * bytesPerRecord

	space7 := math.Pow(62, 7)
	n := 1_000_000_000.0
	lambda := n * (n - 1) / (2 * space7)
	probability := 1 - math.Exp(-lambda)

	fmt.Printf("7-char space: %.0f\n", space7)
	fmt.Printf("five-year records: %.0f\n", records)
	fmt.Printf("logical bytes: %.3f TB (decimal)\n", logicalBytes/1e12)
	fmt.Printf("expected collision pairs at 1B items: %.0f\n", lambda)
	fmt.Printf("probability of at least one collision: %.12f\n", probability)

	value := uint64(9_223_372_036_854_775)
	code := encodeBase62(value)
	decoded := decodeBase62(code)
	fmt.Printf("round trip: %d -> %s -> %d\n", value, code, decoded)

	if math.Abs(logicalBytes/1e12-9.125) > 0.001 {
		panic("storage math is wrong")
	}
	if probability < 0.999999 {
		panic("collision probability should be near one")
	}
	if decoded != value {
		panic("Base62 round trip failed")
	}
}
