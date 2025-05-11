package api

import (
	"context"
	"log"
	"talkFlow/config"
	"talkFlow/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func GetProfile(c *gin.Context) {
	// 获取JWT中的用户名
	username, exists := c.Get("username")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	userCollection := config.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"username": username.(string)}).Decode(&user)
	if err != nil {
		c.JSON(500, gin.H{"code": 50001, "error": "获取用户信息失败"})
		log.Println("获取用户信息失败:", err)
		return
	}

	c.JSON(200, gin.H{"code": 20000, "username": user.Username, "email": user.Email, "avatar": user.Avatar})
}
