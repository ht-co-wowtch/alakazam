package grpc

import (
	"errors"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/bytes"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/encoding/binary"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/websocket"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol"
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
// |---------|--------|---------|-----------|----------|---------|
// |					14 bytes					   |
// |---------------------------------------------------|
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
	_packSize = 4

	// Protocol Header的byte長度
	_headerSize = 2

	// Protocol 動作意義的byte長度
	_opSize = 4

	// 回覆心跳Body的byte長度
	_heartSize = 4

	// Protocol Header的總長度
	_rawHeaderSize = _packSize + _headerSize + _opSize

	_maxPackSize = MaxBodySize + int32(_rawHeaderSize)

	// Protocol 長度的byte位置範圍
	_packOffset = 0

	// Protocol 整個header長度的byte位置範圍
	// Protocol 長度 - header長度 = Body長度
	_headerOffset = _packOffset + _packSize

	// Protocol動作意義的byte位置範圍
	_opOffset = _headerOffset + _headerSize

	// 回覆心跳Body的byte位置範圍
	_heartOffset = _opOffset + _opSize
)

var (
	// 封包長度大小超過限定的長度
	ErrProtoPackLen = errors.New("default server codec pack length error")

	// 封包Header長度不符合規定
	ErrProtoHeaderLen = errors.New("default server codec header length error")
)

var (
	// 處理tcp資料的flag
	ProtoReady = &Proto{Op: protocol.OpProtoReady}

	// tcp close連線
	ProtoFinish = &Proto{Op: protocol.OpProtoFinish}
)

// Proto內容寫至bytes
func (p *Proto) WriteTo(b *bytes.Writer) {
	var (
		packLen = _rawHeaderSize + int32(len(p.Body))
		buf     = b.Peek(_rawHeaderSize)
	)
	binary.BigEndian.PutInt32(buf[_packOffset:], packLen)
	binary.BigEndian.PutInt16(buf[_headerOffset:], int16(_rawHeaderSize))
	binary.BigEndian.PutInt32(buf[_opOffset:], p.Op)
	if p.Body != nil {
		b.Write(p.Body)
	}
}

// 從websocket讀出內容
func (p *Proto) ReadWebsocket(ws *websocket.Conn) (err error) {
	var (
		bodyLen   int
		headerLen int16
		packLen   int32
		buf       []byte
	)
	if _, buf, err = ws.ReadMessage(); err != nil {
		return
	}
	if len(buf) < _rawHeaderSize {
		return ErrProtoPackLen
	}

	packLen = binary.BigEndian.Int32(buf[_packOffset:_headerOffset])
	headerLen = binary.BigEndian.Int16(buf[_headerOffset:_opOffset])
	p.Op = binary.BigEndian.Int32(buf[_opOffset:])
	if packLen > _maxPackSize {
		return ErrProtoPackLen
	}
	if headerLen != _rawHeaderSize {
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
func (p *Proto) WriteWebsocket(ws *websocket.Conn) (err error) {
	var (
		buf     []byte
		packLen int
	)

	packLen = _rawHeaderSize + len(p.Body)
	if err = ws.WriteHeader(websocket.BinaryMessage, packLen); err != nil {
		return
	}
	if buf, err = ws.Peek(_rawHeaderSize); err != nil {
		return
	}
	binary.BigEndian.PutInt32(buf[_packOffset:], int32(packLen))
	binary.BigEndian.PutInt16(buf[_headerOffset:], int16(_rawHeaderSize))
	binary.BigEndian.PutInt32(buf[_opOffset:], p.Op)
	if p.Body != nil {
		err = ws.WriteBody(p.Body)
	}
	return
}

// Websocket回覆心跳結果
func (p *Proto) WriteWebsocketHeart(wr *websocket.Conn, online int32) (err error) {
	var (
		buf     []byte
		packLen int
	)
	packLen = _rawHeaderSize + _heartSize
	// websocket header
	if err = wr.WriteHeader(websocket.BinaryMessage, packLen); err != nil {
		return
	}
	if buf, err = wr.Peek(packLen); err != nil {
		return
	}
	// proto header
	binary.BigEndian.PutInt32(buf[_packOffset:], int32(packLen))
	binary.BigEndian.PutInt16(buf[_headerOffset:], int16(_rawHeaderSize))
	binary.BigEndian.PutInt32(buf[_opOffset:], p.Op)
	// proto body
	binary.BigEndian.PutInt32(buf[_heartOffset:], online)
	return
}
