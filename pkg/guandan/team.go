package guandan

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

// TeamPlayers 表示一队玩家
type TeamPlayers [2]*Player

// Ranks 表示一队玩家的排名
func (tps *TeamPlayers) Ranks() TeamRank {
	return TeamRank{tps[0].Rank, tps[1].Rank}
}
