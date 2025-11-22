package guandan

// Suit 牌的花色
type Suit uint8

const (
	SuitNone    Suit = iota
	SuitSpader       // 黑桃
	SuitHeart        // 红桃
	SuitClub         // 梅花
	SuitDiamond      // 方块
	SuitJoker        // 王
)

// Rank 牌的点数
type Rank uint8

const (
	RankNone Rank = iota
	Rank2
	Rank3
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8
	Rank9
	Rank10
	RankJ
	RankQ
	RankK
	RankA
	RankLevel
	RankJokerSmall
	RankJokerBig
)

// Weight 返回牌的权重
// trump 表示当前的级牌
func (r Rank) Weight(trump Rank) uint8 {
	if r == trump {
		return uint8(RankLevel)
	}
	return uint8(r)
}

// Card 代表一张扑克牌
type Card struct {
	Rank Rank
	Suit Suit
}

// NewCard
func NewCard(rank Rank, suit Suit) Card {
	return Card{
		Rank: rank,
		Suit: suit,
	}
}

// IsWild 判断是否为万能牌（红桃级牌）
func (c Card) IsWild(trump Rank) bool {
	return c.Rank == trump && c.Suit == SuitHeart
}

type Cards []Card

// IsFourJokers 是否是四大天王
func (cs Cards) IsFourJokers() bool {
	cntSmall, cntBig := 0, 0
	for _, c := range cs {
		switch c.Rank {
		case RankJokerSmall:
			cntSmall++
		case RankJokerBig:
			cntBig++
		default:
			return false
		}
	}
	return cntSmall == 2 && cntBig == 2
}
