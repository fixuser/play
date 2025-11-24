package guandan

// Search 查找手牌中是否有大于 target 的牌型，返回能压制的牌 (最小的一个)
func (handCards Cards) Search(targetPattern *Pattern, trump Rank) Cards {

	// 1. 统计手牌
	var wildCards []Card
	cardsMap := make(map[Rank][]Card)
	suitCardsMap := make(map[Suit][]Card) // 用于同花顺检查

	for _, c := range handCards {
		if c.IsWild(trump) {
			wildCards = append(wildCards, c)
		} else {
			cardsMap[c.Rank] = append(cardsMap[c.Rank], c)
			suitCardsMap[c.Suit] = append(suitCardsMap[c.Suit], c)
		}
	}

	// 对手如果是四大天王
	if targetPattern.Type == PatternTypeFourJokers {
		return nil
	}

	targetLevel := targetPattern.GetLevel()

	// 辅助函数：检查四大天王
	checkFourJokers := func() Cards {
		small := cardsMap[RankJokerSmall]
		big := cardsMap[RankJokerBig]
		if len(small) == 2 && len(big) == 2 {
			res := make(Cards, 0, 4)
			res = append(res, small...)
			res = append(res, big...)
			return res
		}
		return nil
	}

	// 2. 检查同等级压制 (优先出同等级的)
	switch targetLevel {
	case 5: // >5张炸弹
		// 张数相同找点数更大的
		if res := searchBomb(cardsMap, wildCards, targetPattern.Length, targetPattern.MainPoint, trump); res != nil {
			return res
		}
		// 先找张数更多的 (直到8张)
		for len := targetPattern.Length + 1; len <= 8; len++ {
			if res := searchBomb(cardsMap, wildCards, len, 0, trump); res != nil {
				return res
			}
		}
	case 4: // 同花顺
		for _, cards := range suitCardsMap {
			subMap := make(map[Rank][]Card)
			for _, c := range cards {
				subMap[c.Rank] = append(subMap[c.Rank], c)
			}
			if res := searchSequence(subMap, wildCards, 5, 1, targetPattern.MainPoint, trump); res != nil {
				return res
			}
		}
	case 3: // 5张炸弹
		if res := searchBomb(cardsMap, wildCards, 5, targetPattern.MainPoint, trump); res != nil {
			return res
		}
	case 2: // 4张炸弹
		if res := searchBomb(cardsMap, wildCards, 4, targetPattern.MainPoint, trump); res != nil {
			return res
		}
	case 1: // 普通牌型
		switch targetPattern.Type {
		case PatternTypeSingle:
			// 找单张 > target.MainPoint
			for r := Rank2; r <= RankJokerBig; r++ {
				if r.Weight(trump) > targetPattern.MainPoint {
					if list, ok := cardsMap[r]; ok && len(list) > 0 {
						return Cards{list[0]}
					}
				}
			}
			// 万能牌也可以当单张 (RankLevel)
			if uint8(RankLevel) > targetPattern.MainPoint && len(wildCards) > 0 {
				return Cards{wildCards[0]}
			}
		case PatternTypePair:
			if res := searchBomb(cardsMap, wildCards, 2, targetPattern.MainPoint, trump); res != nil {
				return res
			}
		case PatternTypeTrips:
			if res := searchBomb(cardsMap, wildCards, 3, targetPattern.MainPoint, trump); res != nil {
				return res
			}
		case PatternTypeFullHouse:
			if res := searchFullHouse(cardsMap, wildCards, targetPattern.MainPoint, targetPattern.SubPoint, trump); res != nil {
				return res
			}
		case PatternTypeStraight:
			if res := searchSequence(cardsMap, wildCards, targetPattern.Length, 1, targetPattern.MainPoint, trump); res != nil {
				return res
			}
		case PatternTypeTripsSeq:
			if res := searchSequence(cardsMap, wildCards, targetPattern.Length/3, 3, targetPattern.MainPoint, trump); res != nil {
				return res
			}
		case PatternTypePairSeq:
			if res := searchSequence(cardsMap, wildCards, targetPattern.Length/2, 2, targetPattern.MainPoint, trump); res != nil {
				return res
			}
		}
	}

	// 3. 检查更高等级的压制
	for l := targetLevel + 1; l <= 5; l++ {
		switch l {
		case 2: // 4张炸弹
			if res := searchBomb(cardsMap, wildCards, 4, 0, trump); res != nil {
				return res
			}
		case 3: // 5张炸弹
			if res := searchBomb(cardsMap, wildCards, 5, 0, trump); res != nil {
				return res
			}
		case 4: // 同花顺
			for _, cards := range suitCardsMap {
				subMap := make(map[Rank][]Card)
				for _, c := range cards {
					subMap[c.Rank] = append(subMap[c.Rank], c)
				}
				if res := searchSequence(subMap, wildCards, 5, 1, 0, trump); res != nil {
					return res
				}
			}
		case 5: // >5张炸弹
			// 搜索 6, 7, 8 张炸弹
			for len := 6; len <= 8; len++ {
				if res := searchBomb(cardsMap, wildCards, len, 0, trump); res != nil {
					return res
				}
			}
		}
	}

	// 4. 检查四大天王
	if res := checkFourJokers(); res != nil {
		return res
	}

	return nil
}

// SearchAll 查找手牌中所有大于 target 的牌型
func (handCards Cards) SearchAll(targetPattern *Pattern, trump Rank) []Cards {
	var results []Cards

	// 1. 统计手牌
	var wildCards []Card
	cardsMap := make(map[Rank][]Card)
	suitCardsMap := make(map[Suit][]Card)

	for _, c := range handCards {
		if c.IsWild(trump) {
			wildCards = append(wildCards, c)
		} else {
			cardsMap[c.Rank] = append(cardsMap[c.Rank], c)
			suitCardsMap[c.Suit] = append(suitCardsMap[c.Suit], c)
		}
	}

	// 对手如果是四大天王，无牌可压
	if targetPattern.Type == PatternTypeFourJokers {
		return nil
	}

	targetLevel := targetPattern.GetLevel()

	// 辅助函数：检查四大天王
	checkFourJokers := func() Cards {
		small := cardsMap[RankJokerSmall]
		big := cardsMap[RankJokerBig]
		if len(small) == 2 && len(big) == 2 {
			res := make(Cards, 0, 4)
			res = append(res, small...)
			res = append(res, big...)
			return res
		}
		return nil
	}

	// 2. 检查同等级压制
	switch targetLevel {
	case 5: // >5张炸弹
		// 张数相同找点数更大的
		if res := searchBombAll(cardsMap, wildCards, targetPattern.Length, targetPattern.MainPoint, trump); len(res) > 0 {
			results = append(results, res...)
		}
		// 找张数更多的 (直到8张)
		for bombLen := targetPattern.Length + 1; bombLen <= 8; bombLen++ {
			if res := searchBombAll(cardsMap, wildCards, bombLen, 0, trump); len(res) > 0 {
				results = append(results, res...)
			}
		}
	case 4: // 同花顺
		for _, cards := range suitCardsMap {
			subMap := make(map[Rank][]Card)
			for _, c := range cards {
				subMap[c.Rank] = append(subMap[c.Rank], c)
			}
			if res := searchSequenceAll(subMap, wildCards, 5, 1, targetPattern.MainPoint, trump); len(res) > 0 {
				results = append(results, res...)
			}
		}
	case 3: // 5张炸弹
		if res := searchBombAll(cardsMap, wildCards, 5, targetPattern.MainPoint, trump); len(res) > 0 {
			results = append(results, res...)
		}
	case 2: // 4张炸弹
		if res := searchBombAll(cardsMap, wildCards, 4, targetPattern.MainPoint, trump); len(res) > 0 {
			results = append(results, res...)
		}
	case 1: // 普通牌型
		switch targetPattern.Type {
		case PatternTypeSingle:
			// 找单张 > target.MainPoint
			for r := Rank2; r <= RankJokerBig; r++ {
				if r.Weight(trump) > targetPattern.MainPoint {
					if list, ok := cardsMap[r]; ok && len(list) > 0 {
						results = append(results, Cards{list[0]})
					}
				}
			}
			// 万能牌也可以当单张
			if uint8(RankLevel) > targetPattern.MainPoint && len(wildCards) > 0 {
				results = append(results, Cards{wildCards[0]})
			}
		case PatternTypePair:
			if res := searchBombAll(cardsMap, wildCards, 2, targetPattern.MainPoint, trump); len(res) > 0 {
				results = append(results, res...)
			}
		case PatternTypeTrips:
			if res := searchBombAll(cardsMap, wildCards, 3, targetPattern.MainPoint, trump); len(res) > 0 {
				results = append(results, res...)
			}
		case PatternTypeFullHouse:
			if res := searchFullHouseAll(cardsMap, wildCards, targetPattern.MainPoint, targetPattern.SubPoint, trump); len(res) > 0 {
				results = append(results, res...)
			}
		case PatternTypeStraight:
			if res := searchSequenceAll(cardsMap, wildCards, targetPattern.Length, 1, targetPattern.MainPoint, trump); len(res) > 0 {
				results = append(results, res...)
			}
		case PatternTypeTripsSeq:
			if res := searchSequenceAll(cardsMap, wildCards, targetPattern.Length/3, 3, targetPattern.MainPoint, trump); len(res) > 0 {
				results = append(results, res...)
			}
		case PatternTypePairSeq:
			if res := searchSequenceAll(cardsMap, wildCards, targetPattern.Length/2, 2, targetPattern.MainPoint, trump); len(res) > 0 {
				results = append(results, res...)
			}
		}
	}

	// 3. 检查更高等级的压制
	for l := targetLevel + 1; l <= 5; l++ {
		switch l {
		case 2: // 4张炸弹
			if res := searchBombAll(cardsMap, wildCards, 4, 0, trump); len(res) > 0 {
				results = append(results, res...)
			}
		case 3: // 5张炸弹
			if res := searchBombAll(cardsMap, wildCards, 5, 0, trump); len(res) > 0 {
				results = append(results, res...)
			}
		case 4: // 同花顺
			for _, cards := range suitCardsMap {
				subMap := make(map[Rank][]Card)
				for _, c := range cards {
					subMap[c.Rank] = append(subMap[c.Rank], c)
				}
				if res := searchSequenceAll(subMap, wildCards, 5, 1, 0, trump); len(res) > 0 {
					results = append(results, res...)
				}
			}
		case 5: // >5张炸弹
			// 搜索 6, 7, 8 张炸弹
			for bombLen := 6; bombLen <= 8; bombLen++ {
				if res := searchBombAll(cardsMap, wildCards, bombLen, 0, trump); len(res) > 0 {
					results = append(results, res...)
				}
			}
		}
	}

	// 4. 检查四大天王
	if res := checkFourJokers(); res != nil {
		results = append(results, res)
	}

	return results
}

// searchBomb 查找炸弹 (或对子、三张)
// length: 需要的张数
// minMainPoint: 最小 MainPoint (不包含)
func searchBomb(cardsMap map[Rank][]Card, wildCards []Card, length int, minMainPoint uint8, trump Rank) Cards {
	// 遍历所有点数
	for r := Rank2; r <= RankA; r++ { // 普通牌
		if r.Weight(trump) > minMainPoint {
			list := cardsMap[r]
			count := len(list)
			if count+len(wildCards) >= length {
				res := make(Cards, 0, length)
				take := length
				if count < take {
					take = count
				}
				res = append(res, list[:take]...)
				needed := length - len(res)
				if needed > 0 {
					res = append(res, wildCards[:needed]...)
				}
				return res
			}
		}
	}
	return nil
}

// searchSequence 查找序列 (顺子、连对、三连张)
// length: 几连 (如顺子5连，length=5)
// width: 每项几张 (如顺子width=1)
// minMainPoint: 最小 MainPoint (不包含)
func searchSequence(cardsMap map[Rank][]Card, wildCards []Card, length int, width int, minMainPoint uint8, trump Rank) Cards {
	maxStart := int(RankA) - length + 1
	starts := make([]int, 0, maxStart+1)
	starts = append(starts, 0) // A, 2, ...
	for i := 1; i <= maxStart; i++ {
		starts = append(starts, i)
	}

	for _, start := range starts {
		var result Cards
		usedWildCount := 0
		possible := true
		var currentMainPoint uint8

		for i := 0; i < length; i++ {
			var r Rank
			if start == 0 && i == 0 {
				r = RankA
			} else {
				r = Rank(start + i)
			}

			list := cardsMap[r]
			count := len(list)
			take := width
			if count < take {
				take = count
			}
			result = append(result, list[:take]...)

			needed := width - take
			if needed > 0 {
				usedWildCount += needed
				if usedWildCount > len(wildCards) {
					possible = false
					break
				}
			}

			if i == length-1 {
				currentMainPoint = uint8(r)
			}
		}

		if possible {
			if currentMainPoint > minMainPoint {
				result = append(result, wildCards[:usedWildCount]...)
				return result
			}
		}
	}
	return nil
}

// searchFullHouse 查找三带二
// minMainPoint: 必须 > minMainPoint
// minSubPoint: 如果 MainPoint == minMainPoint，则 SubPoint 必须 > minSubPoint
func searchFullHouse(cardsMap map[Rank][]Card, wildCards []Card, minMainPoint uint8, minSubPoint uint8, trump Rank) Cards {
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
			listT := cardsMap[tripsRank]
			countT := len(listT)
			listP := cardsMap[pairRank]
			countP := len(listP)

			needed := 0
			if countT < 3 {
				needed += 3 - countT
			}
			if countP < 2 {
				needed += 2 - countP
			}

			if needed <= len(wildCards) {
				var res Cards
				// Add trips
				takeT := 3
				if countT < 3 {
					takeT = countT
				}
				res = append(res, listT[:takeT]...)

				// Add pair
				takeP := 2
				if countP < 2 {
					takeP = countP
				}
				res = append(res, listP[:takeP]...)

				// Add wildcards
				res = append(res, wildCards[:needed]...)
				return res
			}
		}
	}
	return nil
}

func searchBombAll(cardsMap map[Rank][]Card, wildCards []Card, length int, minMainPoint uint8, trump Rank) []Cards {
	var results []Cards
	for r := Rank2; r <= RankA; r++ {
		if r.Weight(trump) > minMainPoint {
			list := cardsMap[r]
			count := len(list)
			if count+len(wildCards) >= length {
				res := make(Cards, 0, length)
				take := length
				if count < take {
					take = count
				}
				res = append(res, list[:take]...)
				needed := length - len(res)
				if needed > 0 {
					res = append(res, wildCards[:needed]...)
				}
				results = append(results, res)
			}
		}
	}
	return results
}

func searchSequenceAll(cardsMap map[Rank][]Card, wildCards []Card, length int, width int, minMainPoint uint8, trump Rank) []Cards {
	var results []Cards
	maxStart := int(RankA) - length + 1
	starts := make([]int, 0, maxStart+1)
	starts = append(starts, 0) // A
	for i := 1; i <= maxStart; i++ {
		starts = append(starts, i)
	}

	for _, start := range starts {
		var result Cards
		usedWildCount := 0
		possible := true
		var currentMainPoint uint8

		for i := 0; i < length; i++ {
			var r Rank
			if start == 0 && i == 0 {
				r = RankA
			} else {
				r = Rank(start + i)
			}

			list := cardsMap[r]
			count := len(list)
			take := width
			if count < take {
				take = count
			}
			result = append(result, list[:take]...)

			needed := width - take
			if needed > 0 {
				usedWildCount += needed
				if usedWildCount > len(wildCards) {
					possible = false
					break
				}
			}

			if i == length-1 {
				currentMainPoint = uint8(r)
			}
		}

		if possible {
			if currentMainPoint > minMainPoint {
				result = append(result, wildCards[:usedWildCount]...)
				results = append(results, result)
			}
		}
	}
	return results
}

func searchFullHouseAll(cardsMap map[Rank][]Card, wildCards []Card, minMainPoint uint8, minSubPoint uint8, trump Rank) []Cards {
	var results []Cards
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

			if mp < minMainPoint {
				continue
			}
			if mp == minMainPoint && sp <= minSubPoint {
				continue
			}

			listT := cardsMap[tripsRank]
			countT := len(listT)
			listP := cardsMap[pairRank]
			countP := len(listP)

			needed := 0
			if countT < 3 {
				needed += 3 - countT
			}
			if countP < 2 {
				needed += 2 - countP
			}

			if needed <= len(wildCards) {
				var res Cards
				takeT := 3
				if countT < 3 {
					takeT = countT
				}
				res = append(res, listT[:takeT]...)

				takeP := 2
				if countP < 2 {
					takeP = countP
				}
				res = append(res, listP[:takeP]...)

				res = append(res, wildCards[:needed]...)
				results = append(results, res)
			}
		}
	}
	return results
}
