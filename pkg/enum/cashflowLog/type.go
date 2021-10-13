package cashflowLog

type Type int

const (
	DiamondAdd     Type = 20   // 鑽石增加
	DiamondSub     Type = 21   // 鑽石減少
	LiveGiveCharge Type = 3004
	LiveTakeCharge Type = 3006
)

func (t Type) String() string {
	switch t {
	case LiveGiveCharge:
		return "live-give-charge"
	case LiveTakeCharge:
		return "live_take_charge"
	case DiamondSub:
		return "diamond-sub"
	case DiamondAdd:
		return "diamond-add"
	}
	return ""
}
