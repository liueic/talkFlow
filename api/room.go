package api

import (
	"context"
	"math/rand"
	"strconv"
	"strings"
	"talkFlow/config"
	"talkFlow/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 生成房间的随机号
func randomJoinCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	var sb strings.Builder
	for i := 0; i < length; i++ {
		sb.WriteByte(charset[seededRand.Intn(len(charset))])
	}
	return sb.String()
}

// 前端传来的时间应当是分钟数
type CreateRoomRequest struct {
	Name       string `json:"name" binding:"required"`
	ExpireTime string `json:"expire_time" binding:"required"`
}

func CreateRoom(c *gin.Context) {
	roomCollection := config.DB.Collection("rooms")

	var req CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 40001, "error": "参数错误"})
		return
	}

	// 获取JWT中的用户名
	username, exists := c.Get("username")
	if !exists {
		c.JSON(401, gin.H{"code": 40101, "error": "未授权"})
		return
	}

	// 转换过期时间为整数（分钟）
	expireMinutes, err := strconv.Atoi(req.ExpireTime)
	if err != nil || expireMinutes <= 0 {
		c.JSON(400, gin.H{"code": 40002, "error": "过期时间格式错误"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	JoinCode := randomJoinCode(5)

	room := models.Room{
		Creater:    username.(string),
		Name:       req.Name,
		Joiner:     []string{username.(string)},
		JoinCode:   JoinCode,
		CreateTime: primitive.NewDateTimeFromTime(time.Now()),
		ExpireTime: primitive.NewDateTimeFromTime(time.Now().Add(time.Duration(expireMinutes) * time.Minute)), // 过期时间
		Status:     models.RoomOngoing,                                                                        // 0: 进行中
		IP:         c.ClientIP(),                                                                              // 获取创建房间的IP
	}

	_, err = roomCollection.InsertOne(ctx, room)
	if err != nil {
		c.JSON(500, gin.H{"code": 50001, "error": "创建房间失败"})
		return
	}

	c.JSON(200, gin.H{"code": 20000, "message": "房间创建成功", "join_code": JoinCode})
}
