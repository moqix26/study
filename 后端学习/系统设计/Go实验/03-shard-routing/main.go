package main

import "fmt"

const (
	dbCount    = 4
	tableCount = 16
	slotCount  = dbCount * tableCount
)

func CorrectRoute(key uint64) (slot, db, table int) {
	slot = int(key % slotCount)
	db = slot / tableCount
	table = slot % tableCount
	return
}

func CorrelatedRoute(key uint64) (db, table int) {
	return int(key % dbCount), int(key % tableCount)
}

func main() {
	correctPairs := map[[2]int]struct{}{}
	oldPairs := map[[2]int]struct{}{}

	for key := uint64(0); key < slotCount; key++ {
		_, db, table := CorrectRoute(key)
		correctPairs[[2]int{db, table}] = struct{}{}

		oldDB, oldTable := CorrelatedRoute(key)
		oldPairs[[2]int{oldDB, oldTable}] = struct{}{}
	}

	fmt.Printf("correct route uses %d physical pairs\n", len(correctPairs))
	fmt.Printf("correlated mod route uses %d physical pairs\n", len(oldPairs))

	if len(correctPairs) != 64 {
		panic("correct route did not cover 64 pairs")
	}
	if len(oldPairs) != 16 {
		panic("expected old route to expose the 16-pair bug")
	}
}
