package models

import (
	"time"
)

type RoomStatus int

const (
	RoomOngoing RoomStatus = iota // 0: 进行中
	RoomEnded                     // 1: 已结束
)

type Room struct {
	ID         int64      `json:"id" db:"id"`
	Name       string     `json:"name" db:"name"`
	Creater    string     `json:"creater" db:"creater"`
	Joiner     []string   `json:"joiner" db:"-"`
	JoinerStr  string     `json:"-" db:"joiner"` // 用于数据库读写
	JoinCode   string     `json:"join_code" db:"join_code"`
	CreateTime time.Time  `json:"create_time" db:"create_time"`
	ExpireTime time.Time  `json:"expire_time" db:"expire_time"`
	Status     RoomStatus `json:"status" db:"status"`
	IP         string     `json:"ip" db:"ip"`
}

func (r *Room) IsOngoing() bool {
	now := time.Now()
	return r.Status == RoomOngoing && now.Before(r.ExpireTime)
}

func (r *Room) IsEnded() bool {
	now := time.Now()
	return r.Status == RoomEnded || now.After(r.ExpireTime)
}
