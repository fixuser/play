package guandan

import "time"

type GameRoundStatus int8

const (
	GameStatusWaiting  GameRoundStatus = iota // 等待中
	GameStatusPlaying                         // 游戏中
	GameStatusFinished                        // 已结束
)

// TeamRank 团队的排名情况
type TeamRank [2]int8

// IsWinner 判断是否是赢的队伍（有一个人是第1名）
func (tr TeamRank) IsWinner() bool {
	return tr[0] == 1 || tr[1] == 1
}

// WinLevel 获取赢的等级
// 返回 3: 双上(1,2), 2: 中等胜利(1,3), 1: 普通胜利(1,4)
// 如果不是赢的队伍返回 0
func (tr TeamRank) WinLevel() int {
	if !tr.IsWinner() {
		return 0
	}
	// 获取队友的名次（非第1名的那个）
	teammateRank := tr[0]
	if tr[0] == 1 {
		teammateRank = tr[1]
	}

	switch teammateRank {
	case 2:
		return 3 // 双上 (1, 2)
	case 3:
		return 2 // 中等胜利 (1, 3)
	case 4:
		return 1 // 普通胜利 (1, 4)
	default:
		return 0
	}
}

// Score 获取积分
// 双上(1,2): 12分, 中等胜利(1,3): 6分, 普通胜利(1,4): 3分
// 如果不是赢的队伍返回 0
func (tr TeamRank) Score() int {
	switch tr.WinLevel() {
	case 3:
		return 12 // 双上
	case 2:
		return 6 // 中等胜利
	case 1:
		return 3 // 普通胜利
	default:
		return 0
	}
}

// GameRound 游戏回合信息
type GameRound struct {
	Status     GameRoundStatus
	Players    [4]Player
	Trump      Rank        // 当前头游在级牌
	Trumps     [2]Rank     // 当前两队的级牌, 只有打过A的时候才会有值
	StartedAt  int64       // 游戏开始时间（Unix时间戳，毫秒）
	FinishedAt int64       // 游戏结束时间（Unix时间戳，毫秒）
	Rounds     []GameRound // 历史回合记录, 上一局记录在0索引
}

// NewGameRound 创建一个新的游戏回合
func NewGameRound() *GameRound {
	return &GameRound{
		Status: GameStatusWaiting,
	}
}

// IsReady 检查游戏回合是否准备好开始
func (gr *GameRound) IsReady() bool {
	if gr.Status != GameStatusWaiting {
		return false
	}

	for _, player := range gr.Players {
		if player.UserId == 0 || player.Status != StatusReady {
			return false
		}
	}
	return true
}

// Start 开始游戏回合
func (gr *GameRound) Start() bool {
	if !gr.IsReady() {
		return false
	}
	gr.Status = GameStatusPlaying
	gr.StartedAt = time.Now().UnixMilli()
	for i := range gr.Players {
		gr.Players[i].Status = StatusPlaying
	}
	return true
}

// Deal 发牌
func (gr *GameRound) Deal() {
	cards := NewDeck(2)
	ccs := cards.Deal(len(gr.Players))
	for i := range gr.Players {
		gr.Players[i].SetHand(ccs[i])
	}
}

// nextRank 返回下一个可用的名次
func (gr *GameRound) nextRank() int8 {
	maxRank := int8(0)
	for _, player := range gr.Players {
		if player.Rank > maxRank {
			maxRank = player.Rank
		}
	}
	return maxRank + 1
}

// GetTeammate 获取队友的索引 (0,2一队 1,3一队)
func (gr *GameRound) GetTeammate(playerIndex int) int {
	return (playerIndex + 2) % 4
}

// GetIndex 获取玩家索引
func (gr *GameRound) GetIndex(userId int64) int {
	for i, player := range gr.Players {
		if player.UserId == userId {
			return i
		}
	}
	return -1
}

// IsTeammate 判断两个玩家是否是队友
func (gr *GameRound) IsTeammate(p1, p2 int) bool {
	return (p1+p2)%2 == 0 && p1 != p2
}

// Check 检查玩家完成状态并更新排名
// 返回是否有新的排名产生
func (gr *GameRound) Check() bool {
	if gr.Status != GameStatusPlaying {
		return false
	}

	hasNewRank := false

	// 检查每个玩家是否打完了牌
	for i := range gr.Players {
		player := &gr.Players[i]
		// 如果玩家还在游戏中且手牌为空，则完成
		if player.Status == StatusPlaying && player.HandCount() == 0 {
			player.Status = StatusFinished
			player.Rank = gr.nextRank()
			hasNewRank = true
		}
	}

	// 检查是否有一队两人都打完了
	// 队伍A: 玩家0和2, 队伍B: 玩家1和3
	teamAFinished := gr.Players[0].Status == StatusFinished && gr.Players[2].Status == StatusFinished
	teamBFinished := gr.Players[1].Status == StatusFinished && gr.Players[3].Status == StatusFinished

	if teamAFinished || teamBFinished {
		// 给未完成的玩家分配名次（末游都是4）
		for i := range gr.Players {
			player := &gr.Players[i]
			if player.Status == StatusPlaying {
				player.Status = StatusFinished
				player.Rank = 4
				hasNewRank = true
			}
		}
		gr.Status = GameStatusFinished
		gr.FinishedAt = time.Now().UnixMilli()
	}

	return hasNewRank
}

// IsFinished 检查游戏回合是否结束
func (gr *GameRound) IsFinished() bool {
	return gr.Status == GameStatusFinished
}

// GetRanks 获取所有玩家的排名
func (gr *GameRound) GetRanks() [4]int8 {
	return [4]int8{
		gr.Players[0].Rank,
		gr.Players[1].Rank,
		gr.Players[2].Rank,
		gr.Players[3].Rank,
	}
}

// GetTeamRanks 获取两队的排名情况
// 返回 [队伍A玩家0排名, 队伍A玩家2排名], [队伍B玩家1排名, 队伍B玩家3排名]
func (gr *GameRound) GetTeamRanks() [2]TeamRank {
	teamA := [2]int8{gr.Players[0].Rank, gr.Players[2].Rank}
	teamB := [2]int8{gr.Players[1].Rank, gr.Players[3].Rank}
	return [2]TeamRank{teamA, teamB}
}

// GetTeams 获取两队的玩家信息
func (gr *GameRound) GetTeams() [2]TeamPlayers {
	teamA := TeamPlayers{&gr.Players[0], &gr.Players[2]}
	teamB := TeamPlayers{&gr.Players[1], &gr.Players[3]}
	return [2]TeamPlayers{teamA, teamB}
}

// GetWinningTeam 获取获胜队伍
// 返回 0 表示队伍A(玩家0,2)获胜, 1 表示队伍B(玩家1,3)获胜, -1 表示游戏未结束
func (gr *GameRound) GetWinningTeam() int {
	if gr.Status != GameStatusFinished {
		return -1
	}

	// 头游所在队伍获胜
	for i, player := range gr.Players {
		if player.Rank == 1 {
			return i % 2 // 0,2 返回0(队伍A), 1,3 返回1(队伍B)
		}
	}
	return -1
}
