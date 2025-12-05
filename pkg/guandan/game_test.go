package guandan

import (
	"testing"
)

func TestNewGameRound(t *testing.T) {
	gr := NewGameRound(RankA)
	if gr == nil {
		t.Fatal("NewGameRound returned nil")
	}
	if gr.Status != GameStatusWaiting {
		t.Errorf("expected status %v, got %v", GameStatusWaiting, gr.Status)
	}
	if gr.MaxTrump != RankA {
		t.Errorf("expected MaxTrump %v, got %v", RankA, gr.MaxTrump)
	}
}

func TestGameRound_IsReady(t *testing.T) {
	gr := NewGameRound(RankA)

	// 没有玩家时不应该准备好
	if gr.IsReady() {
		t.Error("should not be ready without players")
	}

	// 设置玩家但未准备
	for i := range gr.Players {
		gr.Players[i].UserId = int64(i + 1)
		gr.Players[i].Status = StatusWaiting
	}
	if gr.IsReady() {
		t.Error("should not be ready when players are not ready")
	}

	// 所有玩家准备好
	for i := range gr.Players {
		gr.Players[i].Status = StatusReady
	}
	if !gr.IsReady() {
		t.Error("should be ready when all players are ready")
	}

	// 游戏已开始不应该准备
	gr.Status = GameStatusPlaying
	if gr.IsReady() {
		t.Error("should not be ready when game is playing")
	}
}

func TestGameRound_Start(t *testing.T) {
	gr := NewGameRound(RankA)

	// 未准备时不能开始
	if gr.Start() {
		t.Error("should not start when not ready")
	}

	// 设置玩家并准备
	for i := range gr.Players {
		gr.Players[i].UserId = int64(i + 1)
		gr.Players[i].Status = StatusReady
	}

	// 开始游戏
	if !gr.Start() {
		t.Error("should start when ready")
	}

	if gr.Status != GameStatusPlaying {
		t.Errorf("expected status %v, got %v", GameStatusPlaying, gr.Status)
	}

	if gr.StartedAt == 0 {
		t.Error("StartedAt should be set")
	}

	for i, player := range gr.Players {
		if player.Status != StatusPlaying {
			t.Errorf("player %d should be playing", i)
		}
	}
}

func TestGameRound_Deal(t *testing.T) {
	gr := NewGameRound(RankA)
	gr.Deal()

	// 检查每个玩家都有牌
	for i, player := range gr.Players {
		if len(player.Hand) == 0 {
			t.Errorf("player %d should have cards", i)
		}
		// 2副牌108张，4个玩家，每人27张
		if len(player.Hand) != 27 {
			t.Errorf("player %d expected 27 cards, got %d", i, len(player.Hand))
		}
	}
}

func TestGameRound_GetTeammate(t *testing.T) {
	gr := NewGameRound(RankA)

	tests := []struct {
		playerIndex int
		expected    int
	}{
		{0, 2},
		{1, 3},
		{2, 0},
		{3, 1},
	}

	for _, tt := range tests {
		got := gr.GetTeammate(tt.playerIndex)
		if got != tt.expected {
			t.Errorf("GetTeammate(%d) = %d, want %d", tt.playerIndex, got, tt.expected)
		}
	}
}

func TestGameRound_IsTeammate(t *testing.T) {
	gr := NewGameRound(RankA)

	// 0和2是队友，1和3是队友
	if !gr.IsTeammate(0, 2) {
		t.Error("0 and 2 should be teammates")
	}
	if !gr.IsTeammate(1, 3) {
		t.Error("1 and 3 should be teammates")
	}
	if gr.IsTeammate(0, 1) {
		t.Error("0 and 1 should not be teammates")
	}
	if gr.IsTeammate(0, 0) {
		t.Error("0 and 0 should not be teammates (same player)")
	}
}

func TestGameRound_Check_SingleFinish(t *testing.T) {
	gr := NewGameRound(RankA)

	// 设置玩家
	for i := range gr.Players {
		gr.Players[i].UserId = int64(i + 1)
		gr.Players[i].Status = StatusReady
	}
	gr.Start()
	gr.Deal()

	// 玩家0打完所有牌
	gr.Players[0].Hand = nil

	hasNewRank := gr.Check()
	if !hasNewRank {
		t.Error("should have new rank")
	}
	if gr.Players[0].Rank != 1 {
		t.Errorf("player 0 should be rank 1, got %d", gr.Players[0].Rank)
	}
	if gr.Players[0].Status != StatusFinished {
		t.Error("player 0 should be finished")
	}

	// 游戏应该继续，因为一队还没都完成
	if gr.Status != GameStatusPlaying {
		t.Error("game should still be playing")
	}
}

func TestGameRound_Check_TeamFinish(t *testing.T) {
	gr := NewGameRound(RankA)

	// 设置玩家
	for i := range gr.Players {
		gr.Players[i].UserId = int64(i + 1)
		gr.Players[i].Status = StatusReady
	}
	gr.Start()
	gr.Deal()

	// 队伍A（玩家0和2）都打完
	gr.Players[0].Hand = nil
	gr.Check() // 玩家0获得第1名

	gr.Players[2].Hand = nil
	gr.Check() // 玩家2获得第2名，游戏结束

	if gr.Status != GameStatusFinished {
		t.Error("game should be finished when one team completes")
	}

	// 检查排名
	if gr.Players[0].Rank != 1 {
		t.Errorf("player 0 should be rank 1, got %d", gr.Players[0].Rank)
	}
	if gr.Players[2].Rank != 2 {
		t.Errorf("player 2 should be rank 2, got %d", gr.Players[2].Rank)
	}
	// 剩余玩家应该都是末名
	if gr.Players[1].Rank != 3 && gr.Players[1].Rank != 4 {
		t.Errorf("player 1 should be rank 3 or 4, got %d", gr.Players[1].Rank)
	}
	if gr.Players[3].Rank != 3 && gr.Players[3].Rank != 4 {
		t.Errorf("player 3 should be rank 3 or 4, got %d", gr.Players[3].Rank)
	}
}

func TestGameRound_GetWinningTeam(t *testing.T) {
	gr := NewGameRound(RankA)

	// 游戏未结束
	if gr.GetWinningTeam() != -1 {
		t.Error("should return -1 when game not finished")
	}

	// 设置游戏结束，玩家0是头游
	gr.Status = GameStatusFinished
	gr.Players[0].Rank = 1
	gr.Players[2].Rank = 2
	gr.Players[1].Rank = 3
	gr.Players[3].Rank = 4

	if gr.GetWinningTeam() != 0 {
		t.Errorf("team A (0,2) should win, got team %d", gr.GetWinningTeam())
	}

	// 玩家1是头游
	gr.Players[0].Rank = 3
	gr.Players[1].Rank = 1

	if gr.GetWinningTeam() != 1 {
		t.Errorf("team B (1,3) should win, got team %d", gr.GetWinningTeam())
	}
}

func TestTeamRank_IsWinner(t *testing.T) {
	tests := []struct {
		rank     TeamRank
		expected bool
	}{
		{TeamRank{1, 2}, true},  // 双上
		{TeamRank{1, 3}, true},  // 中等胜利
		{TeamRank{1, 4}, true},  // 普通胜利
		{TeamRank{2, 1}, true},  // 队友是头游
		{TeamRank{3, 4}, false}, // 没有头游
		{TeamRank{2, 3}, false}, // 没有头游
	}

	for _, tt := range tests {
		got := tt.rank.IsWinner()
		if got != tt.expected {
			t.Errorf("TeamRank%v.IsWinner() = %v, want %v", tt.rank, got, tt.expected)
		}
	}
}

func TestTeamRank_WinLevel(t *testing.T) {
	tests := []struct {
		rank     TeamRank
		expected int
	}{
		{TeamRank{1, 2}, 3}, // 双上
		{TeamRank{2, 1}, 3}, // 双上（队友头游）
		{TeamRank{1, 3}, 2}, // 中等胜利
		{TeamRank{3, 1}, 2}, // 中等胜利
		{TeamRank{1, 4}, 1}, // 普通胜利
		{TeamRank{4, 1}, 1}, // 普通胜利
		{TeamRank{3, 4}, 0}, // 不是赢家
	}

	for _, tt := range tests {
		got := tt.rank.WinLevel()
		if got != tt.expected {
			t.Errorf("TeamRank%v.WinLevel() = %v, want %v", tt.rank, got, tt.expected)
		}
	}
}

func TestTeamRank_Score(t *testing.T) {
	tests := []struct {
		rank     TeamRank
		expected int
	}{
		{TeamRank{1, 2}, 12}, // 双上
		{TeamRank{1, 3}, 6},  // 中等胜利
		{TeamRank{1, 4}, 3},  // 普通胜利
		{TeamRank{3, 4}, 0},  // 不是赢家
	}

	for _, tt := range tests {
		got := tt.rank.Score()
		if got != tt.expected {
			t.Errorf("TeamRank%v.Score() = %v, want %v", tt.rank, got, tt.expected)
		}
	}
}

func TestGameRound_CountDouble(t *testing.T) {
	gr := NewGameRound(RankA)

	// 玩家0打出一个6张炸弹（>=6张炸弹计入翻倍）
	gr.Players[0].Played = Patterns{
		{Type: PatternTypeBomb, Length: 6, Cards: Cards{{Rank: Rank5, Suit: SuitSpader}, {Rank: Rank5, Suit: SuitHeart}, {Rank: Rank5, Suit: SuitClub}, {Rank: Rank5, Suit: SuitDiamond}, {Rank: Rank5, Suit: SuitSpader}, {Rank: Rank5, Suit: SuitHeart}}},
		{Type: PatternTypeSingle, Cards: Cards{{Rank: Rank3, Suit: SuitSpader}}},
	}

	// 玩家1打出一个四大天王
	gr.Players[1].Played = Patterns{
		{Type: PatternTypeFourJokers, Cards: Cards{}},
	}

	// 统计翻倍牌型（6张及以上炸弹 或 四大天王等）
	count := gr.CountDouble()
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
}

func TestGameRound_CalcMultiplier(t *testing.T) {
	gr := NewGameRound(RankA)

	// 没有翻倍牌型，倍数为1
	multiplier := gr.CalcMultiplier()
	if multiplier != 1 {
		t.Errorf("expected multiplier 1, got %d", multiplier)
	}

	// 添加1个6张炸弹
	gr.Players[0].Played = Patterns{
		{Type: PatternTypeBomb, Length: 6, Cards: Cards{{Rank: Rank5, Suit: SuitSpader}}},
	}
	multiplier = gr.CalcMultiplier()
	if multiplier != 2 {
		t.Errorf("expected multiplier 2, got %d", multiplier)
	}

	// 添加1个四大天王
	gr.Players[1].Played = Patterns{
		{Type: PatternTypeFourJokers, Cards: Cards{}},
	}
	multiplier = gr.CalcMultiplier()
	if multiplier != 4 {
		t.Errorf("expected multiplier 4, got %d", multiplier)
	}

	// 添加1个7张炸弹
	gr.Players[2].Played = Patterns{
		{Type: PatternTypeBomb, Length: 7, Cards: Cards{}},
	}
	multiplier = gr.CalcMultiplier()
	if multiplier != 8 {
		t.Errorf("expected multiplier 8, got %d", multiplier)
	}
}

func TestGameRound_Settle(t *testing.T) {
	gr := NewGameRound(RankA)

	// 游戏未结束
	err := gr.Settle(10, 100)
	if err != ErrGameNotFinished {
		t.Errorf("expected ErrGameNotFinished, got %v", err)
	}

	// 设置游戏结束
	gr.Status = GameStatusFinished
	gr.Players[0].Rank = 1 // 队伍A头游
	gr.Players[2].Rank = 2 // 队伍A二游（双上）
	gr.Players[1].Rank = 4
	gr.Players[3].Rank = 4

	err = gr.Settle(10, 100)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 检查 WinningInfo
	if gr.Winning.WinningTeam != 0 {
		t.Errorf("expected winning team 0, got %d", gr.Winning.WinningTeam)
	}
	if gr.Winning.WinningLevel != 3 {
		t.Errorf("expected winning level 3 (双上), got %d", gr.Winning.WinningLevel)
	}

	// 双上积分倍率为4 (12/3)，基础积分10，所以积分变化为40
	expectedPoint := int32(10 * 12 / 3) // 40
	if gr.Winning.WinningScore != expectedPoint {
		t.Errorf("expected winning score %d, got %d", expectedPoint, gr.Winning.WinningScore)
	}

	// 检查玩家信息
	if !gr.Players[0].IsWinner || !gr.Players[2].IsWinner {
		t.Error("team A players should be winners")
	}
	if gr.Players[1].IsWinner || gr.Players[3].IsWinner {
		t.Error("team B players should not be winners")
	}

	// 赢家积分增加，输家积分减少
	if gr.Players[0].PointChange != expectedPoint {
		t.Errorf("winner point change should be %d, got %d", expectedPoint, gr.Players[0].PointChange)
	}
	if gr.Players[1].PointChange != -expectedPoint {
		t.Errorf("loser point change should be %d, got %d", -expectedPoint, gr.Players[1].PointChange)
	}
}

func TestGameRound_NextRound(t *testing.T) {
	gr := NewGameRound(RankA)

	// 设置初始级牌
	gr.Trumps[0] = Rank2
	gr.Trumps[1] = Rank2

	// 设置玩家
	for i := range gr.Players {
		gr.Players[i].UserId = int64(i + 1)
		gr.Players[i].Status = StatusReady
	}
	gr.Start()
	gr.Deal()

	// 模拟游戏结束（队伍A双上）
	gr.Status = GameStatusFinished
	gr.Players[0].Rank = 1
	gr.Players[2].Rank = 2
	gr.Players[1].Rank = 4
	gr.Players[3].Rank = 4
	gr.Players[0].Status = StatusFinished
	gr.Players[2].Status = StatusFinished
	gr.Players[1].Status = StatusFinished
	gr.Players[3].Status = StatusFinished

	// 先结算
	err := gr.Settle(10, 100)
	if err != nil {
		t.Fatalf("Settle failed: %v", err)
	}

	// 下一回合
	err = gr.NextRound()
	if err != nil {
		t.Fatalf("NextRound failed: %v", err)
	}

	// 检查历史记录
	if len(gr.Rounds) != 1 {
		t.Errorf("expected 1 round in history, got %d", len(gr.Rounds))
	}

	// 检查级牌升级（双上升3级）
	if gr.Trumps[0] != Rank5 { // 2 + 3 = 5
		t.Errorf("team A trump should be Rank5, got %d", gr.Trumps[0])
	}

	// 检查当前级牌
	if gr.Trump != Rank5 {
		t.Errorf("current trump should be Rank5, got %d", gr.Trump)
	}

	// 检查状态重置
	if gr.Status != GameStatusWaiting {
		t.Errorf("status should be waiting, got %d", gr.Status)
	}

	for i, player := range gr.Players {
		if player.Status != StatusWaiting {
			t.Errorf("player %d status should be waiting", i)
		}
		if player.Hand != nil {
			t.Errorf("player %d hand should be nil", i)
		}
		if player.Rank != 0 {
			t.Errorf("player %d rank should be 0", i)
		}
	}
}

func TestGameRound_GetTeamRanks(t *testing.T) {
	gr := NewGameRound(RankA)
	gr.Players[0].Rank = 1
	gr.Players[1].Rank = 3
	gr.Players[2].Rank = 2
	gr.Players[3].Rank = 4

	teamRanks := gr.GetTeamRanks()

	// 队伍A: 玩家0和2
	if teamRanks[0] != (TeamRank{1, 2}) {
		t.Errorf("team A ranks should be [1, 2], got %v", teamRanks[0])
	}

	// 队伍B: 玩家1和3
	if teamRanks[1] != (TeamRank{3, 4}) {
		t.Errorf("team B ranks should be [3, 4], got %v", teamRanks[1])
	}
}

func TestGameRound_FullGame(t *testing.T) {
	gr := NewGameRound(RankA)

	// 初始化级牌
	gr.Trumps[0] = Rank2
	gr.Trumps[1] = Rank2
	gr.Trump = Rank2

	// 设置玩家
	for i := range gr.Players {
		gr.Players[i].UserId = int64(i + 1)
		gr.Players[i].Status = StatusReady
	}

	// 开始游戏
	if !gr.Start() {
		t.Fatal("failed to start game")
	}

	// 发牌
	gr.Deal()

	// 模拟玩家依次打完牌
	// 玩家0打完
	gr.Players[0].Hand = nil
	gr.Check()
	if gr.Players[0].Rank != 1 {
		t.Errorf("player 0 should be rank 1")
	}

	// 玩家1打完
	gr.Players[1].Hand = nil
	gr.Check()
	if gr.Players[1].Rank != 2 {
		t.Errorf("player 1 should be rank 2")
	}

	// 玩家2打完（此时队伍A完成）
	gr.Players[2].Hand = nil
	gr.Check()

	// 游戏应该结束
	if !gr.IsFinished() {
		t.Error("game should be finished")
	}

	// 结算
	err := gr.Settle(10, 100)
	if err != nil {
		t.Fatalf("Settle failed: %v", err)
	}

	// 队伍A获胜（1,3名次）
	if gr.GetWinningTeam() != 0 {
		t.Errorf("team A should win")
	}
	if gr.Winning.WinningLevel != 2 { // 1,3是中等胜利
		t.Errorf("winning level should be 2, got %d", gr.Winning.WinningLevel)
	}

	// 下一回合
	err = gr.NextRound()
	if err != nil {
		t.Fatalf("NextRound failed: %v", err)
	}

	// 级牌应该升2级
	if gr.Trumps[0] != Rank4 { // 2 + 2 = 4
		t.Errorf("team A trump should be Rank4, got %d", gr.Trumps[0])
	}
}

func TestTeamRank_IsClimbFailed(t *testing.T) {
	tests := []struct {
		rank     TeamRank
		expected bool
	}{
		{TeamRank{1, 2}, false}, // 双上，翻山成功
		{TeamRank{1, 3}, false}, // 中等胜利，翻山成功
		{TeamRank{1, 4}, true},  // 普通胜利，翻山失败
		{TeamRank{4, 1}, true},  // 普通胜利，翻山失败
		{TeamRank{3, 4}, false}, // 不是赢家，不算翻山失败
	}

	for _, tt := range tests {
		got := tt.rank.IsClimbFailed()
		if got != tt.expected {
			t.Errorf("TeamRank%v.IsClimbFailed() = %v, want %v", tt.rank, got, tt.expected)
		}
	}
}

func TestGameRound_IsClimbing(t *testing.T) {
	gr := NewGameRound(RankA)

	// 未设置MaxTrump和Trump，不在翻山
	gr.MaxTrump = RankNone
	if gr.IsClimbing() {
		t.Error("should not be climbing when MaxTrump is RankNone")
	}

	// 设置MaxTrump但Trump不等于MaxTrump
	gr.MaxTrump = RankA
	gr.Trump = Rank10
	if gr.IsClimbing() {
		t.Error("should not be climbing when Trump != MaxTrump")
	}

	// Trump等于MaxTrump，正在翻山
	gr.Trump = RankA
	if !gr.IsClimbing() {
		t.Error("should be climbing when Trump == MaxTrump")
	}
}

func TestGameRound_ClimbFailed(t *testing.T) {
	gr := NewGameRound(RankA)

	// 设置初始级牌到最高（翻山状态）
	gr.Trumps[0] = RankA
	gr.Trumps[1] = Rank2
	gr.Trump = RankA
	gr.MaxTrump = RankA

	// 设置玩家
	for i := range gr.Players {
		gr.Players[i].UserId = int64(i + 1)
		gr.Players[i].Status = StatusReady
	}
	gr.Start()
	gr.Deal()

	// 模拟翻山失败（队伍A排名1,4）
	gr.Status = GameStatusFinished
	gr.Players[0].Rank = 1
	gr.Players[2].Rank = 4 // 队友是4，翻山失败
	gr.Players[1].Rank = 2
	gr.Players[3].Rank = 3
	gr.Players[0].Status = StatusFinished
	gr.Players[2].Status = StatusFinished
	gr.Players[1].Status = StatusFinished
	gr.Players[3].Status = StatusFinished

	// 结算
	err := gr.Settle(10, 100)
	if err != nil {
		t.Fatalf("Settle failed: %v", err)
	}

	// 下一回合
	err = gr.NextRound()
	if err != nil {
		t.Fatalf("NextRound failed: %v", err)
	}

	// 翻山失败，级牌不升级，保持A
	if gr.Trumps[0] != RankA {
		t.Errorf("team A trump should remain RankA after climb failed, got %d", gr.Trumps[0])
	}

	// 翻山失败次数应该是1
	if gr.ClimCounts[0] != 1 {
		t.Errorf("team A climb count should be 1, got %d", gr.ClimCounts[0])
	}
}

func TestGameRound_ClimbFailedThreeTimes(t *testing.T) {
	gr := NewGameRound(RankA)

	// 设置初始级牌到最高（翻山状态）
	gr.Trumps[0] = RankA
	gr.Trumps[1] = Rank2
	gr.Trump = RankA
	gr.MaxTrump = RankA

	// 模拟翻山失败3次
	for round := 1; round <= 3; round++ {
		// 设置玩家
		for i := range gr.Players {
			gr.Players[i].UserId = int64(i + 1)
			gr.Players[i].Status = StatusReady
		}
		gr.Start()
		gr.Deal()

		// 模拟翻山失败（队伍A排名1,4）
		gr.Status = GameStatusFinished
		gr.Players[0].Rank = 1
		gr.Players[2].Rank = 4
		gr.Players[1].Rank = 2
		gr.Players[3].Rank = 3
		gr.Players[0].Status = StatusFinished
		gr.Players[2].Status = StatusFinished
		gr.Players[1].Status = StatusFinished
		gr.Players[3].Status = StatusFinished

		// 结算
		err := gr.Settle(10, 100)
		if err != nil {
			t.Fatalf("Round %d: Settle failed: %v", round, err)
		}

		// 下一回合
		err = gr.NextRound()
		if err != nil {
			t.Fatalf("Round %d: NextRound failed: %v", round, err)
		}

		if round < 3 {
			// 前两次失败，级牌保持A，失败次数累加
			if gr.Trumps[0] != RankA {
				t.Errorf("Round %d: team A trump should remain RankA, got %d", round, gr.Trumps[0])
			}
			if gr.ClimCounts[0] != int8(round) {
				t.Errorf("Round %d: climb count should be %d, got %d", round, round, gr.ClimCounts[0])
			}
		} else {
			// 第三次失败，级牌重置为2
			if gr.Trumps[0] != Rank2 {
				t.Errorf("Round %d: team A trump should reset to Rank2, got %d", round, gr.Trumps[0])
			}
			// 失败次数重置为0
			if gr.ClimCounts[0] != 0 {
				t.Errorf("Round %d: climb count should be reset to 0, got %d", round, gr.ClimCounts[0])
			}
		}

		// 重新设置翻山状态（除了最后一轮）
		if round < 3 {
			gr.Trump = RankA
		}
	}
}

func TestGameRound_ClimbSuccess(t *testing.T) {
	gr := NewGameRound(RankA)

	// 设置初始级牌到最高（翻山状态）
	gr.Trumps[0] = RankA
	gr.Trumps[1] = Rank2
	gr.Trump = RankA
	gr.MaxTrump = RankA

	// 先模拟一次翻山失败
	gr.ClimCounts[0] = 2 // 已经失败2次

	// 设置玩家
	for i := range gr.Players {
		gr.Players[i].UserId = int64(i + 1)
		gr.Players[i].Status = StatusReady
	}
	gr.Start()
	gr.Deal()

	// 模拟翻山成功（队伍A排名1,2 双上）
	gr.Status = GameStatusFinished
	gr.Players[0].Rank = 1
	gr.Players[2].Rank = 2
	gr.Players[1].Rank = 4
	gr.Players[3].Rank = 4
	gr.Players[0].Status = StatusFinished
	gr.Players[2].Status = StatusFinished
	gr.Players[1].Status = StatusFinished
	gr.Players[3].Status = StatusFinished

	// 结算
	err := gr.Settle(10, 100)
	if err != nil {
		t.Fatalf("Settle failed: %v", err)
	}

	// 下一回合
	err = gr.NextRound()
	if err != nil {
		t.Fatalf("NextRound failed: %v", err)
	}

	// 翻山成功，所有状态应该重置（相当于重新开始游戏）
	// 失败次数应该重置为0
	if gr.ClimCounts[0] != 0 {
		t.Errorf("team A climb count should be reset to 0 after success, got %d", gr.ClimCounts[0])
	}
	if gr.ClimCounts[1] != 0 {
		t.Errorf("team B climb count should be reset to 0 after success, got %d", gr.ClimCounts[1])
	}

	// 翻山成功，级牌重置为Rank2（重新开始）
	if gr.Trumps[0] != Rank2 {
		t.Errorf("team A trump should reset to Rank2 after climb success, got %d", gr.Trumps[0])
	}
	if gr.Trumps[1] != Rank2 {
		t.Errorf("team B trump should reset to Rank2 after climb success, got %d", gr.Trumps[1])
	}

	// 当前级牌也应该重置为Rank2
	if gr.Trump != Rank2 {
		t.Errorf("current trump should reset to Rank2 after climb success, got %d", gr.Trump)
	}
}
