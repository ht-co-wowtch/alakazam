package protocol

import (
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/bufio"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/encoding/binary"
	"gitlab.com/jetfueltw/cpw/alakazam/comet/pb"
)

func Read(rr *bufio.Reader, p *pb.Proto) (err error) {
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

func ReadMessage(rr *bufio.Reader, p *pb.Proto) (protos []pb.Proto, err error) {
	var (
		bodyLen   int
		headerLen int16
		packLen   int32
	)

	packLen, headerLen, err = readHeader(rr, p)
	for offset := int(headerLen); offset < int(packLen); offset += int(packLen) {
		proto := new(pb.Proto)
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

func Write(wr *bufio.Writer, p *pb.Proto) (err error) {
	var (
		buf     []byte
		packLen int32
	)

	packLen = pb.RawHeaderSize + int32(len(p.Body))
	if buf, err = wr.Peek(pb.RawHeaderSize); err != nil {
		return
	}
	binary.BigEndian.PutInt32(buf[pb.PackOffset:], packLen)
	binary.BigEndian.PutInt16(buf[pb.HeaderOffset:], int16(pb.RawHeaderSize))
	binary.BigEndian.PutInt32(buf[pb.OpOffset:], p.Op)
	if p.Body != nil {
		_, err = wr.Write(p.Body)
	}
	return wr.Flush()
}

func readHeader(rr *bufio.Reader, p *pb.Proto) (packLen int32, headerLen int16, err error) {
	var (
		buf []byte
	)
	if buf, err = rr.Pop(pb.RawHeaderSize); err != nil {
		return
	}

	packLen = binary.BigEndian.Int32(buf[pb.PackOffset:pb.HeaderOffset])
	headerLen = binary.BigEndian.Int16(buf[pb.HeaderOffset:pb.OpOffset])
	p.Op = binary.BigEndian.Int32(buf[pb.OpOffset:])
	return
}
