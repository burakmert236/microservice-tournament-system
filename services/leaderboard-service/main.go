package main

import (
	"github.com/burakmert236/goodswipe-common/utils"
	"github.com/burakmert236/goodswipe-leaderboard-service/cache"
)

func main() {
	cache.Init()
	utils.WaitForGracefulShutdown()
}
