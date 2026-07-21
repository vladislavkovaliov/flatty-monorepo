package secure

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
)

func SecurityMiddleware() gin.HandlerFunc {
	secureMw := secure.New(secure.Options{
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		ContentSecurityPolicy: "default-src 'self'",
	})

	return func(c *gin.Context) {
		if err := secureMw.Process(c.Writer, c.Request); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Next()
	}
}
