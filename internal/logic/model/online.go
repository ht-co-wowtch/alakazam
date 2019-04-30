package model

// Online ip and room online.
type Online struct {
	Server    string           `json:"server"`
	RoomCount map[string]int32 `json:"room_count"`
	Updated   int64            `json:"updated"`
}