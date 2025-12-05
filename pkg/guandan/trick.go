package guandan

type Trick struct {
	PatternType  PatternType // 占前面四个字节
	PlayerIndex  uint8       // 占后面四个字节
	PatternIndex uint8
}

// IsPass 是否为过牌
func (t Trick) IsPass() bool {
	return t.PatternType == PatternTypeNone
}

type Tricks []Trick

// MarshalBinary 序列化为二进制
func (ts Tricks) MarshalBinary() (data []byte, err error) {
	data = make([]byte, len(ts)*2)
	for i, t := range ts {
		data[i*2] = byte(t.PlayerIndex&0xF) | byte((t.PatternType&0xF)<<4)
		data[i*2+1] = byte(t.PatternIndex)
	}
	return
}

// UnmarshalBinary 从二进制反序列化
func (ts *Tricks) UnmarshalBinary(data []byte) error {
	length := len(data) / 2
	*ts = make(Tricks, length)
	for i := 0; i < length; i++ {
		(*ts)[i] = Trick{
			PlayerIndex:  uint8(data[i*2] & 0x0F),
			PatternType:  PatternType((data[i*2] >> 4) & 0x0F),
			PatternIndex: uint8(data[i*2+1]),
		}
	}
	return nil
}
