package core

import (
	"time"

	"github.com/gin-gonic/gin"
)

var (
	Domain          string
	Port            string
	TikTokSessionID string
	YouTubeKey      string
	OpenAIKey       string
	MinGodInterval  time.Duration

	Gin = gin.Default()
)
