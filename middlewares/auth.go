package middlewares

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"team-service/utils"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			return
		}

		// Bearer token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization format"})
			return
		}

		tokenStr := tokenParts[1]

		// Parse token
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("ACCESS_TOKEN_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("userId", claims["userId"])
			c.Set("role", claims["role"])
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		c.Next()
	}
}

func RequireNotMember() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role == "MEMBER" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Members are not allowed to perform this action",
			})
			return
		}
		c.Next()
	}
}

func RequireManagerOfTeam() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.GetString("userId")
		role := c.GetString("role")
		teamId := c.Param("teamId")

		if role == "MEMBER" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Members cannot manage teams",
			})
			return
		}
		db := c.MustGet("db").(*gorm.DB)

		if role == "MANAGER" {
			isManager, err := utils.IsUserManagerOfTeam(db, userId, teamId)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to verify team access",
				})
				return
			}
			if !isManager {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "You are not a manager of this team",
				})
				return
			}
		}

		c.Next()
	}
}
