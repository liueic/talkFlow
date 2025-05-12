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

	JoinCode := randomJoinCode(5)

	room := models.Room{
		Creater:    username.(string),
		Name:       req.Name,
		Joiner:     []string{username.(string)},
		JoinCode:   JoinCode,
		CreateTime: time.Now(),
		ExpireTime: time.Now().Add(time.Duration(expireMinutes) * time.Minute), // 过期时间
		Status:     models.RoomOngoing,                                         // 0: 进行中
		IP:         c.ClientIP(),                                               // 获取创建房间的IP
	}

	insertSQL := `
		INSERT INTO rooms (creater, name, joiner, join_code, create_time, expire_time, status, ip)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	// 将 joiner 字段（string slice）序列化为字符串存储到数据库
	joinerStr := strings.Join(room.Joiner, ",")

	_, err = config.DB.ExecContext(context.Background(), insertSQL,
		room.Creater, room.Name, joinerStr, room.JoinCode,
		room.CreateTime, room.ExpireTime, room.Status, room.IP,
	)
	if err != nil {
		var logMsg string
		if err != nil {
			logMsg = err.Error()
		} else {
			logMsg = ""
		}
		logID, _ := utils.Logger(username.(string), logMsg, time.Now().Format(time.RFC3339), c.ClientIP())
		c.JSON(500, gin.H{
			"code":    50001,
			"error":   "创建房间失败",
			"eventID": logID,
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

	// 不需要提前声明 room 或 visitor 变量
	var req struct {
		JoinCode  string `json:"join_code" binding:"required"`
		VisitorID string `json:"visitor_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 40001, "error": "参数错误"})
		log.Println("参数错误:", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var existingRoom int
	checkRoomSQL := `SELECT COUNT(*) FROM rooms WHERE join_code = ?`
	err := config.DB.QueryRowContext(ctx, checkRoomSQL, req.JoinCode).Scan(&existingRoom)

	if err != nil || existingRoom == 0 {
		c.JSON(404, gin.H{
			"code":  40401,
			"error": "房间不存在",
		})
		log.Println("房间不存在:", req.JoinCode)
		return
	}

	// 查询房间信息到 room 结构体
	var room models.Room
	roomSQL := `SELECT id, creater, name, joiner, join_code, create_time, expire_time, status, ip FROM rooms WHERE join_code = ?`
	err = config.DB.QueryRowContext(ctx, roomSQL, req.JoinCode).Scan(
		&room.ID,
		&room.Creater,
		&room.Name,
		&room.JoinerStr, // 先用字符串接收
		&room.JoinCode,
		&room.CreateTime,
		&room.ExpireTime,
		&room.Status,
		&room.IP,
	)
	if err != nil {
		c.JSON(404, gin.H{
			"code":  40401,
			"error": "房间不存在",
		})
		log.Println("房间不存在:", req.JoinCode)
		return
	}
	// 反序列化 joiner 字段
	room.Joiner = strings.Split(room.JoinerStr, ",")

	// 查询房间信息到 room 结构体后
	if !room.IsOngoing() {
		c.JSON(400, gin.H{
			"code":  40002,
			"error": "房间已结束",
		})
		log.Println("房间已结束:", room.ID)
		return
	}

	insertVisitorSQL := `
        INSERT INTO visitors (visitor_id, created_at, visitor_ip, is_register)
        VALUES (?, ?, ?, ?)
    `
	_, err = config.DB.ExecContext(ctx, insertVisitorSQL, req.VisitorID, time.Now(), c.ClientIP(), false)
	if err != nil {
		logID, _ := utils.Logger(req.VisitorID, err.Error(), time.Now().Format(time.RFC3339), c.ClientIP())
		log.Println("记录访客失败:", err)

		c.JSON(500, gin.H{
			"code":    50001,
			"error":   "记录访客失败",
			"eventID": logID,
		})
		return
	}

	// 查询房间ID（假设rooms表有id字段且为INTEGER PRIMARY KEY AUTOINCREMENT）
	var roomID int64
	getRoomIDSQL := `SELECT id FROM rooms WHERE join_code = ?`
	err = config.DB.QueryRowContext(ctx, getRoomIDSQL, req.JoinCode).Scan(&roomID)
	if err != nil {
		c.JSON(500, gin.H{
			"code":  50003,
			"error": "获取房间ID失败",
		})
		log.Println("获取房间ID失败:", err)
		return
	}

	c.JSON(200, gin.H{
		"code":    20000,
		"message": "加入房间成功",
		"room":    roomID,
		"url":     "/api/v1/ws?join_code=" + req.JoinCode + "&id=" + req.VisitorID,
	})
}
