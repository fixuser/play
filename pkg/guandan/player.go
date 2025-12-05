package guandan

type PlayStatus int8

const (
	StatusWaiting  PlayStatus = iota // 等待中
	StatusReady                      // 准备好
	StatusPlaying                    // 游戏中
	StatusFinished                   // 已结束
)

// Player 玩家信息
type Player struct {
	UserId int64      // 玩家ID
	Status PlayStatus // 玩家状态
	Hand   Cards      // 当前手里的牌
	Played Patterns   // 已经打出去的牌（记录每次打出的牌型）
	Rank   int8       // 玩家名次，可能是0（未完成），1，2，3，4
}

// NewPlayer 创建一个新玩家
func NewPlayer(UserId int64) *Player {
	return &Player{
		UserId: UserId,
	}
}

// SetHand 设置玩家手牌
func (p *Player) SetHand(cards Cards) {
	p.Hand = cards
}

// Play 打出指定的牌型
// 返回是否成功打出（手牌中是否有这些牌）
func (p *Player) Play(pattern *Pattern) bool {
	if pattern == nil || len(pattern.Cards) == 0 || pattern.Type == PatternTypeNone {
		return false
	}

	// 检查手牌中是否有这些牌
	handCopy := make(Cards, len(p.Hand))
	copy(handCopy, p.Hand)

	for _, card := range pattern.Cards {
		found := false
		for i, handCard := range handCopy {
			if handCard.Equal(card) {
				// 从手牌副本中移除
				handCopy = append(handCopy[:i], handCopy[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 更新手牌
	p.Hand = handCopy
	// 记录打出的牌型
	p.Played = append(p.Played, *pattern)
	return true
}

// HandCount 返回手牌数量
func (p *Player) HandCount() int {
	return len(p.Hand)
}

// PlayedCount 返回已打出牌的次数
func (p *Player) PlayedCount() int {
	return len(p.Played)
}

// PlayedCards 返回已打出的所有牌
func (p *Player) PlayedCards() Cards {
	var cards Cards
	for _, pattern := range p.Played {
		cards = append(cards, pattern.Cards...)
	}
	return cards
}

// PlayedCardCount 返回已打出牌的张数
func (p *Player) PlayedCardCount() int {
	count := 0
	for _, pattern := range p.Played {
		count += len(pattern.Cards)
	}
	return count
}

// TeamPlayers 表示一队玩家
type TeamPlayers [2]*Player

// Ranks 表示一队玩家的排名
func (tps *TeamPlayers) Ranks() TeamRank {
	return TeamRank{tps[0].Rank, tps[1].Rank}
}
