package protocol

import (
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/bufio"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/encoding/binary"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"
)

func Read(rr *bufio.Reader, p *grpc.Proto) (err error) {
	var (
		bodyLen   int
		headerLen int16
		packLen   int32
	)

	packLen, headerLen, err = readHeader(rr, p)
	if bodyLen = int(packLen - int32(headerLen)); bodyLen > 0 {
		p.Body, err = rr.Pop(bodyLen)
	} else {
		p.Body = nil
	}
	return
}

func ReadMessage(rr *bufio.Reader, p *grpc.Proto) (protos []grpc.Proto, err error) {
	var (
		bodyLen   int
		headerLen int16
		packLen   int32
	)

	packLen, headerLen, err = readHeader(rr, p)
	for offset := int(headerLen); offset < int(packLen); offset += int(packLen) {
		proto := new(grpc.Proto)
		packLen, headerLen, err = readHeader(rr, proto)
		if bodyLen = int(packLen - int32(headerLen)); bodyLen > 0 {
			proto.Body, err = rr.Pop(bodyLen)
		} else {
			proto.Body = nil
		}
		protos = append(protos, *proto)
	}
	return
}

func Write(wr *bufio.Writer, p *grpc.Proto) (err error) {
	var (
		buf     []byte
		packLen int32
	)

	packLen = grpc.RawHeaderSize + int32(len(p.Body))
	if buf, err = wr.Peek(grpc.RawHeaderSize); err != nil {
		return
	}
	binary.BigEndian.PutInt32(buf[grpc.PackOffset:], packLen)
	binary.BigEndian.PutInt16(buf[grpc.HeaderOffset:], int16(grpc.RawHeaderSize))
	binary.BigEndian.PutInt32(buf[grpc.OpOffset:], p.Op)
	if p.Body != nil {
		_, err = wr.Write(p.Body)
	}
	return wr.Flush()
}

func readHeader(rr *bufio.Reader, p *grpc.Proto) (packLen int32, headerLen int16, err error) {
	var (
		buf []byte
	)
	if buf, err = rr.Pop(grpc.RawHeaderSize); err != nil {
		return
	}

	packLen = binary.BigEndian.Int32(buf[grpc.PackOffset:grpc.HeaderOffset])
	headerLen = binary.BigEndian.Int16(buf[grpc.HeaderOffset:grpc.OpOffset])
	p.Op = binary.BigEndian.Int32(buf[grpc.OpOffset:])
	return
}
