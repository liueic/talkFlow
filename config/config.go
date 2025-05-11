package config

import (
	"log"

	"github.com/joho/godotenv"
)

func InitEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("无法正常读取环境变量文件", err)
	}
}
