package message

import (
	"fmt"
	"time"
)

type refectorError struct {
	errs    []error
	time    time.Time
	occur   string
	swallow bool
}

type messageError struct {
	error   error
	msgId   int64
	mid     int64
	message string
}

func (m messageError) Error() string {
	return fmt.Sprintf("insert message error: %s msg_id: %d mid: %d message: %s", m.error.Error(), m.msgId, m.mid, m.message)
}

type MysqlRoomMessageError struct {
	error   error
	msgId   int64
	room    []int32
	message string
}

func (m MysqlRoomMessageError) Error() string {
	return fmt.Sprintf("insert room message error: %s msg_id: %d room: %v message: %s", m.error.Error(), m.msgId, m.room, m.message)
}

type MysqlRedEnvelopeMessageError struct {
	error         error
	redEnvelopeId string
	msgId         int64
	mid           int64
	message       string
}

func (m MysqlRedEnvelopeMessageError) Error() string {
	return fmt.Sprintf("insert red envelope message error: %s msg_id: %d mid: %d red_envelope_id: %s message: %s", m.error.Error(), m.msgId, m.mid, m.redEnvelopeId, m.message)
}

type MysqlAdminMessageError struct {
	error   error
	msgId   int64
	room    []int32
	message string
}

func (m MysqlAdminMessageError) Error() string {
	return fmt.Sprintf("insert admin message error: %s msg_id: %d room: %v message: %s", m.error.Error(), m.msgId, m.room, m.message)
}
