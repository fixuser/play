package guandan

import (
	"testing"
)

func TestSearchSequence_A2345(t *testing.T) {
	// Trump is 6, so 2 is not wild (unless 2 is level, but here we assume level is 6)
	// Wait, Rank.Weight(trump) checks if r == trump.
	// If trump is Rank6, then Rank6 is level. Rank2 is just Rank2 (1).
	trump := Rank6

	cardsMap := make(map[Rank][]Card)
	cardsMap[RankA] = []Card{NewCard(RankA, SuitSpader)}
	cardsMap[Rank2] = []Card{NewCard(Rank2, SuitSpader)}
	cardsMap[Rank3] = []Card{NewCard(Rank3, SuitSpader)}
	cardsMap[Rank4] = []Card{NewCard(Rank4, SuitSpader)}
	cardsMap[Rank5] = []Card{NewCard(Rank5, SuitSpader)}

	wildCards := []Card{}

	// searchSequence(cardsMap, wildCards, length, width, minMainPoint, trump)
	// length=5, width=1
	res := searchSequence(cardsMap, wildCards, 5, 1, 0, trump)

	if res == nil {
		t.Errorf("Expected A, 2, 3, 4, 5 to be found")
	} else {
		if len(res) != 5 {
			t.Errorf("Expected 5 cards, got %d", len(res))
		}
		// Check ranks
		// The order in result depends on implementation.
		// searchSequence appends in order of sequence.
		// For start=0: A, 2, 3, 4, 5.
		expectedRanks := []Rank{RankA, Rank2, Rank3, Rank4, Rank5}
		for i, c := range res {
			if c.Rank != expectedRanks[i] {
				t.Errorf("Index %d: expected rank %d, got %d", i, expectedRanks[i], c.Rank)
			}
		}
	}
}

func TestSearchSequence_10JQKA(t *testing.T) {
	trump := Rank6

	cardsMap := make(map[Rank][]Card)
	cardsMap[Rank10] = []Card{NewCard(Rank10, SuitSpader)}
	cardsMap[RankJ] = []Card{NewCard(RankJ, SuitSpader)}
	cardsMap[RankQ] = []Card{NewCard(RankQ, SuitSpader)}
	cardsMap[RankK] = []Card{NewCard(RankK, SuitSpader)}
	cardsMap[RankA] = []Card{NewCard(RankA, SuitSpader)}

	wildCards := []Card{}

	res := searchSequence(cardsMap, wildCards, 5, 1, 0, trump)

	if res == nil {
		t.Errorf("Expected 10, J, Q, K, A to be found")
	} else {
		if len(res) != 5 {
			t.Errorf("Expected 5 cards, got %d", len(res))
		}
		expectedRanks := []Rank{Rank10, RankJ, RankQ, RankK, RankA}
		for i, c := range res {
			if c.Rank != expectedRanks[i] {
				t.Errorf("Index %d: expected rank %d, got %d", i, expectedRanks[i], c.Rank)
			}
		}
	}
}

func TestSearchSequenceAll_Both(t *testing.T) {
	trump := Rank6
	cardsMap := make(map[Rank][]Card)
	// Add A, 2, 3, 4, 5
	cardsMap[Rank2] = append(cardsMap[Rank2], NewCard(Rank2, SuitSpader))
	cardsMap[Rank3] = append(cardsMap[Rank3], NewCard(Rank3, SuitSpader))
	cardsMap[Rank4] = append(cardsMap[Rank4], NewCard(Rank4, SuitSpader))
	cardsMap[Rank5] = append(cardsMap[Rank5], NewCard(Rank5, SuitSpader))

	// Add 10, J, Q, K
	cardsMap[Rank10] = append(cardsMap[Rank10], NewCard(Rank10, SuitHeart))
	cardsMap[RankJ] = append(cardsMap[RankJ], NewCard(RankJ, SuitHeart))
	cardsMap[RankQ] = append(cardsMap[RankQ], NewCard(RankQ, SuitHeart))
	cardsMap[RankK] = append(cardsMap[RankK], NewCard(RankK, SuitHeart))

	// Add 2 Aces
	cardsMap[RankA] = append(cardsMap[RankA], NewCard(RankA, SuitSpader))
	cardsMap[RankA] = append(cardsMap[RankA], NewCard(RankA, SuitHeart))

	wildCards := []Card{}

	results := searchSequenceAll(cardsMap, wildCards, 5, 1, 0, trump)

	// Should find both sequences
	foundLow := false
	foundHigh := false

	for _, res := range results {
		if len(res) != 5 {
			continue
		}
		// Check if low
		if res[0].Rank == RankA && res[1].Rank == Rank2 {
			foundLow = true
		}
		// Check if high
		if res[0].Rank == Rank10 && res[4].Rank == RankA {
			foundHigh = true
		}
	}

	if !foundLow {
		t.Errorf("Did not find A, 2, 3, 4, 5")
	}
	if !foundHigh {
		t.Errorf("Did not find 10, J, Q, K, A")
	}
}
