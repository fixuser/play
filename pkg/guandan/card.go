package guandan

import (
	"fmt"
	"math/rand/v2"
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

// MarshalBinary 序列化为二进制
func (c Card) MarshalBinary() (data []byte, err error) {
	data = make([]byte, 2)
	data[0] = byte(c.Rank)
	data[1] = byte(c.Suit)
	return
}

// UnmarshalBinary 从二进制反序列化
func (c *Card) UnmarshalBinary(data []byte) error {
	if len(data) < 2 {
		return fmt.Errorf("invalid card data")
	}
	c.Rank = Rank(data[0])
	c.Suit = Suit(data[1])
	return nil
}

// Equal 判断两张牌是否相等
func (c Card) Equal(other Card) bool {
	return c.Rank == other.Rank && c.Suit == other.Suit
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

// HasBigJoker 是否包含指定数量的大王
func (cs Cards) HasBigJoker(size int) bool {
	for _, c := range cs {
		if c.Rank == RankJokerBig {
			size--
		}
	}
	return size <= 0
}

// HasFourJokers 是否包含四大天王
func (cs Cards) HasFourJokers() bool {
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

// NewDeck 生成指定副数的扑克牌
// decks 表示几副扑克牌，每副牌包含 52 张普通牌 + 2 张大小王 = 54 张
func NewDeck(decks int) Cards {
	if decks <= 0 {
		return nil
	}

	// 每副牌 54 张
	cards := make(Cards, 0, decks*54)

	// 4 种花色
	suits := []Suit{SuitSpader, SuitHeart, SuitClub, SuitDiamond}
	// 13 种点数（2-A）
	ranks := []Rank{Rank2, Rank3, Rank4, Rank5, Rank6, Rank7, Rank8, Rank9, Rank10, RankJ, RankQ, RankK, RankA}

	for range decks {
		// 生成 52 张普通牌
		for _, suit := range suits {
			for _, rank := range ranks {
				cards = append(cards, NewCard(rank, suit))
			}
		}
		// 生成大小王
		cards = append(cards, NewCard(RankJokerSmall, SuitJoker))
		cards = append(cards, NewCard(RankJokerBig, SuitJoker))
	}

	return cards
}

// Shuffle 洗牌，随机打乱牌的顺序
func (cs Cards) Shuffle() {
	rand.Shuffle(len(cs), func(i, j int) {
		cs[i], cs[j] = cs[j], cs[i]
	})
}

// Deal 发牌，将牌随机发给指定数量的玩家
// players 表示玩家数量
// 返回每个玩家的手牌，如果牌数不能被玩家数整除，剩余的牌会被丢弃
func (cs Cards) Deal(players int) []Cards {
	if players <= 0 || len(cs) == 0 {
		return nil
	}

	// 先洗牌
	shuffled := make(Cards, len(cs))
	copy(shuffled, cs)
	shuffled.Shuffle()

	// 每个玩家的牌数
	cardsPerPlayer := len(shuffled) / players

	// 分配牌给每个玩家
	hands := make([]Cards, players)
	for i := range players {
		start := i * cardsPerPlayer
		end := start + cardsPerPlayer
		hands[i] = make(Cards, cardsPerPlayer)
		copy(hands[i], shuffled[start:end])
	}

	return hands
}
