package cache

import (
	"log"
)

func Init() *LeaderboardRepo {
	redisConfig := Load()
	redisClient, err := NewRedisClient(*redisConfig)
	if err != nil {
		log.Fatalln(err.Error())
	}

	repo := NewLeaderboardRepo(redisClient.GetClient())
	log.Println("Redis repository initialized.")

	return repo
}
