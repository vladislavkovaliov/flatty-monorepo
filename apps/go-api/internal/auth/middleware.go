package auth

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AuthMiddleware(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := ""
		log.Printf("AuthMiddleware: path=%s", c.Request.URL.Path)

		if cookie, err := c.Cookie("better-auth.session_token"); err == nil && cookie != "" {
			token = cookie
			log.Println("AuthMiddleware: extracted token from cookie")
		} else if header := c.GetHeader("Authorization"); header != "" {
			if strings.HasPrefix(header, "Bearer ") {
				token = strings.TrimPrefix(header, "Bearer ")
				log.Println("AuthMiddleware: extracted token from Bearer header")
			}
		}

		if token == "" {
			log.Println("AuthMiddleware: no token, rejecting with 401")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if idx := strings.Index(token, "."); idx != -1 {
			token = token[:idx]
			log.Println("AuthMiddleware: stripped compound token")
		}

		var userID string
		err := pool.QueryRow(c.Request.Context(),
			`SELECT "userId" FROM "session" WHERE token = $1 AND "expiresAt" > NOW()`,
			token,
		).Scan(&userID)

		if err != nil {
			log.Printf("AuthMiddleware: DB query failed: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		log.Printf("AuthMiddleware: authenticated user=%s", userID)
		c.Set("userID", userID)
		c.Next()
	}
}
