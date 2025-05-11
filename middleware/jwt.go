package middleware

import (
	"net/http"
	"talkFlow/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 40001, "error": "未提供 token"})
			c.Abort()
			return
		}

		token, err := utils.ParseToken(tokenStr)
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 40002, "error": "无效 token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 40003, "error": "无效 token"})
			c.Abort()
			return
		}

		c.Set("username", claims["username"])
		c.Next()
	}
}
