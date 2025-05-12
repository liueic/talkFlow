package config

import (
	"log"

	"github.com/joho/godotenv"
)

func InitEnv() {
	// 尝试加载 .env 文件，不存在则使用环境变量
	err := godotenv.Load()
	if err != nil {
		log.Println("未找到 .env 文件，跳过加载，直接使用环境变量")
	}
}
