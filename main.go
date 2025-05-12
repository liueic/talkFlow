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

	// 创建房间
	r.POST("/api/v1/room/create", middleware.JWTAuth(), api.CreateRoom)
	// 加入房间
	r.POST("/api/v1/room/join", api.JoinRoom)

	// ws
	r.GET("/api/v1/ws", api.TalkHandler)
	// 清除僵尸房间
	api.StartRoomCleaner()

	// 测试页面
	r.StaticFile("/chat.html", "./test/chat.html")

	r.Run(":8080")
}
