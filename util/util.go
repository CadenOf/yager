package util

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/teris-io/shortid"
)

func GenShortId() (string, error) {
	return shortid.Generate()
}

func GetReqID(c *gin.Context) string {
	v, ok := c.Get("X-Request-Id")
	fmt.Println(v, ok)
	if !ok {
		return ""
	}
	fmt.Println(ok)
	if requestId, ok := v.(string); ok {
		return requestId
	}
	return ""
}
