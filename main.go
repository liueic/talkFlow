package main

import (
	"talkFlow/api"
	"talkFlow/config"
	"talkFlow/controllers"
	"talkFlow/middleware" // JWT中间件

	"github.com/gin-gonic/gin"
)

func main() {
	config.InitEnv()
	config.InitMongoDB()

	r := gin.Default()

	r.POST("/api/v1/auth/register", controllers.Register)
	r.POST("/api/v1/auth/login", controllers.Login)

	// 获取用户信息
	r.GET("/api/v1/profile", middleware.JWTAuth(), api.GetProfile)

	r.Run(":8080")
}
