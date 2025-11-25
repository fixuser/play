package guandan

import (
	"testing"
)

// TestSearchBomb 测试 searchBomb 函数
func TestSearchBomb(t *testing.T) {
	trump := Rank6
	tests := []struct {
		name          string
		cardsMap      map[Rank][]Card
		wildCards     []Card
		length        int
		minMainPoint  uint8
		expectedFound bool
		expectedRank  Rank
		description   string
	}{
		{
			"找4张炸弹-有4张5",
			map[Rank][]Card{
				Rank5: {
					NewCard(Rank5, SuitSpader),
					NewCard(Rank5, SuitHeart),
					NewCard(Rank5, SuitDiamond),
					NewCard(Rank5, SuitClub),
				},
			},
			[]Card{},
			4,
			0,
			true,
			Rank5,
			"有足够的牌",
		},
		{
			"找4张炸弹-3张A加1万能牌",
			map[Rank][]Card{
				RankA: {
					NewCard(RankA, SuitSpader),
					NewCard(RankA, SuitHeart),
					NewCard(RankA, SuitDiamond),
				},
			},
			[]Card{NewCard(Rank6, SuitHeart)},
			4,
			0,
			true,
			RankA,
			"用万能牌补齐",
		},
		{
			"找5张炸弹-3张K加2万能牌",
			map[Rank][]Card{
				RankK: {
					NewCard(RankK, SuitSpader),
					NewCard(RankK, SuitHeart),
					NewCard(RankK, SuitDiamond),
				},
			},
			[]Card{
				NewCard(Rank6, SuitHeart),
				NewCard(Rank6, SuitHeart),
			},
			5,
			0,
			true,
			RankK,
			"用2张万能牌补齐",
		},
		{
			"找对子-大于5",
			map[Rank][]Card{
				Rank3: {NewCard(Rank3, SuitSpader), NewCard(Rank3, SuitHeart)},
				Rank7: {NewCard(Rank7, SuitSpader), NewCard(Rank7, SuitHeart)},
			},
			[]Card{},
			2,
			uint8(Rank5),
			true,
			Rank7,
			"找到7对",
		},
		{
			"找不到炸弹-牌不够",
			map[Rank][]Card{
				Rank5: {
					NewCard(Rank5, SuitSpader),
					NewCard(Rank5, SuitHeart),
				},
			},
			[]Card{},
			4,
			0,
			false,
			RankNone,
			"只有2张5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := searchBomb(tt.cardsMap, tt.wildCards, tt.length, tt.minMainPoint, trump)
			if tt.expectedFound {
				if result == nil {
					t.Errorf("%s: 期望找到炸弹，但返回nil", tt.description)
					return
				}
				if len(result) != tt.length {
					t.Errorf("%s: 期望长度 %d，实际 %d", tt.description, tt.length, len(result))
				}
				// 检查主要点数
				foundExpectedRank := false
				for _, c := range result {
					if c.Rank == tt.expectedRank || c.IsWild(trump) {
						foundExpectedRank = true
						break
					}
				}
				if !foundExpectedRank {
					t.Errorf("%s: 未找到预期的点数 %v", tt.description, tt.expectedRank)
				}
			} else {
				if result != nil {
					t.Errorf("%s: 期望nil，但找到了牌", tt.description)
				}
			}
		})
	}
}

// TestSearchBombAll 测试 searchBombAll 函数
func TestSearchBombAll(t *testing.T) {
	trump := Rank6
	cardsMap := map[Rank][]Card{
		Rank3: {
			NewCard(Rank3, SuitSpader),
			NewCard(Rank3, SuitHeart),
			NewCard(Rank3, SuitDiamond),
			NewCard(Rank3, SuitClub),
		},
		Rank7: {
			NewCard(Rank7, SuitSpader),
			NewCard(Rank7, SuitHeart),
			NewCard(Rank7, SuitDiamond),
			NewCard(Rank7, SuitClub),
		},
		RankK: {
			NewCard(RankK, SuitSpader),
			NewCard(RankK, SuitHeart),
		},
	}
	wildCards := []Card{NewCard(Rank6, SuitHeart), NewCard(Rank6, SuitHeart)}

	// 找所有4张炸弹
	results := searchBombAll(cardsMap, wildCards, 4, 0, trump)

	// 应该找到 3、7、K（用万能牌）三种
	if len(results) != 3 {
		t.Errorf("期望找到3种4张炸弹，实际找到 %d 种", len(results))
	}
}

// TestSearchFullHouse 测试 searchFullHouse 函数
func TestSearchFullHouse(t *testing.T) {
	trump := Rank6
	tests := []struct {
		name          string
		cardsMap      map[Rank][]Card
		wildCards     []Card
		minMainPoint  uint8
		minSubPoint   uint8
		expectedFound bool
		expectedTrips Rank
		expectedPair  Rank
		description   string
	}{
		{
			"三张5带两张3",
			map[Rank][]Card{
				Rank5: {
					NewCard(Rank5, SuitSpader),
					NewCard(Rank5, SuitHeart),
					NewCard(Rank5, SuitDiamond),
				},
				Rank3: {
					NewCard(Rank3, SuitSpader),
					NewCard(Rank3, SuitHeart),
				},
			},
			[]Card{},
			0,
			0,
			true,
			Rank5,
			Rank3,
			"标准三带二",
		},
		{
			"用万能牌凑三带二",
			map[Rank][]Card{
				RankA: {
					NewCard(RankA, SuitSpader),
					NewCard(RankA, SuitHeart),
				},
				Rank7: {
					NewCard(Rank7, SuitSpader),
					NewCard(Rank7, SuitHeart),
				},
			},
			[]Card{NewCard(Rank6, SuitHeart)},
			0,
			0,
			true,
			RankA,
			Rank7,
			"用1张万能牌补A",
		},
		{
			"找大于指定点数的三带二",
			map[Rank][]Card{
				Rank3: {
					NewCard(Rank3, SuitSpader),
					NewCard(Rank3, SuitHeart),
					NewCard(Rank3, SuitDiamond),
				},
				Rank2: {
					NewCard(Rank2, SuitSpader),
					NewCard(Rank2, SuitHeart),
				},
				RankK: {
					NewCard(RankK, SuitSpader),
					NewCard(RankK, SuitHeart),
					NewCard(RankK, SuitDiamond),
				},
				RankQ: {
					NewCard(RankQ, SuitSpader),
					NewCard(RankQ, SuitHeart),
				},
			},
			[]Card{},
			uint8(Rank5),
			0,
			true,
			RankK,
			RankQ,
			"找到K带Q",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := searchFullHouse(tt.cardsMap, tt.wildCards, tt.minMainPoint, tt.minSubPoint, trump)
			if tt.expectedFound {
				if result == nil {
					t.Errorf("%s: 期望找到三带二，但返回nil", tt.description)
					return
				}
				if len(result) != 5 {
					t.Errorf("%s: 期望5张牌，实际 %d", tt.description, len(result))
				}
			} else {
				if result != nil {
					t.Errorf("%s: 期望nil，但找到了牌", tt.description)
				}
			}
		})
	}
}

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

// TestCardsSearch_Single 测试 Search 方法 - 单张
func TestCardsSearch_Single(t *testing.T) {
	trump := Rank6
	handCards := Cards{
		NewCard(Rank3, SuitSpader),
		NewCard(Rank5, SuitHeart),
		NewCard(Rank7, SuitDiamond),
		NewCard(RankK, SuitClub),
		NewCard(RankA, SuitSpader),
	}

	// 对方出单张5，我方应该能找到7
	target := NewPattern(Cards{NewCard(Rank5, SuitSpader)}, trump)
	result := handCards.Search(target, trump)

	if result == nil {
		t.Error("期望找到能压制的单张")
		return
	}
	if len(result) != 1 {
		t.Errorf("期望1张牌，实际 %d", len(result))
	}
	if result[0].Rank != Rank7 {
		t.Errorf("期望找到7（最小的压制牌），实际找到 %v", result[0].Rank)
	}
}

// TestCardsSearch_Pair 测试 Search 方法 - 对子
func TestCardsSearch_Pair(t *testing.T) {
	trump := Rank6
	handCards := Cards{
		NewCard(Rank3, SuitSpader),
		NewCard(Rank3, SuitHeart),
		NewCard(Rank7, SuitDiamond),
		NewCard(Rank7, SuitClub),
		NewCard(RankK, SuitSpader),
		NewCard(RankK, SuitHeart),
	}

	// 对方出对5
	target := NewPattern(Cards{
		NewCard(Rank5, SuitSpader),
		NewCard(Rank5, SuitHeart),
	}, trump)
	result := handCards.Search(target, trump)

	if result == nil {
		t.Error("期望找到能压制的对子")
		return
	}
	if len(result) != 2 {
		t.Errorf("期望2张牌，实际 %d", len(result))
	}
	// 应该找到对7（最小的）
	if result[0].Rank != Rank7 {
		t.Errorf("期望找到对7，实际找到 %v", result[0].Rank)
	}
}

// TestCardsSearch_Bomb 测试 Search 方法 - 炸弹
func TestCardsSearch_Bomb(t *testing.T) {
	trump := Rank6
	handCards := Cards{
		NewCard(Rank5, SuitSpader),
		NewCard(Rank5, SuitHeart),
		NewCard(Rank5, SuitDiamond),
		NewCard(Rank5, SuitClub),
		NewCard(Rank7, SuitSpader),
		NewCard(Rank7, SuitHeart),
		NewCard(Rank7, SuitDiamond),
		NewCard(Rank7, SuitClub),
	}

	// 对方出4张炸弹3
	target := NewPattern(Cards{
		NewCard(Rank3, SuitSpader),
		NewCard(Rank3, SuitHeart),
		NewCard(Rank3, SuitDiamond),
		NewCard(Rank3, SuitClub),
	}, trump)
	result := handCards.Search(target, trump)

	if result == nil {
		t.Error("期望找到能压制的炸弹")
		return
	}
	if len(result) != 4 {
		t.Errorf("期望4张牌，实际 %d", len(result))
	}
	// 应该找到5的炸弹（点数更大但最小的）
	if result[0].Rank != Rank5 {
		t.Errorf("期望找到5炸弹，实际找到 %v", result[0].Rank)
	}
}

// TestCardsSearch_BombWithWildCard 测试用万能牌炸弹
func TestCardsSearch_BombWithWildCard(t *testing.T) {
	trump := Rank6
	handCards := Cards{
		NewCard(RankA, SuitSpader),
		NewCard(RankA, SuitHeart),
		NewCard(RankA, SuitDiamond),
		NewCard(Rank6, SuitHeart), // 万能牌
		NewCard(Rank7, SuitSpader),
	}

	// 对方出单张5
	target := NewPattern(Cards{NewCard(Rank5, SuitSpader)}, trump)
	result := handCards.Search(target, trump)

	// 应该找到单张7，而不是用炸弹
	if result == nil {
		t.Error("期望找到能压制的牌")
		return
	}
	if len(result) != 1 {
		t.Errorf("期望1张牌（单张），实际 %d", len(result))
	}
}

// TestCardsSearch_Straight 测试 Search 方法 - 顺子
func TestCardsSearch_Straight(t *testing.T) {
	trump := Rank2
	handCards := Cards{
		NewCard(Rank4, SuitSpader),
		NewCard(Rank5, SuitHeart),
		NewCard(Rank6, SuitDiamond),
		NewCard(Rank7, SuitClub),
		NewCard(Rank8, SuitSpader),
		NewCard(Rank9, SuitHeart),
	}

	// 对方出顺子3,4,5,6,7
	target := NewPattern(Cards{
		NewCard(Rank3, SuitSpader),
		NewCard(Rank4, SuitHeart),
		NewCard(Rank5, SuitDiamond),
		NewCard(Rank6, SuitClub),
		NewCard(Rank7, SuitSpader),
	}, trump)
	result := handCards.Search(target, trump)

	if result == nil {
		t.Error("期望找到能压制的顺子")
		return
	}
	if len(result) != 5 {
		t.Errorf("期望5张牌，实际 %d", len(result))
	}
	// 应该找到4,5,6,7,8
	if result[0].Rank != Rank4 || result[4].Rank != Rank8 {
		t.Errorf("期望找到4-8的顺子")
	}
}

// TestCardsSearch_FourJokers 测试四大天王
func TestCardsSearch_FourJokers(t *testing.T) {
	trump := Rank6
	handCards := Cards{
		NewCard(RankA, SuitSpader),
		NewCard(RankA, SuitHeart),
		NewCard(RankA, SuitDiamond),
		NewCard(RankA, SuitClub),
		NewCard(RankA, SuitSpader),
	}

	// 对方出四大天王
	target := NewPattern(Cards{
		NewCard(RankJokerSmall, SuitJoker),
		NewCard(RankJokerSmall, SuitJoker),
		NewCard(RankJokerBig, SuitJoker),
		NewCard(RankJokerBig, SuitJoker),
	}, trump)
	result := handCards.Search(target, trump)

	// 四大天王无法压制
	if result != nil {
		t.Error("四大天王不应该被任何牌压制")
	}
}

// TestCardsSearch_UseBombToBeaSingle 测试用炸弹压单张
func TestCardsSearch_UseBombToBeaSingle(t *testing.T) {
	trump := Rank6
	handCards := Cards{
		NewCard(Rank5, SuitSpader),
		NewCard(Rank5, SuitHeart),
		NewCard(Rank5, SuitDiamond),
		NewCard(Rank5, SuitClub),
	}

	// 对方出单张A
	target := NewPattern(Cards{NewCard(RankA, SuitSpader)}, trump)
	result := handCards.Search(target, trump)

	// 应该用炸弹压制
	if result == nil {
		t.Error("期望用炸弹压制")
		return
	}
	if len(result) != 4 {
		t.Errorf("期望4张炸弹，实际 %d", len(result))
	}
}

// TestCardsSearchAll_Multiple 测试 SearchAll 找到多个压制方案
func TestCardsSearchAll_Multiple(t *testing.T) {
	trump := Rank6
	handCards := Cards{
		NewCard(Rank7, SuitSpader),
		NewCard(Rank8, SuitHeart),
		NewCard(Rank9, SuitDiamond),
		NewCard(Rank10, SuitClub),
		NewCard(RankJ, SuitSpader),
		NewCard(RankK, SuitHeart),
		NewCard(RankA, SuitDiamond),
	}

	// 对方出单张5
	target := NewPattern(Cards{NewCard(Rank5, SuitSpader)}, trump)
	results := handCards.SearchAll(target, trump)

	// 应该找到7,8,9,10,J,K,A共7种方案
	if len(results) < 5 {
		t.Errorf("期望找到至少5种压制方案，实际找到 %d", len(results))
	}
}

// TestCardsSearchAll_Bombs 测试 SearchAll 找到多个炸弹
func TestCardsSearchAll_Bombs(t *testing.T) {
	trump := Rank6
	handCards := Cards{
		NewCard(Rank3, SuitSpader),
		NewCard(Rank3, SuitHeart),
		NewCard(Rank3, SuitDiamond),
		NewCard(Rank3, SuitClub),
		NewCard(Rank7, SuitSpader),
		NewCard(Rank7, SuitHeart),
		NewCard(Rank7, SuitDiamond),
		NewCard(Rank7, SuitClub),
	}

	// 对方出对子5
	target := NewPattern(Cards{
		NewCard(Rank5, SuitSpader),
		NewCard(Rank5, SuitHeart),
	}, trump)
	results := handCards.SearchAll(target, trump)

	// 应该找到3和7两种4张炸弹
	if len(results) < 2 {
		t.Errorf("期望找到至少2种炸弹方案，实际找到 %d", len(results))
	}
}

// TestCardsSearch_NoValidMove 测试没有合法出牌
func TestCardsSearch_NoValidMove(t *testing.T) {
	trump := Rank6
	handCards := Cards{
		NewCard(Rank2, SuitSpader),
		NewCard(Rank3, SuitHeart),
		NewCard(Rank4, SuitDiamond),
	}

	// 对方出A
	target := NewPattern(Cards{NewCard(RankA, SuitSpader)}, trump)
	result := handCards.Search(target, trump)

	// 没有能压制的牌
	if result != nil {
		t.Error("期望没有能压制的牌")
	}
}

// TestCardsSearch_FullHouse 测试 Search 方法 - 三带二
func TestCardsSearch_FullHouse(t *testing.T) {
	trump := Rank6
	handCards := Cards{
		NewCard(RankK, SuitSpader),
		NewCard(RankK, SuitHeart),
		NewCard(RankK, SuitDiamond),
		NewCard(RankQ, SuitSpader),
		NewCard(RankQ, SuitHeart),
		NewCard(Rank5, SuitSpader),
		NewCard(Rank5, SuitHeart),
		NewCard(Rank5, SuitDiamond),
		NewCard(Rank3, SuitSpader),
		NewCard(Rank3, SuitHeart),
	}

	// 对方出三张7带两张4
	target := NewPattern(Cards{
		NewCard(Rank7, SuitSpader),
		NewCard(Rank7, SuitHeart),
		NewCard(Rank7, SuitDiamond),
		NewCard(Rank4, SuitSpader),
		NewCard(Rank4, SuitHeart),
	}, trump)
	result := handCards.Search(target, trump)

	if result == nil {
		t.Error("期望找到能压制的三带二")
		return
	}
	if len(result) != 5 {
		t.Errorf("期望5张牌，实际 %d", len(result))
	}
}

// TestCardsSearch_PairSeq 测试 Search 方法 - 连对
func TestCardsSearch_PairSeq(t *testing.T) {
	trump := Rank2
	handCards := Cards{
		NewCard(Rank5, SuitSpader),
		NewCard(Rank5, SuitHeart),
		NewCard(Rank6, SuitDiamond),
		NewCard(Rank6, SuitClub),
		NewCard(Rank7, SuitSpader),
		NewCard(Rank7, SuitHeart),
	}

	// 对方出3,4,5连对
	target := NewPattern(Cards{
		NewCard(Rank3, SuitSpader),
		NewCard(Rank3, SuitHeart),
		NewCard(Rank4, SuitDiamond),
		NewCard(Rank4, SuitClub),
		NewCard(Rank5, SuitSpader),
		NewCard(Rank5, SuitHeart),
	}, trump)
	result := handCards.Search(target, trump)

	if result == nil {
		t.Error("期望找到能压制的连对")
		return
	}
	if len(result) != 6 {
		t.Errorf("期望6张牌，实际 %d", len(result))
	}
}

// TestCardsSearch_TripsSeq 测试 Search 方法 - 三同连张
func TestCardsSearch_TripsSeq(t *testing.T) {
	trump := Rank2
	handCards := Cards{
		NewCard(Rank7, SuitSpader),
		NewCard(Rank7, SuitHeart),
		NewCard(Rank7, SuitDiamond),
		NewCard(Rank8, SuitSpader),
		NewCard(Rank8, SuitHeart),
		NewCard(Rank8, SuitDiamond),
	}

	// 对方出5,6三同连张
	target := NewPattern(Cards{
		NewCard(Rank5, SuitSpader),
		NewCard(Rank5, SuitHeart),
		NewCard(Rank5, SuitDiamond),
		NewCard(Rank6, SuitSpader),
		NewCard(Rank6, SuitHeart),
		NewCard(Rank6, SuitDiamond),
	}, trump)
	result := handCards.Search(target, trump)

	if result == nil {
		t.Error("期望找到能压制的三同连张")
		return
	}
	if len(result) != 6 {
		t.Errorf("期望6张牌，实际 %d", len(result))
	}
}

// TestCardsSearch_StraightFlush 测试 Search 方法 - 同花顺
func TestCardsSearch_StraightFlush(t *testing.T) {
	trump := Rank2
	handCards := Cards{
		NewCard(Rank7, SuitSpader),
		NewCard(Rank8, SuitSpader),
		NewCard(Rank9, SuitSpader),
		NewCard(Rank10, SuitSpader),
		NewCard(RankJ, SuitSpader),
		NewCard(Rank3, SuitHeart),
	}

	// 对方出同花顺3,4,5,6,7
	target := NewPattern(Cards{
		NewCard(Rank3, SuitDiamond),
		NewCard(Rank4, SuitDiamond),
		NewCard(Rank5, SuitDiamond),
		NewCard(Rank6, SuitDiamond),
		NewCard(Rank7, SuitDiamond),
	}, trump)
	result := handCards.Search(target, trump)

	if result == nil {
		t.Error("期望找到能压制的同花顺")
		return
	}
	if len(result) != 5 {
		t.Errorf("期望5张牌，实际 %d", len(result))
	}
	// 验证是同花
	firstSuit := result[0].Suit
	for _, c := range result {
		if c.Suit != firstSuit {
			t.Error("期望找到同花顺")
			break
		}
	}
}

// TestCardsSearch_UseBiggerBomb 测试用更大的炸弹压制
func TestCardsSearch_UseBiggerBomb(t *testing.T) {
	trump := Rank6
	handCards := Cards{
		NewCard(Rank5, SuitSpader),
		NewCard(Rank5, SuitHeart),
		NewCard(Rank5, SuitDiamond),
		NewCard(Rank5, SuitClub),
		NewCard(Rank5, SuitSpader),
	}

	// 对方出4张炸弹3
	target := NewPattern(Cards{
		NewCard(Rank3, SuitSpader),
		NewCard(Rank3, SuitHeart),
		NewCard(Rank3, SuitDiamond),
		NewCard(Rank3, SuitClub),
	}, trump)
	result := handCards.Search(target, trump)

	// 应该优先用4张5炸弹压制（同等级但点数更大）
	if result == nil {
		t.Error("期望用炸弹压制")
		return
	}
	if len(result) != 4 {
		t.Errorf("期望4张炸弹（同等级），实际 %d", len(result))
	}
}

// TestCardsSearch_UseHigherLevelBomb 测试用更高等级的炸弹压制
func TestCardsSearch_UseHigherLevelBomb(t *testing.T) {
	trump := Rank6
	handCards := Cards{
		NewCard(Rank2, SuitSpader),
		NewCard(Rank2, SuitHeart),
		NewCard(Rank2, SuitDiamond),
		NewCard(Rank2, SuitClub),
		NewCard(Rank2, SuitSpader), // 5张2
	}

	// 对方出4张炸弹A（更大的点数）
	target := NewPattern(Cards{
		NewCard(RankA, SuitSpader),
		NewCard(RankA, SuitHeart),
		NewCard(RankA, SuitDiamond),
		NewCard(RankA, SuitClub),
	}, trump)
	result := handCards.Search(target, trump)

	// 因为4张2无法压制4张A，所以应该用5张2炸弹（更高等级）
	if result == nil {
		t.Error("期望用5张炸弹压制")
		return
	}
	if len(result) != 5 {
		t.Errorf("期望5张炸弹（更高等级），实际 %d", len(result))
	}
}

// TestCardsSearch_LevelBomb 测试级牌炸弹
func TestCardsSearch_LevelBomb(t *testing.T) {
	trump := Rank6
	handCards := Cards{
		NewCard(Rank6, SuitSpader),
		NewCard(Rank6, SuitDiamond),
		NewCard(Rank6, SuitClub),
		NewCard(Rank6, SuitHeart), // 万能牌也算
		NewCard(Rank5, SuitSpader),
	}

	// 对方出对子3
	target := NewPattern(Cards{
		NewCard(Rank3, SuitSpader),
		NewCard(Rank3, SuitHeart),
	}, trump)
	result := handCards.Search(target, trump)

	// 应该用级牌炸弹压制
	if result == nil {
		t.Error("期望找到压制方案")
		return
	}
	// 可能是4张级牌炸弹
	if len(result) == 4 {
		for _, c := range result {
			if c.Rank != Rank6 {
				t.Error("期望级牌炸弹")
				break
			}
		}
	}
}

// TestCardsSearchAll_StraightFlush 测试 SearchAll - 同花顺
func TestCardsSearchAll_StraightFlush(t *testing.T) {
	trump := Rank2
	handCards := Cards{
		NewCard(Rank3, SuitSpader),
		NewCard(Rank4, SuitSpader),
		NewCard(Rank5, SuitSpader),
		NewCard(Rank6, SuitSpader),
		NewCard(Rank7, SuitSpader),
		NewCard(Rank8, SuitSpader),
		NewCard(Rank9, SuitSpader),
	}

	// 对方出普通顺子
	target := NewPattern(Cards{
		NewCard(Rank3, SuitHeart),
		NewCard(Rank4, SuitDiamond),
		NewCard(Rank5, SuitClub),
		NewCard(Rank6, SuitSpader),
		NewCard(Rank7, SuitHeart),
	}, trump)
	results := handCards.SearchAll(target, trump)

	// 应该找到多个同花顺方案
	if len(results) == 0 {
		t.Error("期望找到同花顺方案")
	}
}

// TestCardsSearchAll_FullHouse 测试 SearchAll - 多个三带二方案
func TestCardsSearchAll_FullHouse(t *testing.T) {
	trump := Rank6
	handCards := Cards{
		NewCard(RankA, SuitSpader),
		NewCard(RankA, SuitHeart),
		NewCard(RankA, SuitDiamond),
		NewCard(RankK, SuitSpader),
		NewCard(RankK, SuitHeart),
		NewCard(RankK, SuitDiamond),
		NewCard(RankQ, SuitSpader),
		NewCard(RankQ, SuitHeart),
	}

	// 对方出三张5带两张3
	target := NewPattern(Cards{
		NewCard(Rank5, SuitSpader),
		NewCard(Rank5, SuitHeart),
		NewCard(Rank5, SuitDiamond),
		NewCard(Rank3, SuitSpader),
		NewCard(Rank3, SuitHeart),
	}, trump)
	results := handCards.SearchAll(target, trump)

	// 应该找到多种三带二方案：A带K、A带Q、K带Q等
	if len(results) < 2 {
		t.Errorf("期望找到至少2种三带二方案，实际找到 %d", len(results))
	}
}

// TestSearchFullHouseAll 测试 searchFullHouseAll 函数
func TestSearchFullHouseAll(t *testing.T) {
	trump := Rank6
	cardsMap := map[Rank][]Card{
		RankA: {
			NewCard(RankA, SuitSpader),
			NewCard(RankA, SuitHeart),
			NewCard(RankA, SuitDiamond),
		},
		RankK: {
			NewCard(RankK, SuitSpader),
			NewCard(RankK, SuitHeart),
		},
		RankQ: {
			NewCard(RankQ, SuitSpader),
			NewCard(RankQ, SuitHeart),
			NewCard(RankQ, SuitDiamond),
		},
	}
	wildCards := []Card{}

	results := searchFullHouseAll(cardsMap, wildCards, 0, 0, trump)

	// 应该找到：A带K、Q带K、Q带A等多种组合
	if len(results) < 2 {
		t.Errorf("期望找到至少2种三带二组合，实际找到 %d", len(results))
	}
}

// TestCardsSearch_WithOnlyWildCards 测试只有万能牌的情况
func TestCardsSearch_WithOnlyWildCards(t *testing.T) {
	trump := Rank6
	handCards := Cards{
		NewCard(Rank6, SuitHeart),
		NewCard(Rank6, SuitHeart),
	}

	// 对方出单张5
	target := NewPattern(Cards{NewCard(Rank5, SuitSpader)}, trump)
	result := handCards.Search(target, trump)

	// 万能牌可以作为级牌（比5大）打出
	if result == nil {
		t.Error("期望万能牌能压制")
		return
	}
	if len(result) != 1 {
		t.Errorf("期望1张牌，实际 %d", len(result))
	}
}
