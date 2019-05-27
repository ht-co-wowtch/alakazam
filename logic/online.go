package logic

// 根據房間type與room id取房間在線人數
func (l *Logic) OnlineRoom(rooms []string) (res map[string]int32, err error) {
	res = make(map[string]int32, len(rooms))
	for _, room := range rooms {
		res[room] = l.roomCount[room]
	}
	return
}
