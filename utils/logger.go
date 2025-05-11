package utils

import (
	"context"
	"talkFlow/config"
	"time"
)

type LogEntry struct {
	Username  string `bson:"username"`
	Error     string `bson:"error"`
	Timestamp string `bson:"timestamp"`
	IP        string `bson:"ip"`
}

// 记录日志到MongoDB
func Logger(username, errMsg, timestamp, ip string) error {
	logCollection := config.DB.Collection("log")

	logEntry := LogEntry{
		Username:  username,
		Error:     errMsg,
		Timestamp: timestamp,
		IP:        ip,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := logCollection.InsertOne(ctx, logEntry)
	return err
}
