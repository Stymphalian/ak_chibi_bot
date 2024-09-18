package akdb

import "time"

type Room struct {
	RoomId      int64
	ChannelName string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
