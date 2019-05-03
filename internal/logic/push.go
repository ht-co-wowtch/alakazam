package logic

import (
	"context"
	log "github.com/golang/glog"
)

// 根據user key推送
func (l *Logic) PushKeys(c context.Context, keys []string, msg []byte) (err error) {
	// 取該user key所在的server name
	servers, err := l.dao.ServersByKeys(c, keys)
	if err != nil {
		return
	}

	// 整理出以下結構
	// key => server name
	// value => 該server name下的user key
	pushKeys := make(map[string][]string)
	for i, key := range keys {
		server := servers[i]
		if server != "" && key != "" {
			pushKeys[server] = append(pushKeys[server], key)
		}
	}
	// 根據server name與user key來推送，另還有operation條件是不變的
	for server := range pushKeys {
		if err = l.dao.PushMsg(c, server, pushKeys[server], msg); err != nil {
			return
		}
	}
	return
}

// 根據user id推送
func (l *Logic) PushMids(c context.Context, mids []int64, msg []byte) (err error) {
	// 根據user id拿 user key
	keyServers, _, err := l.dao.KeysByMids(c, mids)
	if err != nil {
		return
	}
	keys := make(map[string][]string)

	// key: user key
	// server: user所在的server name
	for key, server := range keyServers {
		if key == "" || server == "" {
			log.Warningf("push key:%s server:%s is empty", key, server)
			continue
		}
		// 根據server name分組
		keys[server] = append(keys[server], key)
	}
	// 根據server name與user key來推送，另還有operation條件是不變的
	for server, keys := range keys {
		if err = l.dao.PushMsg(c, server, keys, msg); err != nil {
			return
		}
	}
	return
}

// 單一房間推送
func (l *Logic) PushRoom(c context.Context, room string, msg []byte) (err error) {
	return l.dao.BroadcastRoomMsg(c, room, msg)
}

// 所有房間推送但有限制operation
func (l *Logic) PushAll(c context.Context, speed int32, msg []byte) (err error) {
	return l.dao.BroadcastMsg(c, speed, msg)
}
