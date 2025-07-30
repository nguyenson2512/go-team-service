package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

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

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		// Check token expiration
		if exp, ok := claims["exp"].(float64); !ok || int64(exp) < time.Now().Unix() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
			return
		}

		// Check required claims
		userId, ok1 := claims["userId"].(string)
		role, ok2 := claims["role"].(string)
		if !ok1 || !ok2 || userId == "" || role == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid user claims"})
			return
		}
		c.Set("userId", userId)
		c.Set("role", role)
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
			isManager, err := IsUserManagerOfTeam(db, userId, teamId)
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

// Helper function to check if user is manager of team
func IsUserManagerOfTeam(db *gorm.DB, userId string, teamId string) (bool, error) {
	var count int64
	err := db.Table("rosters").Where(`"userId" = ? AND "teamId" = ? AND "isLeader" = TRUE`, userId, teamId).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
