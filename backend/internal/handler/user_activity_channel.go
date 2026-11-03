package handler

import "github.com/gin-gonic/gin"

func userActivityChannel(c *gin.Context) string {
	channel, _ := c.Get("auth_channel")
	if channel == "miniapp" {
		return "miniapp"
	}
	return "web"
}
