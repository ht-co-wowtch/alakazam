package pb

// websocket message Operation type
const (
	// OpAuth
	// client要求連線到某一個房間
	OpAuth = int32(1)

	// OpAuthReply
	// server回覆連線到某一個房間結果
	OpAuthReply = int32(2)

	// OpHeartbeat
	// client 發送心跳
	OpHeartbeat = int32(3)

	// OpHeartbeatReply
	// server 回覆心跳結果
	OpHeartbeatReply = int32(4)

	// OpBatchRaw
	// server批次訊息推送給client
	OpBatchRaw = int32(5)

	// OpRaw
	// server訊息推送給client
	OpRaw = int32(6)

	// OpChangeRoom
	// 更換房間
	OpChangeRoom = int32(7)

	// OpChangeRoomReply
	// 回覆更換房間結果
	OpChangeRoomReply = int32(8)

	// OpCloseTopMessage
	// 取消置頂訊息
	OpCloseTopMessage = int32(9)

	// OpPaidRoomExpiry
	// 付費房驗證(月卡)
	OpPaidRoomExpiry = int32(10)

	// OpPaidRoomExpiryReply
	// 回覆付費房驗證(月卡)結果
	OpPaidRoomExpiryReply = int32(11)

	// OpPaidRoomDiamond
	// 付費房驗證(鑽石)
	OpPaidRoomDiamond = int32(12)

	// OpPaidRoomDiamondReply
	// 回覆付費房驗證(鑽石)
	OpPaidRoomDiamondReply = int32(13)

	// OpProtoFinish
	// close連線
	OpProtoFinish = int32(20)

	// OpProtoReady
	// websocket送訊息要回應
	OpProtoReady = int32(21)
)
