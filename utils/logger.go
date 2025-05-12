package utils

import (
	"context"
	"errors"
	"talkFlow/config"
	"time"
)

type LogEntry struct {
	Username  string
	Error     string
	Timestamp string
	IP        string
}

// 记录日志到SQLite，并返回插入的行ID和错误
func Logger(username, errMsg, timestamp, ip string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	insertSQL := `
        INSERT INTO log (username, error, timestamp, ip)
        VALUES (?, ?, ?, ?)
    `
	// 检查 config.DB 是否为 nil，防止空指针异常
	if config.DB == nil {
		return 0, errors.New("database is not initialized")
	}

	result, err := config.DB.ExecContext(ctx, insertSQL, username, errMsg, timestamp, ip)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}
