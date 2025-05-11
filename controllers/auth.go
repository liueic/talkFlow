package controllers

import (
	"context"
	"crypto/md5"
	"fmt"
	"strings"
	"talkFlow/config"
	"talkFlow/models"
	"talkFlow/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// 注册逻辑
func Register(c *gin.Context) {
	userCollection := config.DB.Collection("users")

	// 只接收 Username、Password、Email
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	// 检查用户名是否已存在
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var existingUser models.User
	err := userCollection.FindOne(ctx, bson.M{"username": input.Username}).Decode(&existingUser)
	if err == nil {
		c.JSON(400, gin.H{"error": "用户名已存在"})
		return
	}

	// 生成gravatar
	email := strings.TrimSpace(strings.ToLower(input.Email))
	hash := md5.Sum([]byte(email))
	avatar := fmt.Sprintf("https://www.gravatar.com/avatar/%x", hash)

	// 加密密码
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	// 构造User对象
	user := models.User{
		ID:         primitive.NewObjectID(),
		Username:   input.Username,
		Password:   string(hashedPassword),
		Email:      input.Email,
		Avatar:     avatar,
		CreatedAt:  primitive.NewDateTimeFromTime(time.Now()),
		RegisterIP: c.ClientIP(), // 获取注册IP，依赖Nginx或者Caddy
	}

	// 插入新用户
	_, err = userCollection.InsertOne(ctx, user)
	if err != nil {
		c.JSON(500, gin.H{"code": 50001, "error": "注册失败"})
		return
	}

	c.JSON(200, gin.H{"message": "注册成功"})
}

// 登录逻辑
func Login(c *gin.Context) {
	userCollection := config.DB.Collection("users")

	// 只接收用户名和密码
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"code": 40001, "error": "参数错误"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"username": input.Username}).Decode(&user)
	if err != nil {
		c.JSON(400, gin.H{"code": 40002, "error": "用户名或密码错误"})
		return
	}

	// 校验密码
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)) != nil {
		c.JSON(400, gin.H{"code": 40002, "error": "用户名或密码错误"})
		return
	}

	token, err := utils.GenerateToken(user.Username)
	if err != nil {
		c.JSON(500, gin.H{"code": 50002, "error": "生成token失败"})
		return
	}

	// 记录最后一次登录的IP
	loginIP := c.ClientIP()
	loginTime := time.Now()
	userCollection.UpdateOne(
		ctx,
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{
			"last_login_ip":   loginIP,
			"last_login_time": loginTime,
		}},
	)

	c.JSON(200, gin.H{"code": 20000, "message": "登录成功", "token": token})
}
