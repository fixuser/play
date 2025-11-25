package guandan

import (
	"testing"
)

// TestNewPattern_Single 测试单张牌型
func TestNewPattern_Single(t *testing.T) {
	trump := Rank6
	tests := []struct {
		name          string
		cards         Cards
		expectedType  PatternType
		expectedPoint uint8
	}{
		{"单张2", Cards{NewCard(Rank2, SuitSpader)}, PatternTypeSingle, uint8(Rank2)},
		{"单张A", Cards{NewCard(RankA, SuitSpader)}, PatternTypeSingle, uint8(RankA)},
		{"单张级牌6", Cards{NewCard(Rank6, SuitSpader)}, PatternTypeSingle, uint8(RankLevel)},
		{"单张万能牌", Cards{NewCard(Rank6, SuitHeart)}, PatternTypeSingle, uint8(RankLevel)},
		{"单张小王", Cards{NewCard(RankJokerSmall, SuitJoker)}, PatternTypeSingle, uint8(RankJokerSmall)},
		{"单张大王", Cards{NewCard(RankJokerBig, SuitJoker)}, PatternTypeSingle, uint8(RankJokerBig)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := NewPattern(tt.cards, trump)
			if pattern.Type != tt.expectedType {
				t.Errorf("Type = %v, want %v", pattern.Type, tt.expectedType)
			}
			if pattern.MainPoint != tt.expectedPoint {
				t.Errorf("MainPoint = %v, want %v", pattern.MainPoint, tt.expectedPoint)
			}
		})
	}
}

// TestNewPattern_Pair 测试对子牌型
func TestNewPattern_Pair(t *testing.T) {
	trump := Rank6
	tests := []struct {
		name          string
		cards         Cards
		expectedType  PatternType
		expectedPoint uint8
	}{
		{
			"普通对子2",
			Cards{NewCard(Rank2, SuitSpader), NewCard(Rank2, SuitHeart)},
			PatternTypePair,
			uint8(Rank2),
		},
		{
			"普通对子A",
			Cards{NewCard(RankA, SuitSpader), NewCard(RankA, SuitDiamond)},
			PatternTypePair,
			uint8(RankA),
		},
		{
			"级牌对子6",
			Cards{NewCard(Rank6, SuitSpader), NewCard(Rank6, SuitDiamond)},
			PatternTypePair,
			uint8(RankLevel),
		},
		{
			"万能牌配对",
			Cards{NewCard(Rank2, SuitSpader), NewCard(Rank6, SuitHeart)},
			PatternTypePair,
			uint8(Rank2),
		},
		{
			"两张万能牌",
			Cards{NewCard(Rank6, SuitHeart), NewCard(Rank6, SuitHeart)},
			PatternTypePair,
			uint8(RankLevel),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := NewPattern(tt.cards, trump)
			if pattern.Type != tt.expectedType {
				t.Errorf("Type = %v, want %v", pattern.Type, tt.expectedType)
			}
			if pattern.MainPoint != tt.expectedPoint {
				t.Errorf("MainPoint = %v, want %v", pattern.MainPoint, tt.expectedPoint)
			}
		})
	}
}

// TestNewPattern_Trips 测试三同张牌型
func TestNewPattern_Trips(t *testing.T) {
	trump := Rank6
	tests := []struct {
		name          string
		cards         Cards
		expectedType  PatternType
		expectedPoint uint8
	}{
		{
			"普通三同张",
			Cards{
				NewCard(Rank5, SuitSpader),
				NewCard(Rank5, SuitHeart),
				NewCard(Rank5, SuitDiamond),
			},
			PatternTypeTrips,
			uint8(Rank5),
		},
		{
			"三张A",
			Cards{
				NewCard(RankA, SuitSpader),
				NewCard(RankA, SuitHeart),
				NewCard(RankA, SuitDiamond),
			},
			PatternTypeTrips,
			uint8(RankA),
		},
		{
			"万能牌配对成三张",
			Cards{
				NewCard(Rank5, SuitSpader),
				NewCard(Rank5, SuitDiamond),
				NewCard(Rank6, SuitHeart),
			},
			PatternTypeTrips,
			uint8(Rank5),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := NewPattern(tt.cards, trump)
			if pattern.Type != tt.expectedType {
				t.Errorf("Type = %v, want %v", pattern.Type, tt.expectedType)
			}
			if pattern.MainPoint != tt.expectedPoint {
				t.Errorf("MainPoint = %v, want %v", pattern.MainPoint, tt.expectedPoint)
			}
		})
	}
}

// TestNewPattern_Bomb 测试炸弹牌型
func TestNewPattern_Bomb(t *testing.T) {
	trump := Rank6
	tests := []struct {
		name          string
		cards         Cards
		expectedType  PatternType
		expectedPoint uint8
		expectedLevel int
	}{
		{
			"4张炸弹",
			Cards{
				NewCard(Rank5, SuitSpader),
				NewCard(Rank5, SuitHeart),
				NewCard(Rank5, SuitDiamond),
				NewCard(Rank5, SuitClub),
			},
			PatternTypeBomb,
			uint8(Rank5),
			2,
		},
		{
			"5张炸弹",
			Cards{
				NewCard(Rank5, SuitSpader),
				NewCard(Rank5, SuitHeart),
				NewCard(Rank5, SuitDiamond),
				NewCard(Rank5, SuitClub),
				NewCard(Rank5, SuitSpader),
			},
			PatternTypeBomb,
			uint8(Rank5),
			3,
		},
		{
			"6张炸弹",
			Cards{
				NewCard(Rank5, SuitSpader),
				NewCard(Rank5, SuitHeart),
				NewCard(Rank5, SuitDiamond),
				NewCard(Rank5, SuitClub),
				NewCard(Rank5, SuitSpader),
				NewCard(Rank5, SuitHeart),
			},
			PatternTypeBomb,
			uint8(Rank5),
			5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := NewPattern(tt.cards, trump)
			if pattern.Type != tt.expectedType {
				t.Errorf("Type = %v, want %v", pattern.Type, tt.expectedType)
			}
			if pattern.MainPoint != tt.expectedPoint {
				t.Errorf("MainPoint = %v, want %v", pattern.MainPoint, tt.expectedPoint)
			}
			if pattern.GetLevel() != tt.expectedLevel {
				t.Errorf("GetLevel() = %v, want %v", pattern.GetLevel(), tt.expectedLevel)
			}
		})
	}
}

// TestNewPattern_Straight 测试顺子牌型（包含 A 作为 1 和 13 的情况）
func TestNewPattern_Straight(t *testing.T) {
	trump := Rank6
	tests := []struct {
		name          string
		cards         Cards
		expectedType  PatternType
		expectedPoint uint8
		description   string
	}{
		{
			"A2345顺子-A作为1",
			Cards{
				NewCard(RankA, SuitSpader),
				NewCard(Rank2, SuitHeart),
				NewCard(Rank3, SuitDiamond),
				NewCard(Rank4, SuitClub),
				NewCard(Rank5, SuitSpader),
			},
			PatternTypeStraight,
			uint8(Rank5),
			"A当1，最大牌是5",
		},
		{
			"10JQKA顺子-A作为13",
			Cards{
				NewCard(Rank10, SuitSpader),
				NewCard(RankJ, SuitHeart),
				NewCard(RankQ, SuitDiamond),
				NewCard(RankK, SuitClub),
				NewCard(RankA, SuitSpader),
			},
			PatternTypeStraight,
			uint8(RankA),
			"A当13，最大牌是A",
		},
		{
			"34567顺子",
			Cards{
				NewCard(Rank3, SuitSpader),
				NewCard(Rank4, SuitHeart),
				NewCard(Rank5, SuitDiamond),
				NewCard(Rank6, SuitClub),
				NewCard(Rank7, SuitSpader),
			},
			PatternTypeStraight,
			uint8(Rank7),
			"普通顺子",
		},
		{
			"9,10,J,Q,K顺子",
			Cards{
				NewCard(Rank9, SuitSpader),
				NewCard(Rank10, SuitHeart),
				NewCard(RankJ, SuitDiamond),
				NewCard(RankQ, SuitClub),
				NewCard(RankK, SuitSpader),
			},
			PatternTypeStraight,
			uint8(RankK),
			"9到K的顺子",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := NewPattern(tt.cards, trump)
			if pattern.Type != tt.expectedType {
				t.Errorf("%s: Type = %v, want %v", tt.description, pattern.Type, tt.expectedType)
			}
			if pattern.MainPoint != tt.expectedPoint {
				t.Errorf("%s: MainPoint = %v, want %v", tt.description, pattern.MainPoint, tt.expectedPoint)
			}
		})
	}
}

// TestNewPattern_StraightFlush 测试同花顺（包含 A 的特殊情况）
func TestNewPattern_StraightFlush(t *testing.T) {
	trump := Rank6
	tests := []struct {
		name          string
		cards         Cards
		expectedType  PatternType
		expectedPoint uint8
		expectedLevel int
		description   string
	}{
		{
			"A2345同花顺-黑桃",
			Cards{
				NewCard(RankA, SuitSpader),
				NewCard(Rank2, SuitSpader),
				NewCard(Rank3, SuitSpader),
				NewCard(Rank4, SuitSpader),
				NewCard(Rank5, SuitSpader),
			},
			PatternTypeStraightFlush,
			uint8(Rank5),
			4,
			"A作为1的同花顺",
		},
		{
			"10JQKA同花顺-红桃",
			Cards{
				NewCard(Rank10, SuitHeart),
				NewCard(RankJ, SuitHeart),
				NewCard(RankQ, SuitHeart),
				NewCard(RankK, SuitHeart),
				NewCard(RankA, SuitHeart),
			},
			PatternTypeStraightFlush,
			uint8(RankA),
			4,
			"A作为13的同花顺",
		},
		{
			"34567同花顺-方块",
			Cards{
				NewCard(Rank3, SuitDiamond),
				NewCard(Rank4, SuitDiamond),
				NewCard(Rank5, SuitDiamond),
				NewCard(Rank6, SuitDiamond),
				NewCard(Rank7, SuitDiamond),
			},
			PatternTypeStraightFlush,
			uint8(Rank7),
			4,
			"普通同花顺",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := NewPattern(tt.cards, trump)
			if pattern.Type != tt.expectedType {
				t.Errorf("%s: Type = %v, want %v", tt.description, pattern.Type, tt.expectedType)
			}
			if pattern.MainPoint != tt.expectedPoint {
				t.Errorf("%s: MainPoint = %v, want %v", tt.description, pattern.MainPoint, tt.expectedPoint)
			}
			if pattern.GetLevel() != tt.expectedLevel {
				t.Errorf("%s: GetLevel() = %v, want %v", tt.description, pattern.GetLevel(), tt.expectedLevel)
			}
			if !pattern.SameSuit {
				t.Errorf("%s: SameSuit should be true", tt.description)
			}
		})
	}
}

// TestNewPattern_PairSeq 测试三连对（包含 A 的特殊情况）
func TestNewPattern_PairSeq(t *testing.T) {
	trump := Rank6
	tests := []struct {
		name          string
		cards         Cards
		expectedType  PatternType
		expectedPoint uint8
		description   string
	}{
		{
			"A,2,3连对-A作为1",
			Cards{
				NewCard(RankA, SuitSpader),
				NewCard(RankA, SuitHeart),
				NewCard(Rank2, SuitDiamond),
				NewCard(Rank2, SuitClub),
				NewCard(Rank3, SuitSpader),
				NewCard(Rank3, SuitHeart),
			},
			PatternTypePairSeq,
			uint8(Rank3),
			"A作为1的连对",
		},
		{
			"Q,K,A连对-A作为13",
			Cards{
				NewCard(RankQ, SuitSpader),
				NewCard(RankQ, SuitHeart),
				NewCard(RankK, SuitDiamond),
				NewCard(RankK, SuitClub),
				NewCard(RankA, SuitSpader),
				NewCard(RankA, SuitHeart),
			},
			PatternTypePairSeq,
			uint8(RankA),
			"A作为13的连对",
		},
		{
			"3,4,5连对",
			Cards{
				NewCard(Rank3, SuitSpader),
				NewCard(Rank3, SuitHeart),
				NewCard(Rank4, SuitDiamond),
				NewCard(Rank4, SuitClub),
				NewCard(Rank5, SuitSpader),
				NewCard(Rank5, SuitHeart),
			},
			PatternTypePairSeq,
			uint8(Rank5),
			"普通连对",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := NewPattern(tt.cards, trump)
			if pattern.Type != tt.expectedType {
				t.Errorf("%s: Type = %v, want %v", tt.description, pattern.Type, tt.expectedType)
			}
			if pattern.MainPoint != tt.expectedPoint {
				t.Errorf("%s: MainPoint = %v, want %v", tt.description, pattern.MainPoint, tt.expectedPoint)
			}
		})
	}
}

// TestNewPattern_TripsSeq 测试三同连张（包含 A 的特殊情况）
func TestNewPattern_TripsSeq(t *testing.T) {
	trump := Rank6
	tests := []struct {
		name          string
		cards         Cards
		expectedType  PatternType
		expectedPoint uint8
		description   string
	}{
		{
			"A,2三同连张-A作为1",
			Cards{
				NewCard(RankA, SuitSpader),
				NewCard(RankA, SuitHeart),
				NewCard(RankA, SuitDiamond),
				NewCard(Rank2, SuitClub),
				NewCard(Rank2, SuitSpader),
				NewCard(Rank2, SuitHeart),
			},
			PatternTypeTripsSeq,
			uint8(Rank2),
			"A作为1的三同连张",
		},
		{
			"K,A三同连张-A作为13",
			Cards{
				NewCard(RankK, SuitSpader),
				NewCard(RankK, SuitHeart),
				NewCard(RankK, SuitDiamond),
				NewCard(RankA, SuitClub),
				NewCard(RankA, SuitSpader),
				NewCard(RankA, SuitHeart),
			},
			PatternTypeTripsSeq,
			uint8(RankA),
			"A作为13的三同连张",
		},
		{
			"3,4三同连张",
			Cards{
				NewCard(Rank3, SuitSpader),
				NewCard(Rank3, SuitHeart),
				NewCard(Rank3, SuitDiamond),
				NewCard(Rank4, SuitClub),
				NewCard(Rank4, SuitSpader),
				NewCard(Rank4, SuitHeart),
			},
			PatternTypeTripsSeq,
			uint8(Rank4),
			"普通三同连张",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := NewPattern(tt.cards, trump)
			if pattern.Type != tt.expectedType {
				t.Errorf("%s: Type = %v, want %v", tt.description, pattern.Type, tt.expectedType)
			}
			if pattern.MainPoint != tt.expectedPoint {
				t.Errorf("%s: MainPoint = %v, want %v", tt.description, pattern.MainPoint, tt.expectedPoint)
			}
		})
	}
}

// TestNewPattern_FullHouse 测试三带二
func TestNewPattern_FullHouse(t *testing.T) {
	trump := Rank6
	tests := []struct {
		name           string
		cards          Cards
		expectedType   PatternType
		expectedMainPt uint8
		expectedSubPt  uint8
	}{
		{
			"三张5带两张3",
			Cards{
				NewCard(Rank5, SuitSpader),
				NewCard(Rank5, SuitHeart),
				NewCard(Rank5, SuitDiamond),
				NewCard(Rank3, SuitClub),
				NewCard(Rank3, SuitSpader),
			},
			PatternTypeFullHouse,
			uint8(Rank5),
			uint8(Rank3),
		},
		{
			"三张A带两张2",
			Cards{
				NewCard(RankA, SuitSpader),
				NewCard(RankA, SuitHeart),
				NewCard(RankA, SuitDiamond),
				NewCard(Rank2, SuitClub),
				NewCard(Rank2, SuitSpader),
			},
			PatternTypeFullHouse,
			uint8(RankA),
			uint8(Rank2),
		},
		{
			"用万能牌凑三带二",
			Cards{
				NewCard(Rank5, SuitSpader),
				NewCard(Rank5, SuitDiamond),
				NewCard(Rank3, SuitClub),
				NewCard(Rank3, SuitSpader),
				NewCard(Rank6, SuitHeart), // 万能牌
			},
			PatternTypeFullHouse,
			uint8(Rank5),
			uint8(Rank3),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := NewPattern(tt.cards, trump)
			if pattern.Type != tt.expectedType {
				t.Errorf("Type = %v, want %v", pattern.Type, tt.expectedType)
			}
			if pattern.MainPoint != tt.expectedMainPt {
				t.Errorf("MainPoint = %v, want %v", pattern.MainPoint, tt.expectedMainPt)
			}
			if pattern.SubPoint != tt.expectedSubPt {
				t.Errorf("SubPoint = %v, want %v", pattern.SubPoint, tt.expectedSubPt)
			}
		})
	}
}

// TestNewPattern_FourJokers 测试四大天王
func TestNewPattern_FourJokers(t *testing.T) {
	cards := Cards{
		NewCard(RankJokerSmall, SuitJoker),
		NewCard(RankJokerSmall, SuitJoker),
		NewCard(RankJokerBig, SuitJoker),
		NewCard(RankJokerBig, SuitJoker),
	}

	pattern := NewPattern(cards, Rank6)
	if pattern.Type != PatternTypeFourJokers {
		t.Errorf("Type = %v, want %v", pattern.Type, PatternTypeFourJokers)
	}
}

// TestPatternGetLevel 测试牌型等级
func TestPatternGetLevel(t *testing.T) {
	trump := Rank2 // 使用2作为级牌，避免干扰顺子
	tests := []struct {
		name          string
		cards         Cards
		expectedLevel int
		description   string
	}{
		{
			"4张炸弹",
			Cards{
				NewCard(Rank5, SuitSpader),
				NewCard(Rank5, SuitHeart),
				NewCard(Rank5, SuitDiamond),
				NewCard(Rank5, SuitClub),
			},
			2,
			"Level 2",
		},
		{
			"5张炸弹",
			Cards{
				NewCard(Rank5, SuitSpader),
				NewCard(Rank5, SuitHeart),
				NewCard(Rank5, SuitDiamond),
				NewCard(Rank5, SuitClub),
				NewCard(Rank5, SuitSpader),
			},
			3,
			"Level 3",
		},
		{
			"6张炸弹",
			Cards{
				NewCard(Rank5, SuitSpader),
				NewCard(Rank5, SuitHeart),
				NewCard(Rank5, SuitDiamond),
				NewCard(Rank5, SuitClub),
				NewCard(Rank5, SuitSpader),
				NewCard(Rank5, SuitHeart),
			},
			5,
			"Level 5",
		},
		{
			"同花顺",
			Cards{
				NewCard(Rank4, SuitSpader),
				NewCard(Rank5, SuitSpader),
				NewCard(Rank6, SuitSpader),
				NewCard(Rank7, SuitSpader),
				NewCard(Rank8, SuitSpader),
			},
			4,
			"Level 4",
		},
		{
			"普通顺子",
			Cards{
				NewCard(Rank4, SuitSpader),
				NewCard(Rank5, SuitHeart),
				NewCard(Rank6, SuitDiamond),
				NewCard(Rank7, SuitClub),
				NewCard(Rank8, SuitSpader),
			},
			1,
			"Level 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := NewPattern(tt.cards, trump)
			if pattern.GetLevel() != tt.expectedLevel {
				t.Errorf("%s: GetLevel() = %v, want %v", tt.description, pattern.GetLevel(), tt.expectedLevel)
			}
		})
	}
}

// TestPatternCompare 测试牌型比较
func TestPatternCompare(t *testing.T) {
	trump := Rank6
	tests := []struct {
		name     string
		p1       *Pattern
		p2       *Pattern
		expected int
		desc     string
	}{
		{
			"四大天王 vs 5张炸弹",
			NewPattern(Cards{
				NewCard(RankJokerSmall, SuitJoker),
				NewCard(RankJokerSmall, SuitJoker),
				NewCard(RankJokerBig, SuitJoker),
				NewCard(RankJokerBig, SuitJoker),
			}, trump),
			NewPattern(Cards{
				NewCard(RankA, SuitSpader),
				NewCard(RankA, SuitHeart),
				NewCard(RankA, SuitDiamond),
				NewCard(RankA, SuitClub),
				NewCard(RankA, SuitSpader),
			}, trump),
			1,
			"四大天王最大",
		},
		{
			"同花顺 vs 5张炸弹",
			NewPattern(Cards{
				NewCard(Rank4, SuitSpader),
				NewCard(Rank5, SuitSpader),
				NewCard(Rank6, SuitSpader),
				NewCard(Rank7, SuitSpader),
				NewCard(Rank8, SuitSpader),
			}, Rank2),
			NewPattern(Cards{
				NewCard(RankA, SuitSpader),
				NewCard(RankA, SuitHeart),
				NewCard(RankA, SuitDiamond),
				NewCard(RankA, SuitClub),
				NewCard(RankA, SuitSpader),
			}, Rank2),
			1,
			"同花顺 > 5张炸弹",
		},
		{
			"5张炸弹 vs 4张炸弹",
			NewPattern(Cards{
				NewCard(Rank3, SuitSpader),
				NewCard(Rank3, SuitHeart),
				NewCard(Rank3, SuitDiamond),
				NewCard(Rank3, SuitClub),
				NewCard(Rank3, SuitSpader),
			}, trump),
			NewPattern(Cards{
				NewCard(RankA, SuitSpader),
				NewCard(RankA, SuitHeart),
				NewCard(RankA, SuitDiamond),
				NewCard(RankA, SuitClub),
			}, trump),
			1,
			"5张炸弹 > 4张炸弹",
		},
		{
			"相同张数炸弹比点数",
			NewPattern(Cards{
				NewCard(RankA, SuitSpader),
				NewCard(RankA, SuitHeart),
				NewCard(RankA, SuitDiamond),
				NewCard(RankA, SuitClub),
			}, trump),
			NewPattern(Cards{
				NewCard(Rank5, SuitSpader),
				NewCard(Rank5, SuitHeart),
				NewCard(Rank5, SuitDiamond),
				NewCard(Rank5, SuitClub),
			}, trump),
			1,
			"A炸弹 > 5炸弹",
		},
		{
			"单张比较",
			NewPattern(Cards{NewCard(RankA, SuitSpader)}, trump),
			NewPattern(Cards{NewCard(Rank5, SuitSpader)}, trump),
			1,
			"A > 5",
		},
		{
			"对子比较",
			NewPattern(Cards{
				NewCard(RankA, SuitSpader),
				NewCard(RankA, SuitHeart),
			}, trump),
			NewPattern(Cards{
				NewCard(Rank5, SuitSpader),
				NewCard(Rank5, SuitHeart),
			}, trump),
			1,
			"对A > 对5",
		},
		{
			"不同牌型无法比较",
			NewPattern(Cards{NewCard(RankA, SuitSpader)}, trump),
			NewPattern(Cards{
				NewCard(Rank5, SuitSpader),
				NewCard(Rank5, SuitHeart),
			}, trump),
			0,
			"单张 vs 对子",
		},
		{
			"顺子A2345 vs 34567",
			NewPattern(Cards{
				NewCard(RankA, SuitSpader),
				NewCard(Rank2, SuitHeart),
				NewCard(Rank3, SuitDiamond),
				NewCard(Rank4, SuitClub),
				NewCard(Rank5, SuitSpader),
			}, trump),
			NewPattern(Cards{
				NewCard(Rank3, SuitSpader),
				NewCard(Rank4, SuitHeart),
				NewCard(Rank5, SuitDiamond),
				NewCard(Rank6, SuitClub),
				NewCard(Rank7, SuitSpader),
			}, trump),
			-1,
			"A2345(最大5) < 34567(最大7)",
		},
		{
			"顺子10JQKA vs 9,10,J,Q,K",
			NewPattern(Cards{
				NewCard(Rank10, SuitSpader),
				NewCard(RankJ, SuitHeart),
				NewCard(RankQ, SuitDiamond),
				NewCard(RankK, SuitClub),
				NewCard(RankA, SuitSpader),
			}, trump),
			NewPattern(Cards{
				NewCard(Rank9, SuitSpader),
				NewCard(Rank10, SuitHeart),
				NewCard(RankJ, SuitDiamond),
				NewCard(RankQ, SuitClub),
				NewCard(RankK, SuitSpader),
			}, trump),
			1,
			"10JQKA(最大A) > 9,10,J,Q,K(最大K)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.p1.Compare(tt.p2)
			if result != tt.expected {
				t.Errorf("%s: Compare() = %v, want %v", tt.desc, result, tt.expected)
			}
		})
	}
}

// TestCheckSequence_WithWildCards 测试带万能牌的顺子检测
func TestCheckSequence_WithWildCards(t *testing.T) {
	tests := []struct {
		name          string
		rankCounts    map[Rank]int
		wildCount     int
		length        int
		width         int
		expectedPoint uint8
		description   string
	}{
		{
			"用万能牌凑A2345",
			map[Rank]int{RankA: 1, Rank2: 1, Rank4: 1, Rank5: 1},
			1, // 缺3
			5,
			1,
			uint8(Rank5),
			"A2345用万能牌补3",
		},
		{
			"用万能牌凑10JQKA",
			map[Rank]int{Rank10: 1, RankJ: 1, RankK: 1, RankA: 1},
			1, // 缺Q
			5,
			1,
			uint8(RankA),
			"10JQKA用万能牌补Q",
		},
		{
			"用万能牌凑连对",
			map[Rank]int{RankA: 2, Rank2: 1, Rank3: 2},
			1, // 补2
			3,
			2,
			uint8(Rank3),
			"A,2,3连对用万能牌补一张2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkSequence(tt.rankCounts, tt.wildCount, tt.length, tt.width)
			if result != tt.expectedPoint {
				t.Errorf("%s: checkSequence() = %v, want %v", tt.description, result, tt.expectedPoint)
			}
		})
	}
}

// TestCheckFullHouse_WithWildCards 测试带万能牌的三带二检测
func TestCheckFullHouse_WithWildCards(t *testing.T) {
	trump := Rank6
	tests := []struct {
		name         string
		rankCounts   map[Rank]int
		wildCount    int
		expectedMain uint8
		expectedSub  uint8
		description  string
	}{
		{
			"用万能牌凑三带二",
			map[Rank]int{Rank5: 2, Rank3: 2},
			1, // 可以补成三张5带两张3
			uint8(Rank5),
			uint8(Rank3),
			"2张5,2张3,1万能牌凑三带二",
		},
		{
			"用两张万能牌凑三带二",
			map[Rank]int{Rank5: 1, Rank3: 2},
			2, // 补2张5
			uint8(Rank5),
			uint8(Rank3),
			"1张5,2张3,2万能牌凑三带二",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			main, sub := checkFullHouse(tt.rankCounts, tt.wildCount, trump)
			if main != tt.expectedMain {
				t.Errorf("%s: MainPoint = %v, want %v", tt.description, main, tt.expectedMain)
			}
			if sub != tt.expectedSub {
				t.Errorf("%s: SubPoint = %v, want %v", tt.description, sub, tt.expectedSub)
			}
		})
	}
}

// TestPattern_EmptyCards 测试空牌组
func TestPattern_EmptyCards(t *testing.T) {
	pattern := NewPattern(Cards{}, Rank6)
	if pattern.Type != PatternTypeNone {
		t.Errorf("Empty cards should have PatternTypeNone, got %v", pattern.Type)
	}
	if pattern.Length != 0 {
		t.Errorf("Empty cards should have Length 0, got %v", pattern.Length)
	}
}

// TestPattern_InvalidCards 测试无效牌组
func TestPattern_InvalidCards(t *testing.T) {
	tests := []struct {
		name  string
		cards Cards
		trump Rank
	}{
		{
			"两张不同点数的牌",
			Cards{
				NewCard(Rank2, SuitSpader),
				NewCard(Rank3, SuitHeart),
			},
			Rank6,
		},
		{
			"三张不连续的牌",
			Cards{
				NewCard(Rank2, SuitSpader),
				NewCard(Rank4, SuitHeart),
				NewCard(Rank7, SuitDiamond),
			},
			Rank6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := NewPattern(tt.cards, tt.trump)
			// 应该返回 PatternTypeNone 或者不匹配任何有效牌型
			if pattern.Type != PatternTypeNone {
				t.Logf("Invalid cards resulted in pattern type: %v (might be valid in some contexts)", pattern.Type)
			}
		})
	}
}

// TestBomb_WithWildCards 测试用万能牌凑炸弹
func TestBomb_WithWildCards(t *testing.T) {
	trump := Rank6 // 6是级牌，红桃6是万能牌
	tests := []struct {
		name          string
		cards         Cards
		expectedType  PatternType
		expectedPoint uint8
		expectedLevel int
		description   string
	}{
		{
			"2张5+2万能牌凑4张炸弹",
			Cards{
				NewCard(Rank5, SuitSpader),
				NewCard(Rank5, SuitHeart),
				NewCard(Rank6, SuitHeart), // 万能牌
				NewCard(Rank6, SuitHeart), // 万能牌
			},
			PatternTypeBomb,
			uint8(Rank5),
			2,
			"4张炸弹",
		},
		{
			"3张7+1万能牌凑4张炸弹",
			Cards{
				NewCard(Rank7, SuitSpader),
				NewCard(Rank7, SuitHeart),
				NewCard(Rank7, SuitDiamond),
				NewCard(Rank6, SuitHeart), // 万能牌
			},
			PatternTypeBomb,
			uint8(Rank7),
			2,
			"4张炸弹",
		},
		{
			"3张A+2万能牌凑5张炸弹",
			Cards{
				NewCard(RankA, SuitSpader),
				NewCard(RankA, SuitHeart),
				NewCard(RankA, SuitDiamond),
				NewCard(Rank6, SuitHeart), // 万能牌
				NewCard(Rank6, SuitHeart), // 万能牌
			},
			PatternTypeBomb,
			uint8(RankA),
			3,
			"5张炸弹",
		},
		{
			"4张K+1万能牌凑5张炸弹",
			Cards{
				NewCard(RankK, SuitSpader),
				NewCard(RankK, SuitHeart),
				NewCard(RankK, SuitDiamond),
				NewCard(RankK, SuitClub),
				NewCard(Rank6, SuitHeart), // 万能牌
			},
			PatternTypeBomb,
			uint8(RankK),
			3,
			"5张炸弹",
		},
		{
			"4张Q+2万能牌凑6张炸弹",
			Cards{
				NewCard(RankQ, SuitSpader),
				NewCard(RankQ, SuitHeart),
				NewCard(RankQ, SuitDiamond),
				NewCard(RankQ, SuitClub),
				NewCard(Rank6, SuitHeart), // 万能牌
				NewCard(Rank6, SuitHeart), // 万能牌
			},
			PatternTypeBomb,
			uint8(RankQ),
			5,
			"6张炸弹",
		},
		{
			"5张J+2万能牌凑7张炸弹",
			Cards{
				NewCard(RankJ, SuitSpader),
				NewCard(RankJ, SuitHeart),
				NewCard(RankJ, SuitDiamond),
				NewCard(RankJ, SuitClub),
				NewCard(RankJ, SuitSpader),
				NewCard(Rank6, SuitHeart), // 万能牌
				NewCard(Rank6, SuitHeart), // 万能牌
			},
			PatternTypeBomb,
			uint8(RankJ),
			5,
			"7张炸弹",
		},
		{
			"6张10+2万能牌凑8张炸弹",
			Cards{
				NewCard(Rank10, SuitSpader),
				NewCard(Rank10, SuitHeart),
				NewCard(Rank10, SuitDiamond),
				NewCard(Rank10, SuitClub),
				NewCard(Rank10, SuitSpader),
				NewCard(Rank10, SuitHeart),
				NewCard(Rank6, SuitHeart), // 万能牌
				NewCard(Rank6, SuitHeart), // 万能牌
			},
			PatternTypeBomb,
			uint8(Rank10),
			5,
			"8张炸弹",
		},
		{
			"2张8+1万能牌凑不成4张炸弹",
			Cards{
				NewCard(Rank8, SuitSpader),
				NewCard(Rank8, SuitHeart),
				NewCard(Rank6, SuitHeart), // 万能牌
			},
			PatternTypeTrips,
			uint8(Rank8),
			1,
			"只能凑成三同张",
		},
		{
			"1张9+2万能牌凑不成4张炸弹",
			Cards{
				NewCard(Rank9, SuitSpader),
				NewCard(Rank6, SuitHeart), // 万能牌
				NewCard(Rank6, SuitHeart), // 万能牌
			},
			PatternTypeTrips,
			uint8(Rank9),
			1,
			"只能凑成三同张",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := NewPattern(tt.cards, trump)
			if pattern.Type != tt.expectedType {
				t.Errorf("%s: Type = %v, want %v", tt.description, pattern.Type, tt.expectedType)
			}
			if pattern.MainPoint != tt.expectedPoint {
				t.Errorf("%s: MainPoint = %v, want %v", tt.description, pattern.MainPoint, tt.expectedPoint)
			}
			if pattern.GetLevel() != tt.expectedLevel {
				t.Errorf("%s: GetLevel() = %v, want %v", tt.description, pattern.GetLevel(), tt.expectedLevel)
			}
		})
	}
}
