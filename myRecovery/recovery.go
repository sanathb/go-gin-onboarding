package myRecovery

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered from panic")
			}
		}()
		c.Next()
	}

}
