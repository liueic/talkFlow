package api

import (
	"context"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"talkFlow/config"
	"talkFlow/models"
	"talkFlow/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
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
		c.JSON(401, gin.H{
			"code":  40101,
			"error": "未授权",
		})
		log.Printf("未授权: %s", c.ClientIP())
		utils.Logger("unknown", "未授权", time.Now().Format(time.RFC3339), c.ClientIP())
		return
	}

	// 转换过期时间为整数（分钟）
	expireMinutes, err := strconv.Atoi(req.ExpireTime)
	if err != nil || expireMinutes <= 0 {
		c.JSON(400, gin.H{
			"code":  40002,
			"error": "过期时间格式错误",
		})
		log.Println("过期时间格式错误:", err)
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

	logID, _ := utils.Logger(username.(string), err.Error(), time.Now().Format(time.RFC3339), c.ClientIP())
	if err != nil {
		c.JSON(500, gin.H{
			"code":   50001,
			"error":  "创建房间失败",
			"log_id": logID.Hex(),
		})
		log.Println("创建房间失败:", err)
		return
	}

	c.JSON(200, gin.H{
		"code":      20000,
		"message":   "房间创建成功",
		"join_code": JoinCode,
	})

}

// 用户ID依赖前端通过浏览器指纹进行收集
func JoinRoom(c *gin.Context) {

	var req struct {
		JoinCode  string `json:"join_code" binding:"required"`
		VisitorID string `json:"visitor_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 40001, "error": "参数错误"})
		log.Println("参数错误:", err)
		return
	}

	roomCollection := config.DB.Collection("rooms")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var room models.Room
	err := roomCollection.FindOne(ctx, bson.M{"join_code": req.JoinCode}).Decode(&room)
	if err != nil {
		// 房间不存在
		c.JSON(404, gin.H{
			"code":  40401,
			"error": "房间不存在",
		})
		log.Println("房间不存在:", req.JoinCode)
		return
	}

	// 检查房间是否已过期
	if room.IsEnded() {
		c.JSON(400, gin.H{
			"code":  40002,
			"error": "房间已结束",
		})
		log.Println("房间已结束:", room.ID.Hex())
		return
	}

	c.JSON(200, gin.H{
		"code":    20000,
		"message": "加入房间成功",
		"room":    room,
	})

}
