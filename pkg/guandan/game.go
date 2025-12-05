package guandan

import (
	"errors"
	"math"
	"time"
)

// 错误定义
var (
	ErrGameNotPlaying  = errors.New("game not playing")
	ErrNotYourTurn     = errors.New("not your turn")
	ErrPlayFailed      = errors.New("play failed")
	ErrPlayerNotFound  = errors.New("player not found")
	ErrGameNotFinished = errors.New("game not finished")
	ErrNoWinningTeam   = errors.New("no winning team")
)

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

// IsClimbFailed 判断翻山是否失败
// 翻山时如果排名是 1,4 则翻山失败
func (tr TeamRank) IsClimbFailed() bool {
	return tr.IsWinner() && tr.WinLevel() == 1 // 1,4 排名
}

// WinningInfo 本局获胜信息
type WinningInfo struct {
	WinningTeam  int8        // 获胜队伍 0: 队伍A(玩家0,2), 1: 队伍B(玩家1,3)
	TeamRanks    [2]TeamRank // 两队的排名情况
	WinningLevel int         // 本局获胜等级 0: 未结束或未获胜, 1: 普通胜利, 2: 中等胜利, 3: 双上
	WinningScore int32       // 本局获胜积分
	WinningCoin  int32       // 本局获胜金币
	IsClimbing   bool        // 本局是否为翻山获胜
}

// GameRound 游戏回合信息
type GameRound struct {
	Status     GameRoundStatus
	Players    [4]Player   // 玩家
	Index      int8        // 当前从哪位玩家开始出牌，0-3 对应 Players 索引
	MinType    PatternType // 用于计算翻倍的最小牌型
	MaxTrump   Rank        // 当前游戏中的最高级牌
	Trump      Rank        // 当前头游在级牌
	Trumps     [2]Rank     // 当前两队的级牌, 只有打过A的时候才会有值
	ClimCounts [2]int8     // 当前两队的翻山次数, 翻山三次后失败级牌自动变为Rank2
	StartedAt  int64       // 游戏开始时间（Unix时间戳，毫秒）
	FinishedAt int64       // 游戏结束时间（Unix时间戳，毫秒）
	Winning    WinningInfo // 本局获胜信息, 如果游戏未结束则为空
	Trick      uint8       // 当前轮次
	Tricks     []Tricks    // 每轮出过的牌型记录
	Rounds     []GameRound // 历史回合记录, 上一局记录在0索引
}

// NewGameRound 创建一个新的游戏回合
func NewGameRound(maxTrump Rank) *GameRound {
	return &GameRound{
		Status:   GameStatusWaiting,
		MaxTrump: maxTrump,
	}
}

// IsClimbing 是否在翻山
func (gr *GameRound) IsClimbing() bool {
	return gr.Trump == gr.MaxTrump && gr.MaxTrump != RankNone
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

// Play 玩家出牌
// userId: 出牌玩家的用户ID
// pattern: 出牌的牌型，如果Type为PatternTypeNone表示过牌（不出）
// 返回值：error 错误信息
func (gr *GameRound) Play(userId int64, pattern Pattern) error {
	// 检查游戏状态
	if gr.Status != GameStatusPlaying {
		return ErrGameNotPlaying
	}

	// 获取当前应该出牌的玩家
	currentPlayer := &gr.Players[gr.Index]

	// 检查是否轮到该玩家出牌
	if currentPlayer.UserId != userId {
		return ErrNotYourTurn
	}

	// 玩家出牌（包括过牌，Type为None也会记录）
	pattern.PlayerId = gr.Index
	if !currentPlayer.Play(pattern) {
		return ErrPlayFailed
	}

	// 确保当前轮次的Tricks数组存在
	for len(gr.Tricks) <= int(gr.Trick) {
		gr.Tricks = append(gr.Tricks, Tricks{})
	}
	// 记录到Tricks
	playedIndex := uint8(len(currentPlayer.Played) - 1)
	trickRecord := Trick{
		PlayerIndex:  uint8(gr.Index),
		PatternType:  pattern.Type,
		PatternIndex: playedIndex,
	}
	gr.Tricks[gr.Trick] = append(gr.Tricks[gr.Trick], trickRecord)

	return nil
}

// ActivePlayerCount 获取还在游戏且还有手牌的玩家数量
func (gr *GameRound) ActivePlayerCount() int {
	count := 0
	for _, player := range gr.Players {
		if player.Status == StatusPlaying && player.HandCount() > 0 {
			count++
		}
	}
	return count
}

// IsTrickFinished 判断当前轮次是否完成
// 完成条件：从最后一个出实牌的人开始，后续连续Pass的人数达到了限制
// 限制取决于出牌者是否还有牌：
// 1. 如果出牌者还有牌：需要 (ActivePlayerCount - 1) 个Pass
// 2. 如果出牌者没牌了：需要 ActivePlayerCount 个Pass
func (gr *GameRound) IsTrickFinished() bool {
	if int(gr.Trick) >= len(gr.Tricks) {
		return false
	}

	tricks := gr.Tricks[gr.Trick]
	if len(tricks) == 0 {
		return false
	}

	// 1. 找到最后一个出实牌的索引
	lastRealIndex := -1
	for i := len(tricks) - 1; i >= 0; i-- {
		if !tricks[i].IsPass() {
			lastRealIndex = i
			break
		}
	}

	if lastRealIndex == -1 {
		return false
	}

	// 2. 统计后面的连续Pass数量
	passCount := len(tricks) - 1 - lastRealIndex

	// 3. 获取活跃玩家数量（还在玩且有牌）
	activeCount := gr.ActivePlayerCount()

	// 4. 检查出最大牌的玩家状态
	lastPlayerIndex := tricks[lastRealIndex].PlayerIndex
	lastPlayer := &gr.Players[lastPlayerIndex]

	// 判断出牌者是否还是活跃玩家
	lastPlayerIsActive := lastPlayer.Status == StatusPlaying && lastPlayer.HandCount() > 0
	if lastPlayerIsActive {
		activeCount--
	}

	return passCount >= activeCount
}

// GetTrickWinner 获取当前轮次的赢家索引
// 返回值：赢家玩家索引，如果本轮未完成返回 -1
// Check 完成后调用此方法以获取赢家
func (gr *GameRound) GetTrickWinner() int8 {
	if !gr.IsTrickFinished() {
		return -1
	}

	tricks := gr.Tricks[gr.Trick]

	// 找到最后一个非Pass的出牌记录，那就是赢家
	for i := len(tricks) - 1; i >= 0; i-- {
		if !tricks[i].IsPass() {
			winnerIndex := int8(tricks[i].PlayerIndex)

			// 检查赢家是否已经打完牌
			winner := &gr.Players[winnerIndex]
			if winner.Status != StatusPlaying {
				// 赢家已经打完牌，把出牌权交给队友
				teammateIndex := gr.GetTeammate(int(winnerIndex))
				if gr.Players[teammateIndex].Status == StatusPlaying {
					return int8(teammateIndex)
				}
			}
			return winnerIndex
		}
	}

	return -1
}

// FinishTrick 完成当前轮次，开始新的一轮
// 返回值：是否成功完成轮次
func (gr *GameRound) FinishTrick() bool {
	winnerIndex := gr.GetTrickWinner()
	if winnerIndex < 0 {
		return false
	}

	gr.Trick++
	gr.Index = winnerIndex
	return true
}

func (gr *GameRound) NextPlayer() {
	// 循环找到下一个还在游戏中的玩家
	for range 4 {
		gr.Index = (gr.Index + 1) % 4
		if gr.Players[gr.Index].Status == StatusPlaying {
			return
		}
	}
}

// NewTrick 开始新的一轮（当一轮结束时调用）
// winnerIndex: 赢得这一轮的玩家索引
func (gr *GameRound) NewTrick(winnerIndex int8) {
	gr.Trick++
	gr.Index = winnerIndex
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

// GetWinningIndex 获取头游的玩家索引
func (gr *GameRound) GetWinningIndex() int8 {
	for i, player := range gr.Players {
		if player.Rank == 1 {
			return int8(i)
		}
	}
	return 0 // 默认玩家0
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

// NextIndex 设置下一个出牌玩家索引
func (gr *GameRound) NextIndex() {
	gr.Index = (gr.Index + 1) % int8(len(gr.Players))
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
		// 给未完成的玩家分配名次
		for i := range gr.Players {
			player := &gr.Players[i]
			if player.Status == StatusPlaying {
				player.Status = StatusFinished
				player.Rank = gr.nextRank()
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
func (gr *GameRound) GetWinningTeam() int8 {
	if gr.Status != GameStatusFinished {
		return -1
	}

	// 头游所在队伍获胜
	for i, player := range gr.Players {
		if player.Rank == 1 {
			return int8(i) % 2 // 0,2 返回0(队伍A), 1,3 返回1(队伍B)
		}
	}
	return -1
}

// CountDouble 计算符合翻倍条件的牌型数量
func (gr *GameRound) CountDouble() (count int32) {
	for _, player := range gr.Players {
		for _, pattern := range player.Played {
			if (pattern.Type == PatternTypeBomb && pattern.Length >= 6) || pattern.Type > PatternTypeBomb {
				count++
			}
		}
	}
	return count
}

// CalcMultiplier 计算翻倍倍数
// minType: 最小牌型，统计 >= minType 的数量
// 返回 2^N，N 为符合条件的牌型数量
func (gr *GameRound) CalcMultiplier() int32 {
	count := gr.CountDouble()
	if count == 0 {
		return 1
	}
	// 2^N
	return int32(math.Pow(2, float64(count)))
}

// Settle 结算函数
// basePoint: 基础积分
// baseCoin: 基础金币
func (gr *GameRound) Settle(basePoint, baseCoin int32) error {
	if gr.Status != GameStatusFinished {
		return ErrGameNotFinished
	}

	// 获取获胜队伍
	winningTeam := gr.GetWinningTeam()
	if winningTeam < 0 {
		return ErrNoWinningTeam
	}

	// 获取队伍排名
	teamRanks := gr.GetTeamRanks()

	// 计算翻倍
	multiplier := int32(gr.CalcMultiplier())

	// 获取获胜队伍的积分倍率
	winTeamRank := teamRanks[winningTeam]
	scoreMultiplier := int32(winTeamRank.Score()) // 12, 6, 或 3

	// 计算最终积分和金币变化
	pointChange := basePoint * scoreMultiplier * multiplier / 3 // 除以3是因为 Score 返回的是 3, 6, 12
	coinChange := baseCoin * scoreMultiplier * multiplier / 3

	// 更新 WinningInfo
	gr.Winning.WinningTeam = winningTeam
	gr.Winning.TeamRanks = teamRanks
	gr.Winning.WinningLevel = winTeamRank.WinLevel()
	gr.Winning.WinningScore = pointChange
	gr.Winning.WinningCoin = coinChange
	// 记录是否翻山成功（在翻山状态且不是1,4排名）
	gr.Winning.IsClimbing = gr.IsClimbing() && !winTeamRank.IsClimbFailed()

	// 更新玩家信息
	for i := range gr.Players {
		player := &gr.Players[i]
		playerTeam := int8(i % 2) // 0,2 属于队伍0，1,3 属于队伍1

		if playerTeam == winningTeam {
			// 赢家
			player.IsWinner = true
			player.PointChange = pointChange
			player.CoinChange = coinChange
		} else {
			// 输家
			player.IsWinner = false
			player.PointChange = -pointChange
			player.CoinChange = -coinChange
		}
	}
	return nil
}

// NextRound 准备下一回合
// 将当前回合保存到历史记录，重置玩家状态，更新级牌
func (gr *GameRound) NextRound() error {
	if gr.Status != GameStatusFinished {
		return ErrGameNotFinished
	}

	// 获取获胜队伍
	winningTeam := gr.GetWinningTeam()
	if winningTeam < 0 {
		return ErrNoWinningTeam
	}

	// 设置下一局由头游出牌
	gr.Index = gr.GetWinningIndex()

	// 保存当前回合到历史记录（插入到最前面）
	currentRound := *gr
	currentRound.Rounds = nil // 避免嵌套保存历史记录
	gr.Rounds = append([]GameRound{currentRound}, gr.Rounds...)

	// 根据获胜等级计算升级数
	winTeamRank := gr.Winning.TeamRanks[winningTeam]
	levelUp := Rank(winTeamRank.WinLevel()) // 双上升3级，中等升2级，普通升1级

	// 检查翻山情况
	isClimbing := gr.IsClimbing()
	climbFailed := isClimbing && winTeamRank.IsClimbFailed()

	if climbFailed {
		// 翻山失败，增加失败次数
		gr.ClimCounts[winningTeam]++

		// 如果翻山失败次数 >= 3，级牌重置为 Rank2
		if gr.ClimCounts[winningTeam] >= 3 {
			gr.Trumps[winningTeam] = Rank2
			gr.ClimCounts[winningTeam] = 0 // 重置失败次数
		}
		// 翻山失败不升级，级牌保持不变
	} else {
		// 翻山成功或非翻山情况，正常升级
		if isClimbing {
			// 翻山成功，重置所有状态（相当于重新开始）
			gr.Trumps = [2]Rank{Rank2, Rank2}
			gr.ClimCounts = [2]int8{0, 0}
			gr.Trump = Rank2
		} else {
			// 非翻山情况，正常升级
			// 更新获胜队伍的级牌
			if gr.MaxTrump != RankNone {
				gr.Trumps[winningTeam] += levelUp
				if gr.Trumps[winningTeam] > gr.MaxTrump {
					gr.Trumps[winningTeam] = gr.MaxTrump // 最高为MaxTrump
				}
			}
		}
	}

	// 更新当前级牌为获胜队伍的级牌（翻山成功时已经设置为Rank2）
	if !isClimbing || climbFailed {
		if gr.MaxTrump != RankNone {
			gr.Trump = gr.Trumps[winningTeam]
		} else {
			gr.Trump = Rank2 // 如果没有最高级牌，则直接设为2
		}
	}

	// 重置游戏状态
	gr.Status = GameStatusWaiting
	gr.StartedAt = 0
	gr.FinishedAt = 0
	gr.Winning = WinningInfo{}

	// 重置玩家状态
	for i := range gr.Players {
		player := &gr.Players[i]
		player.Status = StatusWaiting
		player.Hand = nil
		player.Played = nil
		player.Rank = 0
		player.IsWinner = false
		player.PointChange = 0
		player.CoinChange = 0
	}
	return nil
}

// IsAllFinished 检查所有玩家是否都已完成
func (gr *GameRound) IsAllFinished() bool {
	if !gr.IsFinished() {
		return false
	}
	// 如果是升级赛，检查是否为翻山获胜
	if gr.MaxTrump != RankNone {
		return gr.Winning.IsClimbing
	}
	return false
}
