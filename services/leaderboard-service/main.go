package main

import (
	"github.com/burakmert236/goodswipe/internal/utils"
	"github.com/burakmert236/goodswipe/leaderboard-service/cache"
)

func main() {
	repo := cache.Init()
	utils.WaitForGracefulShutdown()
}
