package guandan

import (
	"testing"
)

// TestRankWeight 测试 Rank.Weight 方法
func TestRankWeight(t *testing.T) {
	tests := []struct {
		name     string
		rank     Rank
		trump    Rank
		expected uint8
	}{
		{"普通牌2", Rank2, Rank6, uint8(Rank2)},
		{"普通牌A", RankA, Rank6, uint8(RankA)},
		{"级牌", Rank6, Rank6, uint8(RankLevel)},
		{"级牌2", Rank2, Rank2, uint8(RankLevel)},
		{"级牌A", RankA, RankA, uint8(RankLevel)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.rank.Weight(tt.trump)
			if result != tt.expected {
				t.Errorf("Rank.Weight() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestNewCard 测试 NewCard 函数
func TestNewCard(t *testing.T) {
	tests := []struct {
		name         string
		rank         Rank
		suit         Suit
		expectedRank Rank
		expectedSuit Suit
	}{
		{"黑桃A", RankA, SuitSpader, RankA, SuitSpader},
		{"红桃2", Rank2, SuitHeart, Rank2, SuitHeart},
		{"梅花K", RankK, SuitClub, RankK, SuitClub},
		{"方块10", Rank10, SuitDiamond, Rank10, SuitDiamond},
		{"小王", RankJokerSmall, SuitJoker, RankJokerSmall, SuitJoker},
		{"大王", RankJokerBig, SuitJoker, RankJokerBig, SuitJoker},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := NewCard(tt.rank, tt.suit)
			if card.Rank != tt.expectedRank {
				t.Errorf("NewCard().Rank = %v, want %v", card.Rank, tt.expectedRank)
			}
			if card.Suit != tt.expectedSuit {
				t.Errorf("NewCard().Suit = %v, want %v", card.Suit, tt.expectedSuit)
			}
		})
	}
}

// TestCardIsWild 测试 Card.IsWild 方法
func TestCardIsWild(t *testing.T) {
	tests := []struct {
		name     string
		card     Card
		trump    Rank
		expected bool
	}{
		{"红桃6且级牌6", NewCard(Rank6, SuitHeart), Rank6, true},
		{"黑桃6且级牌6", NewCard(Rank6, SuitSpader), Rank6, false},
		{"红桃A且级牌A", NewCard(RankA, SuitHeart), RankA, true},
		{"红桃2且级牌2", NewCard(Rank2, SuitHeart), Rank2, true},
		{"红桃6但级牌2", NewCard(Rank6, SuitHeart), Rank2, false},
		{"方块6且级牌6", NewCard(Rank6, SuitDiamond), Rank6, false},
		{"梅花6且级牌6", NewCard(Rank6, SuitClub), Rank6, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.card.IsWild(tt.trump)
			if result != tt.expected {
				t.Errorf("Card.IsWild() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCardsHasBigJoker 测试 Cards.HasBigJoker 方法
func TestCardsHasBigJoker(t *testing.T) {
	tests := []struct {
		name     string
		cards    Cards
		size     int
		expected bool
	}{
		{
			"有1个大王，需要1个",
			Cards{NewCard(RankJokerBig, SuitJoker)},
			1,
			true,
		},
		{
			"有2个大王，需要2个",
			Cards{
				NewCard(RankJokerBig, SuitJoker),
				NewCard(RankJokerBig, SuitJoker),
			},
			2,
			true,
		},
		{
			"有2个大王，需要1个",
			Cards{
				NewCard(RankJokerBig, SuitJoker),
				NewCard(RankJokerBig, SuitJoker),
			},
			1,
			true,
		},
		{
			"有1个大王，需要2个",
			Cards{NewCard(RankJokerBig, SuitJoker)},
			2,
			false,
		},
		{
			"没有大王，需要1个",
			Cards{
				NewCard(RankJokerSmall, SuitJoker),
				NewCard(RankA, SuitSpader),
			},
			1,
			false,
		},
		{
			"有大王和小王，需要1个大王",
			Cards{
				NewCard(RankJokerBig, SuitJoker),
				NewCard(RankJokerSmall, SuitJoker),
			},
			1,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cards.HasBigJoker(tt.size)
			if result != tt.expected {
				t.Errorf("Cards.HasBigJoker() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCardsHasFourJokers 测试 Cards.HasFourJokers 方法
func TestCardsHasFourJokers(t *testing.T) {
	tests := []struct {
		name     string
		cards    Cards
		expected bool
	}{
		{
			"正好四大天王",
			Cards{
				NewCard(RankJokerSmall, SuitJoker),
				NewCard(RankJokerSmall, SuitJoker),
				NewCard(RankJokerBig, SuitJoker),
				NewCard(RankJokerBig, SuitJoker),
			},
			true,
		},
		{
			"只有2个小王",
			Cards{
				NewCard(RankJokerSmall, SuitJoker),
				NewCard(RankJokerSmall, SuitJoker),
			},
			false,
		},
		{
			"只有2个大王",
			Cards{
				NewCard(RankJokerBig, SuitJoker),
				NewCard(RankJokerBig, SuitJoker),
			},
			false,
		},
		{
			"3个小王1个大王",
			Cards{
				NewCard(RankJokerSmall, SuitJoker),
				NewCard(RankJokerSmall, SuitJoker),
				NewCard(RankJokerSmall, SuitJoker),
				NewCard(RankJokerBig, SuitJoker),
			},
			false,
		},
		{
			"1个小王3个大王",
			Cards{
				NewCard(RankJokerSmall, SuitJoker),
				NewCard(RankJokerBig, SuitJoker),
				NewCard(RankJokerBig, SuitJoker),
				NewCard(RankJokerBig, SuitJoker),
			},
			false,
		},
		{
			"包含其他牌",
			Cards{
				NewCard(RankJokerSmall, SuitJoker),
				NewCard(RankJokerSmall, SuitJoker),
				NewCard(RankJokerBig, SuitJoker),
				NewCard(RankJokerBig, SuitJoker),
				NewCard(RankA, SuitSpader),
			},
			false,
		},
		{
			"空牌组",
			Cards{},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cards.HasFourJokers()
			if result != tt.expected {
				t.Errorf("Cards.HasFourJokers() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestSuitConstants 测试花色常量
func TestSuitConstants(t *testing.T) {
	if SuitNone >= SuitSpader {
		t.Error("SuitNone should be less than SuitSpader")
	}
	if SuitSpader >= SuitHeart {
		t.Error("SuitSpader should be less than SuitHeart")
	}
	if SuitHeart >= SuitClub {
		t.Error("SuitHeart should be less than SuitClub")
	}
	if SuitClub >= SuitDiamond {
		t.Error("SuitClub should be less than SuitDiamond")
	}
	if SuitDiamond >= SuitJoker {
		t.Error("SuitDiamond should be less than SuitJoker")
	}
}

// TestRankConstants 测试点数常量
func TestRankConstants(t *testing.T) {
	expectedOrder := []Rank{
		RankNone, Rank2, Rank3, Rank4, Rank5, Rank6, Rank7, Rank8, Rank9, Rank10,
		RankJ, RankQ, RankK, RankA, RankLevel, RankJokerSmall, RankJokerBig,
	}

	for i := range len(expectedOrder) - 1 {
		if expectedOrder[i] >= expectedOrder[i+1] {
			t.Errorf("Rank order error: %v should be less than %v", expectedOrder[i], expectedOrder[i+1])
		}
	}
}
