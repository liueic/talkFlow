package api

import (
	"context"
	"log"
	"talkFlow/config"
	"talkFlow/models"
	"talkFlow/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func GetProfile(c *gin.Context) {
	// 获取JWT中的用户名
	username, exists := c.Get("username")
	if !exists {
		c.JSON(401, gin.H{"code": 40101, "error": "未授权"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.Register
	query := `SELECT id, username, email, avatar FROM register WHERE username = ?`
	err := config.DB.QueryRowContext(ctx, query, username.(string)).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Avatar,
	)
	if err != nil {
		logID, _ := utils.Logger(username.(string), err.Error(), time.Now().Format(time.RFC3339), c.ClientIP())
		c.JSON(500, gin.H{
			"code":   50001,
			"error":  "获取用户信息失败",
			"log_id": logID,
		})
		log.Println("获取用户信息失败:", err)
		return
	}

	c.JSON(200, gin.H{
		"code":     20000,
		"username": user.Username,
		"email":    user.Email,
		"avatar":   user.Avatar,
	})
}
