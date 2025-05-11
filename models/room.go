package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoomStatus int

const (
	RoomOngoing RoomStatus = iota // 0: 进行中
	RoomEnded                     // 1: 已结束
)

type Room struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Name       string             `bson:"name"`
	Creater    string             `bson:"creater"`
	Joiner     []string           `bson:"joiner"`
	JoinCode   string             `bson:"join_code"`
	CreateTime primitive.DateTime `bson:"create_time"`
	ExpireTime primitive.DateTime `bson:"expire_time"`
	Status     RoomStatus         `bson:"status"` // 0: 进行中, 1: 已结束
	IP         string             `bson:"ip"`
}

func (r *Room) IsOngoing() bool {
	now := primitive.NewDateTimeFromTime(time.Now())
	return r.Status == RoomOngoing && now < r.ExpireTime
}

func (r *Room) IsEnded() bool {
	now := primitive.NewDateTimeFromTime(time.Now())
	return r.Status == RoomEnded || now >= r.ExpireTime
}
