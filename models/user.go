package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Register struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Username   string             `bson:"username"`
	Password   string             `bson:"password"`
	Email      string             `bson:"email"`
	Avatar     string             `bson:"avatar"`
	CreatedAt  primitive.DateTime `bson:"created_at"`
	RegisterIP string             `bson:"register_ip"`
	IsRegister bool               `bson:"is_register"`
}

type Visitor struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Username   string             `bson:"username"`
	CreatedAt  primitive.DateTime `bson:"created_at"`
	VisitorIP  string             `bson:"visitor_ip"`
	IsRegister bool               `bson:"is_register"`
}
