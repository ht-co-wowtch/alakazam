package pb

import (
	"errors"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/bytes"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/encoding/binary"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/websocket"
)

const (
	// MaxBodySize max proto body size
	MaxBodySize = int32(1 << 12)
)

//
// |---------|--------|-----------|---------|
// | Package | Header | Operation |   Body  |
// |---------|--------|-----------|---------|
// | 4 bytes | 2 bytes|  4 bytes  | ? bytes |
// |---------|--------|---------|-----------|
// |					14 bytes			|
// |----------------------------------------|
//
// Package: 整個封包長度
// Header: 整個封包表頭長度
// Operation: 封包意義識別
// Body: 封包真正的內容
// =============================================================
// Package - Header = Body
//
const (
	// Protocol 長度的byte長度
	PackSize = 4

	// Protocol Header的byte長度
	HeaderSize = 2

	// Protocol 動作意義的byte長度
	OpSize = 4

	// 回覆心跳Body的byte長度
	HeartSize = 4

	// Protocol Header的總長度
	RawHeaderSize = PackSize + HeaderSize + OpSize

	maxPackSize = MaxBodySize + int32(RawHeaderSize)

	// Protocol 長度的byte位置範圍
	PackOffset = 0

	// Protocol 整個header長度的byte位置範圍
	// Protocol 長度 - header長度 = Body長度
	HeaderOffset = PackOffset + PackSize

	// Protocol動作意義的byte位置範圍
	OpOffset = HeaderOffset + HeaderSize

	// 回覆心跳Body的byte位置範圍
	heartOffset = OpOffset + OpSize
)

var (
	// 封包長度大小超過限定的長度
	ErrProtoPackLen = errors.New("default server codec pack length error")

	// 封包Header長度不符合規定
	ErrProtoHeaderLen = errors.New("default server codec header length error")
)

var (
	// 處理tcp資料的flag
	ProtoReady = &Proto{Op: OpProtoReady}

	// tcp close連線
	ProtoFinish = &Proto{Op: OpProtoFinish}
)

// Proto內容寫至bytes
func (p *Proto) WriteTo(b *bytes.Writer) {
	var (
		packLen = RawHeaderSize + int32(len(p.Body))
		buf     = b.Peek(RawHeaderSize)
	)
	binary.BigEndian.PutInt32(buf[PackOffset:], packLen)
	binary.BigEndian.PutInt16(buf[HeaderOffset:], int16(RawHeaderSize))
	binary.BigEndian.PutInt32(buf[OpOffset:], p.Op)
	if p.Body != nil {
		b.Write(p.Body)
	}
}

// 從websocket讀出內容
func (p *Proto) ReadWebsocket(ws websocket.Conn) (err error) {
	var (
		bodyLen   int
		headerLen int16
		packLen   int32
		buf       []byte
	)
	if _, buf, err = ws.ReadMessage(); err != nil {
		return
	}
	if len(buf) < RawHeaderSize {
		return ErrProtoPackLen
	}

	packLen = binary.BigEndian.Int32(buf[PackOffset:HeaderOffset])
	headerLen = binary.BigEndian.Int16(buf[HeaderOffset:OpOffset])
	p.Op = binary.BigEndian.Int32(buf[OpOffset:])
	if packLen > maxPackSize {
		return ErrProtoPackLen
	}
	if headerLen != RawHeaderSize {
		return ErrProtoHeaderLen
	}
	if bodyLen = int(packLen - int32(headerLen)); bodyLen > 0 {
		p.Body = buf[headerLen:packLen]
	} else {
		p.Body = nil
	}
	return
}

// Websocket寫入Proto內容
func (p *Proto) WriteWebsocket(ws websocket.Conn) (err error) {
	var (
		buf     []byte
		packLen int
	)

	packLen = RawHeaderSize + len(p.Body)
	if err = ws.WriteHeader(websocket.BinaryMessage, packLen); err != nil {
		return
	}
	if buf, err = ws.Peek(RawHeaderSize); err != nil {
		return
	}
	binary.BigEndian.PutInt32(buf[PackOffset:], int32(packLen))
	binary.BigEndian.PutInt16(buf[HeaderOffset:], int16(RawHeaderSize))
	binary.BigEndian.PutInt32(buf[OpOffset:], p.Op)
	if p.Body != nil {
		err = ws.WriteBody(p.Body)
	}
	return
}

// Websocket回覆心跳結果
func (p *Proto) WriteWebsocketHeart(wr websocket.Conn, online int32) (err error) {
	var (
		buf     []byte
		packLen int
	)
	packLen = RawHeaderSize + HeartSize
	// websocket header
	if err = wr.WriteHeader(websocket.BinaryMessage, packLen); err != nil {
		return
	}
	if buf, err = wr.Peek(packLen); err != nil {
		return
	}
	// proto header
	binary.BigEndian.PutInt32(buf[PackOffset:], int32(packLen))
	binary.BigEndian.PutInt16(buf[HeaderOffset:], int16(RawHeaderSize))
	binary.BigEndian.PutInt32(buf[OpOffset:], p.Op)
	// proto body
	binary.BigEndian.PutInt32(buf[heartOffset:], online)
	return
}
