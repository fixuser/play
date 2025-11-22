package guandan

// CanBeat 判断手牌中是否有大于 target 的牌型
func (handCards Cards) CanBeat(targetPattern *Pattern, trump Rank) bool {
	// 1. 统计手牌
	var wildCount int
	rankCounts := make(map[Rank]int)
	suitCards := make(map[Suit][]Rank) // 用于同花顺检查

	for _, c := range handCards {
		if c.IsWild(trump) {
			wildCount++
		} else {
			rankCounts[c.Rank]++
			suitCards[c.Suit] = append(suitCards[c.Suit], c.Rank)
		}
	}

	// 对手如果是四大天王
	if targetPattern.Type == PatternTypeFourJokers {
		return false
	}
	// 如果包含四大天王，必定能压制任何牌型
	if handCards.IsFourJokers() {
		return true
	}

	targetLevel := targetPattern.GetLevel()

	// 2. 检查更高等级的压制
	for l := targetLevel + 1; l <= 5; l++ {
		switch l {
		case 2: // 4张炸弹
			if searchBomb(rankCounts, wildCount, 4, 0, trump) {
				return true
			}
		case 3: // 5张炸弹
			if searchBomb(rankCounts, wildCount, 5, 0, trump) {
				return true
			}
		case 4: // 同花顺
			// 需要按花色检查
			for _, ranks := range suitCards {
				subRankCounts := make(map[Rank]int)
				for _, r := range ranks {
					subRankCounts[r]++
				}
				if searchSequence(subRankCounts, wildCount, 5, 1, 0, trump) {
					return true
				}
			}
			// 纯万能牌同花顺? 通常万能牌可以配成任意花色
			// 万能牌最多2张，不可能组成5张同花顺
		case 5: // >5张炸弹
			if searchBomb(rankCounts, wildCount, 6, 0, trump) {
				return true
			}
		}
	}

	// 3. 检查同等级压制
	switch targetLevel {
	case 5: // >5张炸弹
		// 先找张数更多的
		if searchBomb(rankCounts, wildCount, targetPattern.Length+1, 0, trump) {
			return true
		}
		// 张数相同找点数更大的
		if searchBomb(rankCounts, wildCount, targetPattern.Length, targetPattern.MainPoint, trump) {
			return true
		}
	case 4: // 同花顺
		for _, ranks := range suitCards {
			subRankCounts := make(map[Rank]int)
			for _, r := range ranks {
				subRankCounts[r]++
			}
			if searchSequence(subRankCounts, wildCount, 5, 1, targetPattern.MainPoint, trump) {
				return true
			}
		}
		// 纯万能牌同花顺 (如果 targetPattern 是同花顺，纯万能牌可以组成更大的同花顺吗？)
		// 万能牌最多2张，不可能组成5张同花顺
	case 3: // 5张炸弹
		if searchBomb(rankCounts, wildCount, 5, targetPattern.MainPoint, trump) {
			return true
		}
	case 2: // 4张炸弹
		if searchBomb(rankCounts, wildCount, 4, targetPattern.MainPoint, trump) {
			return true
		}
	case 1: // 普通牌型
		switch targetPattern.Type {
		case PatternTypeSingle:
			// 找单张 > target.MainPoint
			for r := Rank2; r <= RankJokerBig; r++ {
				if r.Weight(trump) > targetPattern.MainPoint {
					if rankCounts[r] > 0 {
						return true
					}
					// 万能牌也可以当单张
					if r.Weight(trump) == uint8(RankLevel) && wildCount > 0 {
						return true
					}
				}
			}
		case PatternTypePair:
			if searchBomb(rankCounts, wildCount, 2, targetPattern.MainPoint, trump) { // 复用 searchBomb 找对子
				return true
			}
		case PatternTypeTrips:
			if searchBomb(rankCounts, wildCount, 3, targetPattern.MainPoint, trump) { // 复用 searchBomb 找三张
				return true
			}
		case PatternTypeFullHouse:
			if searchFullHouse(rankCounts, wildCount, targetPattern.MainPoint, targetPattern.SubPoint, trump) {
				return true
			}
		case PatternTypeStraight:
			if searchSequence(rankCounts, wildCount, targetPattern.Length, 1, targetPattern.MainPoint, trump) {
				return true
			}
		case PatternTypeTripsSeq:
			if searchSequence(rankCounts, wildCount, targetPattern.Length/3, 3, targetPattern.MainPoint, trump) {
				return true
			}
		case PatternTypePairSeq:
			if searchSequence(rankCounts, wildCount, targetPattern.Length/2, 2, targetPattern.MainPoint, trump) {
				return true
			}
		}
	}

	return false
}

// searchBomb 查找炸弹 (或对子、三张)
// length: 需要的张数
// minMainPoint: 最小 MainPoint (不包含)
func searchBomb(rankCounts map[Rank]int, wildCount int, length int, minMainPoint uint8, trump Rank) bool {
	// 遍历所有点数
	for r := Rank2; r <= RankA; r++ { // 普通牌
		if r.Weight(trump) > minMainPoint {
			if rankCounts[r]+wildCount >= length {
				return true
			}
		}
	}
	// 级牌炸弹 (纯万能牌或级牌+万能牌)
	// 注意：rankCounts 中不包含万能牌(红桃级牌)，但包含其他花色的级牌
	// 如果 trump 是 RankA，那么 RankA 已经在上面循环处理了?
	// 不，RankA 在上面循环里。但是 rankCounts[trump] 只包含非红桃的级牌。
	// 万能牌 wildCount 是红桃级牌。
	// 所以 rankCounts[trump] + wildCount 就是所有级牌的数量。
	// 上面的循环已经覆盖了 trump 的情况。

	// 特殊：如果 minMainPoint < RankLevel，且我们有足够的万能牌+级牌
	// 上面循环 r=trump 时，r.Weight(trump) 是 RankLevel。
	// 所以如果 RankLevel > minMainPoint，就会检查。

	// 还有大小王炸弹? 不，大小王不能组成普通炸弹，只能组成四大天王。
	// 除非规则允许王炸? 掼蛋通常只有四大天王。

	return false
}

// searchSequence 查找序列 (顺子、连对、三连张)
// length: 几连 (如顺子5连，length=5)
// width: 每项几张 (如顺子width=1)
// minMainPoint: 最小 MainPoint (不包含)
func searchSequence(rankCounts map[Rank]int, wildCount int, length int, width int, minMainPoint uint8, trump Rank) bool {
	maxStart := int(RankA) - length + 1
	starts := make([]int, 0, maxStart+1)
	starts = append(starts, 0) // A, 2, ...
	for i := 1; i <= maxStart; i++ {
		starts = append(starts, i)
	}

	for _, start := range starts {
		currentWild := wildCount
		possible := true
		var currentMainPoint uint8

		for i := 0; i < length; i++ {
			var r Rank
			if start == 0 {
				if i == 0 {
					r = RankA
				} else {
					r = Rank(i + 1)
				}
			} else {
				r = Rank(start + i)
			}

			count := rankCounts[r]
			needed := 0
			if count < width {
				needed = width - count
			}

			if needed > currentWild {
				possible = false
				break
			}
			currentWild -= needed

			if i == length-1 {
				currentMainPoint = uint8(r)
			}
		}

		if possible {
			if currentMainPoint > minMainPoint {
				return true
			}
		}
	}
	return false
}

// searchFullHouse 查找三带二
// minMainPoint: 必须 > minMainPoint
// minSubPoint: 如果 MainPoint == minMainPoint，则 SubPoint 必须 > minSubPoint
func searchFullHouse(rankCounts map[Rank]int, wildCount int, minMainPoint uint8, minSubPoint uint8, trump Rank) bool {
	var ranks []Rank
	for r := Rank2; r <= RankA; r++ {
		ranks = append(ranks, r)
	}

	for _, tripsRank := range ranks {
		for _, pairRank := range ranks {
			if tripsRank == pairRank {
				continue
			}

			mp := tripsRank.Weight(trump)
			sp := pairRank.Weight(trump)

			// 检查大小是否满足
			if mp < minMainPoint {
				continue
			}
			if mp == minMainPoint && sp <= minSubPoint {
				continue
			}

			// 检查是否有足够的牌
			currentWild := wildCount

			countT := rankCounts[tripsRank]
			needT := 0
			if countT < 3 {
				needT = 3 - countT
			}
			if needT > currentWild {
				continue
			}
			currentWild -= needT

			countP := rankCounts[pairRank]
			needP := 0
			if countP < 2 {
				needP = 2 - countP
			}
			if needP > currentWild {
				continue
			}

			return true
		}
	}
	return false
}
