package main

import (
	"fmt"
	"os"

	"github.com/go-redis/redis"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	if _, err := rdb.Ping().Result(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Project setup success")
}
