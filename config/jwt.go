package config

import "os"

var JWTSecret = []byte(os.Getenv("JWT_SECRET")) // JWT密钥
