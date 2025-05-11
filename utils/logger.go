package utils

import (
	"context"
	"talkFlow/config"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LogEntry struct {
	Username  string `bson:"username"`
	Error     string `bson:"error"`
	Timestamp string `bson:"timestamp"`
	IP        string `bson:"ip"`
}

// 记录日志到MongoDB，并返回ObjectID
func Logger(username, errMsg, timestamp, ip string) (primitive.ObjectID, error) {
	logCollection := config.DB.Collection("log")

	logEntry := LogEntry{
		Username:  username,
		Error:     errMsg,
		Timestamp: timestamp,
		IP:        ip,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := logCollection.InsertOne(ctx, logEntry)
	if err != nil {
		return primitive.NilObjectID, err
	}

	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, err
	}
	return id, nil
}
