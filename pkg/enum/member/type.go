package member

type Type int

const (
	Guest     Type = 0 // 訪客
	Marketing Type = 1 // 營銷
	PLayer    Type = 2 // 玩家
	Streamer  Type = 3 // 直播主
)

func ToType(code int32) Type {
	switch code {
	case 0:
		return Guest
	case 1:
		return Marketing
	case 2:
		return PLayer
	case 3:
		return Streamer
	}
	return Guest
}

func (t Type) String() string {
	switch t {
	case Guest:
		return "guest"
	case Marketing:
		return "marketing"
	case PLayer:
		return "player"
	case Streamer:
		return "streamer"
	}
	return "guest"
}
