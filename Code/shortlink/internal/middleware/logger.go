package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.Request.URL.Path
		fmt.Println("[IN ]", method, path)
		c.Next()
		fmt.Println("[OUT]", method, path, "->", c.Writer.Status(), "X-Cache=", c.Writer.Header().Get("X-Cache"))
	}
}
