package guandan

import "errors"

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

type PatternType uint8

const (
	PatternTypeNone          PatternType = iota // 单张
	PatternTypeSingle                           // 单张
	PatternTypePair                             // 对
	PatternTypeTrips                            // 三同张
	PatternTypeFullHouse                        // 三带对
	PatternTypeTripsSeq                         // 三同连张
	PatternTypePairSeq                          // 三连对
	PatternTypeStraight                         // 顺子（5张）
	PatternTypeStraightFlush                    // 同花顺（5张）
	PatternTypeBomb                             // 炸弹（>=4张，相同点数或特殊级牌炸弹）
	PatternTypeFourJokers                       // 四大天王
)

type PlayStatus int8

const (
	StatusWaiting  PlayStatus = iota // 等待中
	StatusReady                      // 准备好
	StatusPlaying                    // 游戏中
	StatusFinished                   // 已结束
)

type GameRoundStatus int8

const (
	GameStatusWaiting  GameRoundStatus = iota // 等待中
	GameStatusPlaying                         // 游戏中
	GameStatusFinished                        // 已结束
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
