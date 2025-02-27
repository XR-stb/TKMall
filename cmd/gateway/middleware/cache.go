package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func CacheMiddleware(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age="+strconv.Itoa(int(duration.Seconds())))
		c.Header("Expires", time.Now().Add(duration).UTC().Format(time.RFC1123))
		c.Next()
	}
} 