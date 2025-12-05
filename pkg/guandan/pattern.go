package guandan

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

// 三连对、三同连张、顺子、同花顺时，可视作1
type Pattern struct {
	PlayerId  int8
	Type      PatternType
	Trump     Rank
	Cards     Cards
	MainPoint uint8 // 主要点数（用于比较牌型大小, 当为顺子，三连对的时候返回最大的牌）
	SubPoint  uint8 // 次要点数（备用）
	Length    int   // 多张的炸弹
	SameSuit  bool  // 是否为同花
}

func (cs Cards) Pattern(trump Rank) (pattern *Pattern) {
	return NewPattern(cs, trump)
}

// Detect 检测牌型
func NewPattern(cards Cards, trump Rank) (p *Pattern) {
	p = new(Pattern)
	p.Type = PatternTypeNone
	p.Trump = trump
	p.Cards = cards
	p.MainPoint = 0
	p.SameSuit = false
	p.Length = len(p.Cards)

	if p.Length == 0 {
		return
	}

	if len(p.Cards) == 4 && p.Cards.HasFourJokers() { // 四大天王
		p.Type = PatternTypeFourJokers
		return
	}

	// 1. 分离万能牌和普通牌
	var wildCount int
	var normalCards Cards
	for _, c := range p.Cards {
		if c.IsWild(trump) {
			wildCount++
		} else {
			normalCards = append(normalCards, c)
		}
	}

	// 2. 统计普通牌的点数
	rankCounts := make(map[Rank]int)
	for _, c := range normalCards {
		rankCounts[c.Rank]++
	}

	// 3. 检查是否同花 (仅检查普通牌，万能牌视为匹配)
	isFlush := true
	if len(normalCards) > 0 {
		firstSuit := normalCards[0].Suit
		for _, c := range normalCards {
			if c.Suit != firstSuit {
				isFlush = false
				break
			}
		}
	} else {
		// 全是万能牌，视为同花
		isFlush = true
	}
	p.SameSuit = isFlush

	// 4. 辅助判断：是否为炸弹 (普通牌只有一种点数)
	// 万能牌最多2张，不可能单独构成炸弹(>=4张)
	isBomb := false
	var bombRank Rank
	if len(rankCounts) == 1 {
		for r := range rankCounts {
			bombRank = r
		}
		if p.Length >= 4 {
			isBomb = true
		}
	}

	// 5. 根据张数判断牌型
	switch p.Length {
	case 1:
		p.Type = PatternTypeSingle
		if len(normalCards) > 0 {
			p.MainPoint = normalCards[0].Rank.Weight(trump)
		} else {
			// 单张万能牌，视为当前最大的单张（级牌）
			p.MainPoint = uint8(RankLevel)
		}
	case 2:
		// 对子
		if wildCount >= 1 || (len(rankCounts) == 1) {
			p.Type = PatternTypePair
			if len(normalCards) > 0 {
				p.MainPoint = normalCards[0].Rank.Weight(trump)
			} else {
				p.MainPoint = uint8(RankLevel)
			}
		}
	case 3:
		// 三同张
		if wildCount >= 2 || (len(rankCounts) == 1) {
			p.Type = PatternTypeTrips
			if len(normalCards) > 0 {
				p.MainPoint = normalCards[0].Rank.Weight(trump)
			} else {
				p.MainPoint = uint8(RankLevel)
			}
		}
	case 4:
		if isBomb {
			p.Type = PatternTypeBomb
			p.MainPoint = bombRank.Weight(trump)
		}
	case 5:
		// 优先级：同花顺 > 炸弹 > 顺子 > 三带二
		// 检查同花顺
		if isFlush {
			if mp := checkSequence(rankCounts, wildCount, 5, 1); mp > 0 {
				p.Type = PatternTypeStraightFlush
				p.MainPoint = mp
				return
			}
		}
		// 检查炸弹
		if isBomb {
			p.Type = PatternTypeBomb
			p.MainPoint = bombRank.Weight(trump)
			return
		}
		// 检查顺子
		if mp := checkSequence(rankCounts, wildCount, 5, 1); mp > 0 {
			p.Type = PatternTypeStraight
			p.MainPoint = mp
			return
		}
		// 检查三带二
		if mp, sp := checkFullHouse(rankCounts, wildCount, trump); mp > 0 {
			p.Type = PatternTypeFullHouse
			p.MainPoint = mp
			p.SubPoint = sp
			return
		}
	case 6:
		// 优先级：炸弹 > 三同连张 > 三连对
		if isBomb {
			p.Type = PatternTypeBomb
			p.MainPoint = bombRank.Weight(trump)
			return
		}
		// 三同连张
		if mp := checkSequence(rankCounts, wildCount, 2, 3); mp > 0 {
			p.Type = PatternTypeTripsSeq
			p.MainPoint = mp
			return
		}
		// 三连对
		if mp := checkSequence(rankCounts, wildCount, 3, 2); mp > 0 {
			p.Type = PatternTypePairSeq
			p.MainPoint = mp
			return
		}
	default:
		if isBomb {
			p.Type = PatternTypeBomb
			p.MainPoint = bombRank.Weight(trump)
		}
	}
	return
}

// checkSequence 检查是否构成序列
// length: 序列长度 (如顺子为5，三连对为3)
// width: 每个点数的张数 (如顺子为1，三连对为2)
// 返回 MainPoint (序列最大牌的 Rank 值，如果是 A2345 则返回 5)
func checkSequence(rankCounts map[Rank]int, wildCount int, length int, width int) uint8 {
	// 遍历所有可能的起点
	// Rank2(1) ... RankA(13)
	// 特殊起点: RankNone(0) 代表 A, 2, 3...

	// 最大可能的起点: 13 - length + 1
	// 例如 length=5, maxStart=9 (Rank10). 10, J, Q, K, A
	maxStart := int(RankA) - length + 1

	// 寻找最大的 MainPoint，所以从大到小遍历?
	// 或者找到任意一个? 通常如果有万能牌，可能组成多种顺子。
	// 比如 2, 3, 4, 5, Wild. 可以是 A2345 (Main 5) 或 23456 (Main 6).
	// 显然取最大的。

	var bestMainPoint uint8 = 0

	// 遍历范围包括 0 (A当1) 和 1..maxStart
	starts := make([]int, 0, maxStart+1)
	starts = append(starts, 0) // A, 2, ...
	for i := 1; i <= maxStart; i++ {
		starts = append(starts, i)
	}

	for _, start := range starts {
		currentWild := wildCount
		possible := true
		var currentMainPoint uint8

		for i := range length {
			var r Rank
			if start == 0 && i == 0 {
				r = RankA
			} else {
				r = Rank(start + i)
			}

			count := rankCounts[r]
			if count > width {
				possible = false // 牌多了，不能构成标准序列 (除非多余的是万能牌? 不，rankCounts只包含普通牌)
				// 实际上如果普通牌多了，肯定不行，因为要求恰好构成序列
				break
			}
			needed := width - count
			if needed > currentWild {
				possible = false
				break
			}
			currentWild -= needed

			// 记录当前序列的最大牌
			// 如果是 A2345 (start=0), 最后一个是 i=4, r=Rank5. MainPoint=5.
			// 如果是 10JQKA (start=9), 最后一个是 i=4, r=RankA. MainPoint=13.
			if i == length-1 {
				currentMainPoint = uint8(r)
			}
		}

		// 检查是否有多余的普通牌 (rankCounts 中的牌必须都在序列中)
		if possible {
			for r, c := range rankCounts {
				inSeq := false
				for i := range length {
					var seqR Rank
					if start == 0 && i == 0 {
						seqR = RankA
					} else {
						seqR = Rank(start + i)
					}
					if r == seqR {
						inSeq = true
						break
					}
				}
				if !inSeq && c > 0 {
					possible = false
					break
				}
			}
		}

		if possible {
			if currentMainPoint > bestMainPoint {
				bestMainPoint = currentMainPoint
			}
		}
	}
	return bestMainPoint
}

// checkFullHouse 检查三带二
func checkFullHouse(rankCounts map[Rank]int, wildCount int, trump Rank) (uint8, uint8) {
	// 穷举三张的牌点 (Trips) 和 对子的牌点 (Pair)
	// Trips 可以是任意 Rank (1..13)
	// Pair 可以是任意 Rank (1..13)
	// Trips != Pair

	var bestMainPoint uint8 = 0
	var bestSubPoint uint8 = 0

	// 所有的 Rank 集合
	var ranks []Rank
	for r := Rank2; r <= RankA; r++ {
		ranks = append(ranks, r)
	}

	for _, tripsRank := range ranks {
		for _, pairRank := range ranks {
			if tripsRank == pairRank {
				continue
			}

			// 检查是否可行
			currentWild := wildCount

			// 需要 3 张 tripsRank
			countT := rankCounts[tripsRank]
			if countT > 3 {
				continue
			}
			needT := 3 - countT
			if needT > currentWild {
				continue
			}
			currentWild -= needT

			// 需要 2 张 pairRank
			countP := rankCounts[pairRank]
			if countP > 2 {
				continue
			}
			needP := 2 - countP
			if needP > currentWild {
				continue
			}
			// currentWild -= needP // 不需要真正减，只要够就行

			// 检查是否有多余的普通牌
			// 普通牌只能是 tripsRank 或 pairRank
			validCards := true
			for r := range rankCounts {
				if r != tripsRank && r != pairRank {
					validCards = false
					break
				}
			}
			if !validCards {
				continue
			}

			// 找到一个可行解，记录 MainPoint
			// 三带二的大小由三张的大小决定
			mp := tripsRank.Weight(trump)
			sp := pairRank.Weight(trump)
			if mp > bestMainPoint {
				bestMainPoint = mp
				bestSubPoint = sp
			} else if mp == bestMainPoint {
				if sp > bestSubPoint {
					bestSubPoint = sp
				}
			}
		}
	}
	return bestMainPoint, bestSubPoint
}

// GetLevel 获取牌型等级
// Level 5: >5张炸弹
// Level 4: 同花顺
// Level 3: 5张炸弹
// Level 2: 4张炸弹
// Level 1: 其他
func (p *Pattern) GetLevel() int {
	if p.Type == PatternTypeBomb {
		if p.Length > 5 {
			return 5
		} else if p.Length == 5 {
			return 3
		} else {
			return 2
		}
	}
	if p.Type == PatternTypeStraightFlush {
		return 4
	}
	return 1
}

// Compare 比较牌型大小
// 返回 1 (p > other), -1 (p < other), 0 (无法比较或相等)
func (p *Pattern) Compare(other *Pattern) int {
	// 1. 四大天王最大
	if p.Type == PatternTypeFourJokers {
		if other.Type == PatternTypeFourJokers {
			return 0
		}
		return 1
	}
	if other.Type == PatternTypeFourJokers {
		return -1
	}

	l1 := p.GetLevel()
	l2 := other.GetLevel()

	if l1 != l2 {
		if l1 > l2 {
			return 1
		}
		return -1
	}

	// 同等级比较
	if l1 == 5 { // >5张炸弹
		if p.Length != other.Length {
			if p.Length > other.Length {
				return 1
			}
			return -1
		}
		// 张数相同比点数
		if p.MainPoint > other.MainPoint {
			return 1
		} else if p.MainPoint < other.MainPoint {
			return -1
		}
		return 0
	}

	if l1 == 4 { // 同花顺
		if p.MainPoint > other.MainPoint {
			return 1
		} else if p.MainPoint < other.MainPoint {
			return -1
		}
		return 0
	}

	if l1 == 3 || l1 == 2 { // 5张炸弹 或 4张炸弹
		if p.MainPoint > other.MainPoint {
			return 1
		} else if p.MainPoint < other.MainPoint {
			return -1
		}
		return 0
	}

	// Level 1: 其他牌型 (单张、对子、三同张、三带二、三连对、三同连张、顺子)
	// 只有类型相同才能比较
	if p.Type != other.Type {
		return 0
	}

	// 类型相同，比 MainPoint
	// 注意：顺子、三连对、三同连张 也是比 MainPoint
	if p.MainPoint > other.MainPoint {
		return 1
	} else if p.MainPoint < other.MainPoint {
		return -1
	}

	// 如果 MainPoint 相同，且是三带二，比较 SubPoint
	if p.Type == PatternTypeFullHouse {
		if p.SubPoint > other.SubPoint {
			return 1
		} else if p.SubPoint < other.SubPoint {
			return -1
		}
	}

	return 0
}

type Patterns []Pattern

func (ps Patterns) Cards() (all Cards) {
	for _, p := range ps {
		all = append(all, p.Cards...)
	}
	return
}
