package guandan

import (
	"testing"
)

func TestNewPlayer(t *testing.T) {
	player := NewPlayer(12345)
	if player == nil {
		t.Fatal("NewPlayer returned nil")
	}
	if player.UserId != 12345 {
		t.Errorf("expected UserId 12345, got %d", player.UserId)
	}
	if player.Status != StatusWaiting {
		t.Errorf("expected status %v, got %v", StatusWaiting, player.Status)
	}
}

func TestPlayer_SetHand(t *testing.T) {
	player := NewPlayer(1)
	cards := Cards{
		{Rank: Rank3, Suit: SuitSpader},
		{Rank: Rank4, Suit: SuitHeart},
		{Rank: Rank5, Suit: SuitClub},
	}
	player.SetHand(cards)

	if len(player.Hand) != 3 {
		t.Errorf("expected 3 cards, got %d", len(player.Hand))
	}
}

func TestPlayer_HandCount(t *testing.T) {
	player := NewPlayer(1)
	if player.HandCount() != 0 {
		t.Errorf("expected 0 cards, got %d", player.HandCount())
	}

	player.Hand = Cards{
		{Rank: Rank3, Suit: SuitSpader},
		{Rank: Rank4, Suit: SuitHeart},
	}
	if player.HandCount() != 2 {
		t.Errorf("expected 2 cards, got %d", player.HandCount())
	}
}

func TestPlayer_Play_Success(t *testing.T) {
	player := NewPlayer(1)
	player.Hand = Cards{
		{Rank: Rank3, Suit: SuitSpader},
		{Rank: Rank3, Suit: SuitHeart},
		{Rank: Rank4, Suit: SuitClub},
		{Rank: Rank5, Suit: SuitDiamond},
	}

	// 打出一对3
	pattern := Pattern{
		Type: PatternTypePair,
		Cards: Cards{
			{Rank: Rank3, Suit: SuitSpader},
			{Rank: Rank3, Suit: SuitHeart},
		},
	}

	success := player.Play(pattern)
	if !success {
		t.Error("Play should succeed")
	}

	// 检查手牌减少
	if player.HandCount() != 2 {
		t.Errorf("expected 2 cards remaining, got %d", player.HandCount())
	}

	// 检查已打出的牌
	if player.PlayedCount() != 1 {
		t.Errorf("expected 1 played pattern, got %d", player.PlayedCount())
	}

	// 检查打出的牌型
	if player.Played[0].Type != PatternTypePair {
		t.Errorf("expected PatternTypePair, got %v", player.Played[0].Type)
	}
}

func TestPlayer_Play_Fail_NoCards(t *testing.T) {
	player := NewPlayer(1)
	player.Hand = Cards{
		{Rank: Rank3, Suit: SuitSpader},
		{Rank: Rank4, Suit: SuitHeart},
	}

	// 尝试打出手牌中没有的牌
	pattern := Pattern{
		Type: PatternTypePair,
		Cards: Cards{
			{Rank: Rank5, Suit: SuitSpader},
			{Rank: Rank5, Suit: SuitHeart},
		},
	}

	success := player.Play(pattern)
	if success {
		t.Error("Play should fail when cards not in hand")
	}

	// 手牌不应该变化
	if player.HandCount() != 2 {
		t.Errorf("hand should remain unchanged, got %d cards", player.HandCount())
	}
}

func TestPlayer_Play_Pass(t *testing.T) {
	player := NewPlayer(1)
	player.Hand = Cards{{Rank: Rank3, Suit: SuitSpader}}

	// 过牌（空Pattern，Type为None）
	pattern := Pattern{Type: PatternTypeNone}
	success := player.Play(pattern)
	if !success {
		t.Error("Play should succeed with pass (empty pattern)")
	}

	// 手牌不应该变化
	if player.HandCount() != 1 {
		t.Errorf("hand should remain unchanged, got %d cards", player.HandCount())
	}

	// 应该记录了过牌
	if player.PlayedCount() != 1 {
		t.Errorf("should record pass, got %d played", player.PlayedCount())
	}
}

func TestPlayer_Play_Fail_EmptyCards(t *testing.T) {
	player := NewPlayer(1)
	player.Hand = Cards{{Rank: Rank3, Suit: SuitSpader}}

	pattern := Pattern{
		Type:  PatternTypeSingle,
		Cards: Cards{},
	}

	success := player.Play(pattern)
	if success {
		t.Error("Play should fail with empty cards")
	}
}

func TestPlayer_Play_PassWithCards(t *testing.T) {
	player := NewPlayer(1)
	player.Hand = Cards{{Rank: Rank3, Suit: SuitSpader}}

	// 过牌但带了牌（应该成功，因为过牌时不会移除手牌）
	pattern := Pattern{
		Type:  PatternTypeNone,
		Cards: Cards{{Rank: Rank3, Suit: SuitSpader}},
	}

	success := player.Play(pattern)
	// 过牌时即使带了牌也不会移除手牌（取决于实现）
	if !success {
		t.Error("Play should succeed with PatternTypeNone")
	}
}

func TestPlayer_PlayedCards(t *testing.T) {
	player := NewPlayer(1)
	player.Hand = Cards{
		{Rank: Rank3, Suit: SuitSpader},
		{Rank: Rank3, Suit: SuitHeart},
		{Rank: Rank4, Suit: SuitClub},
		{Rank: Rank5, Suit: SuitDiamond},
	}

	// 打出一对3
	pattern1 := Pattern{
		Type: PatternTypePair,
		Cards: Cards{
			{Rank: Rank3, Suit: SuitSpader},
			{Rank: Rank3, Suit: SuitHeart},
		},
	}
	player.Play(pattern1)

	// 打出单张4
	pattern2 := Pattern{
		Type: PatternTypeSingle,
		Cards: Cards{
			{Rank: Rank4, Suit: SuitClub},
		},
	}
	player.Play(pattern2)

	// 检查已打出的牌
	playedCards := player.PlayedCards()
	if len(playedCards) != 3 {
		t.Errorf("expected 3 played cards, got %d", len(playedCards))
	}
}

func TestPlayer_PlayedCardCount(t *testing.T) {
	player := NewPlayer(1)
	player.Played = Patterns{
		{Type: PatternTypePair, Cards: Cards{{Rank: Rank3, Suit: SuitSpader}, {Rank: Rank3, Suit: SuitHeart}}},
		{Type: PatternTypeSingle, Cards: Cards{{Rank: Rank4, Suit: SuitClub}}},
	}

	if player.PlayedCardCount() != 3 {
		t.Errorf("expected 3 played cards, got %d", player.PlayedCardCount())
	}
}

func TestPlayer_PlayedCount(t *testing.T) {
	player := NewPlayer(1)
	if player.PlayedCount() != 0 {
		t.Errorf("expected 0 played patterns, got %d", player.PlayedCount())
	}

	player.Played = Patterns{
		{Type: PatternTypePair, Cards: Cards{{Rank: Rank3, Suit: SuitSpader}}},
		{Type: PatternTypeSingle, Cards: Cards{{Rank: Rank4, Suit: SuitClub}}},
	}

	if player.PlayedCount() != 2 {
		t.Errorf("expected 2 played patterns, got %d", player.PlayedCount())
	}
}

func TestTeamPlayers_Ranks(t *testing.T) {
	player1 := &Player{Rank: 1}
	player2 := &Player{Rank: 3}

	team := TeamPlayers{player1, player2}
	ranks := team.Ranks()

	if ranks[0] != 1 {
		t.Errorf("expected rank 1, got %d", ranks[0])
	}
	if ranks[1] != 3 {
		t.Errorf("expected rank 3, got %d", ranks[1])
	}
}

func TestPlayer_Play_DuplicateCards(t *testing.T) {
	// 测试打出重复的牌（两副牌中有相同的牌）
	player := NewPlayer(1)
	player.Hand = Cards{
		{Rank: Rank3, Suit: SuitSpader},
		{Rank: Rank3, Suit: SuitSpader}, // 两副牌中的相同牌
		{Rank: Rank4, Suit: SuitHeart},
	}

	// 打出两张相同的3
	pattern := Pattern{
		Type: PatternTypePair,
		Cards: Cards{
			{Rank: Rank3, Suit: SuitSpader},
			{Rank: Rank3, Suit: SuitSpader},
		},
	}

	success := player.Play(pattern)
	if !success {
		t.Error("Play should succeed with duplicate cards from two decks")
	}

	if player.HandCount() != 1 {
		t.Errorf("expected 1 card remaining, got %d", player.HandCount())
	}
}

func TestPlayer_StatusTransitions(t *testing.T) {
	player := NewPlayer(1)

	// 初始状态
	if player.Status != StatusWaiting {
		t.Errorf("initial status should be StatusWaiting")
	}

	// 准备
	player.Status = StatusReady
	if player.Status != StatusReady {
		t.Errorf("status should be StatusReady")
	}

	// 游戏中
	player.Status = StatusPlaying
	if player.Status != StatusPlaying {
		t.Errorf("status should be StatusPlaying")
	}

	// 结束
	player.Status = StatusFinished
	if player.Status != StatusFinished {
		t.Errorf("status should be StatusFinished")
	}
}

func TestPlayer_WinnerInfo(t *testing.T) {
	player := NewPlayer(1)

	// 设置赢家信息
	player.IsWinner = true
	player.PointChange = 100
	player.CoinChange = 50

	if !player.IsWinner {
		t.Error("player should be winner")
	}
	if player.PointChange != 100 {
		t.Errorf("expected point change 100, got %d", player.PointChange)
	}
	if player.CoinChange != 50 {
		t.Errorf("expected coin change 50, got %d", player.CoinChange)
	}
}

func TestPlayer_LoserInfo(t *testing.T) {
	player := NewPlayer(1)

	// 设置输家信息
	player.IsWinner = false
	player.PointChange = -100
	player.CoinChange = -50

	if player.IsWinner {
		t.Error("player should not be winner")
	}
	if player.PointChange != -100 {
		t.Errorf("expected point change -100, got %d", player.PointChange)
	}
	if player.CoinChange != -50 {
		t.Errorf("expected coin change -50, got %d", player.CoinChange)
	}
}

func TestPlayer_LostControl(t *testing.T) {
	player := NewPlayer(1)

	if player.IsLostControl {
		t.Error("player should not be in lost control by default")
	}

	player.IsLostControl = true
	if !player.IsLostControl {
		t.Error("player should be in lost control")
	}
}
