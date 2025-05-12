package controllers

import (
	"context"
	"crypto/md5" // #nosec G401 -- Gravatar requires MD5
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"talkFlow/config"
	"talkFlow/models"
	"talkFlow/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Register handles user registration
func Register(c *gin.Context) {
	// 只接受 Username, Password, Email
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"code": 40001, "error": "参数错误"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 检查用户名是否存在
	var existingUserID int
	checkUserSQL := "SELECT id FROM register WHERE username = ?"
	err := config.DB.QueryRowContext(ctx, checkUserSQL, input.Username).Scan(&existingUserID)

	if err != nil {
		if err == sql.ErrNoRows {
			// 当用户名不存在的时候则创建新用户
		} else {
			// An error occurred during the query
			c.JSON(500, gin.H{"code": 50000, "error": "数据库查询失败"})
			log.Printf("Error checking if username exists: %v", err)
			return
		}
	} else {
		// 用户已存在就返回错误
		c.JSON(400, gin.H{"code": 40003, "error": "用户名已存在"})
		return
	}

	// 生成头像
	email := strings.TrimSpace(strings.ToLower(input.Email))
	hash := md5.Sum([]byte(email)) // #nosec G401 -- Gravatar requires MD5
	avatar := fmt.Sprintf("https://www.gravatar.com/avatar/%x", hash)

	// 加密用户密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"code": 50004, "error": "密码加密失败"})
		log.Printf("Error hashing password: %v", err)
		return
	}

	// Construct the User object
	user := models.Register{
		Username:   input.Username,
		Password:   string(hashedPassword),
		Email:      input.Email,
		Avatar:     avatar,
		CreatedAt:  time.Now(),
		RegisterIP: c.ClientIP(),
		IsRegister: true,
	}

	// Insert the new user into the database
	insertSQL := `
		INSERT INTO register (username, password, email, avatar, created_at, register_ip, is_register)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err = config.DB.ExecContext(ctx, insertSQL,
		user.Username, user.Password, user.Email, user.Avatar,
		user.CreatedAt, user.RegisterIP, user.IsRegister,
	)

	if err != nil {
		c.JSON(500, gin.H{"code": 50001, "error": "注册失败"})
		utils.Logger(user.Username, fmt.Sprintf("Register DB insert error: %v", err), time.Now().Format(time.RFC3339), c.ClientIP())
		log.Printf("Error inserting new user: %v", err)
		return
	}

	c.JSON(200, gin.H{"code": 20000, "message": "注册成功"})
}

// Login handles user login
func Login(c *gin.Context) {
	// Only accept Username and Password
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

	var user models.Register
	query := "SELECT id, username, password, email, avatar, created_at, register_ip, is_register, last_login_ip, last_login_time FROM register WHERE username = ?"
	row := config.DB.QueryRowContext(ctx, query, input.Username)
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.Avatar,
		&user.CreatedAt,
		&user.RegisterIP,
		&user.IsRegister,
		&user.LastLoginIP,
		&user.LastLoginTime,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(400, gin.H{"code": 40002, "error": "用户名或密码错误"})
		} else {
			c.JSON(500, gin.H{"code": 50000, "error": "数据库查询失败"})
			utils.Logger(input.Username, fmt.Sprintf("Login DB scan error: %v", err), time.Now().Format(time.RFC3339), c.ClientIP())
		}
		return
	}

	// 比对密码
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)) != nil {
		c.JSON(400, gin.H{"code": 40002, "error": "用户名或密码错误"})
		return
	}

	// 生成 Auth Token
	token, errToken := utils.GenerateToken(user.Username)
	if errToken != nil {
		c.JSON(500, gin.H{"code": 50002, "error": "生成token失败"})
		utils.Logger(user.Username, errToken.Error(), time.Now().Format(time.RFC3339), c.ClientIP())
		return
	}

	// 更新每一次登录的 IP 和时间
	loginIP := c.ClientIP()
	loginTime := time.Now()
	updateQuery := "UPDATE register SET last_login_ip = ?, last_login_time = ? WHERE id = ?"
	_, errDbUpdate := config.DB.ExecContext(ctx, updateQuery, loginIP, loginTime, user.ID)
	if errDbUpdate != nil {
		c.JSON(500, gin.H{"code": 50003, "error": "更新登录信息失败"})
		utils.Logger(user.Username, errDbUpdate.Error(), time.Now().Format(time.RFC3339), c.ClientIP())
		log.Printf("更新登录信息失败: %v", errDbUpdate)
		return
	}

	c.JSON(200, gin.H{"code": 20000, "message": "登录成功", "token": token})
}

// NullTime is a helper type for scanning nullable time.Time values from the database
type NullTime sql.NullTime

// Scan implements the sql.Scanner interface for NullTime
func (nt *NullTime) Scan(value interface{}) error {
	var t sql.NullTime
	if err := t.Scan(value); err != nil {
		return err
	}
	*nt = NullTime(t)
	return nil
}
