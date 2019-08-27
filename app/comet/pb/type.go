package pb

const (
	// client要求連線到某一個房間
	OpAuth = int32(1)

	// server回覆連線到某一個房間結果
	OpAuthReply = int32(2)

	// client 發送心跳
	OpHeartbeat = int32(3)

	// server 回覆心跳結果
	OpHeartbeatReply = int32(4)

	// server批次訊息推送給client
	OpBatchRaw = int32(5)

	// server訊息推送給client
	OpRaw = int32(6)

	// 更換房間
	OpChangeRoom = int32(7)

	// 回覆更換房間結果
	OpChangeRoomReply = int32(8)

	// close連線
	OpProtoFinish = int32(20)

	// websocket送訊息要回應
	OpProtoReady = int32(21)
)
