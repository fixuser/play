package guandan

import (
	"math"
	"time"
)

// WinningInfo 本局获胜信息
type WinningInfo struct {
	WinningTeam   int8        // 获胜队伍 0: 队伍A(玩家0,2), 1: 队伍B(玩家1,3)
	TeamRanks     [2]TeamRank // 两队的排名情况
	WinningLevel  int         // 本局获胜等级 0: 未结束或未获胜, 1: 普通胜利, 2: 中等胜利, 3: 双上
	WinningScore  int32       // 本局获胜积分
	WinningCoin   int32       // 本局获胜金币
	IsClimbingWin bool        // 本局是否为翻山获胜, 游戏结束
}

type GameOptions struct {
	MaxTrump     Rank          // 最大级牌, 过A，升级
	PatternLevel int           // 用于计算翻倍的最小牌型
	IsRotate     bool          // 是否换人
	MaxCount     int           // 最大局数
	PlayTime     time.Duration // 出牌超时时间, 0不超时
	IsClimbing   bool          // 是否翻山
}

type Option func(*GameOptions)

func WithMaxTrump(rank Rank) Option {
	return func(o *GameOptions) {
		o.MaxTrump = rank
	}
}

func WithPatternLevel(level int) Option {
	return func(o *GameOptions) {
		o.PatternLevel = level
	}
}

func WithIsRotate(isRotate bool) Option {
	return func(o *GameOptions) {
		o.IsRotate = isRotate
	}
}

func WithMaxCount(count int) Option {
	return func(o *GameOptions) {
		o.MaxCount = count
	}
}

func WithPlayTime(playTime time.Duration) Option {
	return func(o *GameOptions) {
		o.PlayTime = playTime
	}
}

// GameRound 游戏回合信息
type GameRound struct {
	Options        GameOptions     // 游戏选项
	Status         GameRoundStatus // 游戏状态
	Players        [4]Player       // 玩家
	Index          int8            // 当前从哪位玩家开始出牌，0-3 对应 Players 索引
	Trump          Rank            // 当前头游在级牌
	TrumpTeamIndex int8            // 头游所在的队伍
	Trumps         [2]Rank         // 当前两队的级牌, 只有打过A的时候才会有值
	ClimCounts     [2]int8         // 当前两队的翻山次数, 翻山三次后失败级牌自动变为Rank2
	MaxTrumpCounts [2]int8         // 当前两队打max trump的次数
	StartedAt      int64           // 游戏开始时间（Unix时间戳，毫秒）
	FinishedAt     int64           // 游戏结束时间（Unix时间戳，毫秒）
	Winning        WinningInfo     // 本局获胜信息, 如果游戏未结束则为空
	Trick          uint8           // 当前轮次
	Tricks         []Tricks        // 每轮出过的牌型记录
	Rounds         []GameRound     // 历史回合记录, 上一局记录在0索引
}

// NewGameRound 创建一个新的游戏回合
func NewGameRound(opts ...Option) *GameRound {
	options := GameOptions{
		MaxTrump:     RankA, // 默认打到A
		PatternLevel: 0,     // 默认不限制
	}

	for _, opt := range opts {
		opt(&options)
	}

	return &GameRound{
		Status:  GameStatusWaiting,
		Options: options,
	}
}

// GetTeamTrump 获取指定队伍的级牌
func (gr *GameRound) GetTeamTrump(team int8) Rank {
	return gr.Trumps[team]
}

// IsClimbing 是否在翻山
func (gr *GameRound) IsClimbing() bool {
	if !gr.Options.IsClimbing {
		return false
	}
	if gr.Options.MaxTrump == RankNone {
		return false
	}

	// 当前
	if gr.Trumps[gr.TrumpTeamIndex] == gr.Options.MaxTrump {
		if gr.MaxTrumpCounts[gr.TrumpTeamIndex] > 0 {
			return true
		}
	}
	return false
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

// CountPatternLevel 计算符合翻倍条件的牌型数量
func (gr *GameRound) CountPatternLevel(patternLevel int) (count int32) {
	if patternLevel <= 0 {
		return 0
	}

	for _, player := range gr.Players {
		for _, pattern := range player.Played {
			if pattern.GetLevel() >= patternLevel {
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
	count := gr.CountPatternLevel(gr.Options.PatternLevel)
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
	pointChange := basePoint * scoreMultiplier * multiplier
	coinChange := baseCoin * scoreMultiplier * multiplier

	// 更新 WinningInfo
	gr.Winning.WinningTeam = winningTeam
	gr.Winning.TeamRanks = teamRanks
	gr.Winning.WinningLevel = winTeamRank.WinLevel()
	gr.Winning.WinningScore = pointChange
	gr.Winning.WinningCoin = coinChange
	// 记录是否翻山成功（在翻山状态且不是1,4排名）
	if winningTeam == gr.TrumpTeamIndex {
		gr.Winning.IsClimbingWin = gr.IsClimbing() && !winTeamRank.IsClimbFailed()
	}
	// 记录打最大级牌的次数
	if gr.Options.MaxTrump != RankNone && gr.Options.MaxTrump == gr.Trumps[gr.TrumpTeamIndex] {
		gr.MaxTrumpCounts[gr.TrumpTeamIndex]++
	}

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
	gr.TrumpTeamIndex = winningTeam

	// 保存当前回合到历史记录（插入到最前面）
	currentRound := *gr
	currentRound.Rounds = nil // 避免嵌套保存历史记录
	gr.Rounds = append([]GameRound{currentRound}, gr.Rounds...)

	// 根据获胜等级计算升级数
	winTeamRank := gr.Winning.TeamRanks[winningTeam]
	levelUp := Rank(winTeamRank.WinLevel()) // 双上升3级，中等升2级，普通升1级

	// 检查翻山情况
	isClimbing := gr.IsClimbing()
	climbFailed := (isClimbing && winTeamRank.IsClimbFailed()) || (winningTeam != gr.TrumpTeamIndex)

	if climbFailed {
		// 翻山失败，增加失败次数
		gr.ClimCounts[gr.TrumpTeamIndex]++

		// 如果翻山失败次数 >= 3，级牌重置为 Rank2
		if gr.ClimCounts[gr.TrumpTeamIndex] >= 3 {
			gr.Trumps[gr.TrumpTeamIndex] = Rank2
			gr.ClimCounts[gr.TrumpTeamIndex] = 0 // 重置失败次数
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
			if gr.Options.MaxTrump != RankNone {
				gr.Trumps[winningTeam] += levelUp
				if gr.Trumps[winningTeam] > gr.Options.MaxTrump {
					gr.Trumps[winningTeam] = gr.Options.MaxTrump // 最高为MaxTrump
				}
			}
		}
	}

	// 更新当前级牌为获胜队伍的级牌（翻山成功时已经设置为Rank2）
	if !isClimbing || climbFailed {
		if gr.Options.MaxTrump != RankNone {
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

	// 检查是否需要换人
	if gr.Options.IsRotate {
		gr.RotatePlayers()
	}

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
	if gr.Options.MaxTrump != RankNone {
		if gr.Options.IsClimbing {
			return gr.Winning.IsClimbingWin
		}
		return gr.Trumps[0] == gr.Options.MaxTrump || gr.Trumps[1] == gr.Options.MaxTrump
	}
	return false
}

// RotatePlayers 轮换玩家位置 (0->1->2->0)
func (gr *GameRound) RotatePlayers() {
	// 暂存玩家0
	p0 := gr.Players[0]
	// 移位
	gr.Players[0] = gr.Players[2]
	gr.Players[2] = gr.Players[1]
	gr.Players[1] = p0

	// 此时:
	// Pos 0 becomes old Pos 2
	// Pos 2 becomes old Pos 1
	// Pos 1 becomes old Pos 0
	// 这样实现了 0->1->2->0 的轮换?
	// Wait, requirements said "0-2 Players换一个位置"
	// User said: "需要在0-2的Players换一个位置，这样子才能队友变化"
	// Players: 0, 1, 2, 3. Teams: (0,2), (1,3).
	// If we rotate 0, 1, 2:
	// New 0 = Old 2 (Team A)
	// New 1 = Old 0 (Team A) -> Was Team A, now Team B pos 1
	// New 2 = Old 1 (Team B) -> Was Team B, now Team A pos 2
	// New 3 = Old 3 (Team B)
	// Teams become:
	// Team A (0,2): Old 2 + Old 1. (Mixed!)
	// Team B (1,3): Old 0 + Old 3. (Mixed!)
	// This changes teammates! Perfect.
}
