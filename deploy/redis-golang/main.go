package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var ctx context.Context
var client *redis.Client

func main() {
	ctx = context.TODO()
	client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	fmt.Println(client.Ping(ctx))

	var choice int
	var key string
	var val string
	for {
		fmt.Println("1 - set, 2 - get")
		fmt.Scanln(&choice)
		if choice == 1 {
			fmt.Scanln(&key)
			fmt.Scanln(&val)
			setVal(key, val)
		} else if choice == 2 {
			fmt.Scanln(&key)
			fmt.Println(getVal(key))
		}
	}
}

func setVal(key string, val string) {
	err := client.Set(ctx, key, val, 0).Err()
	if err != nil {
		panic(err)
	}
}

func getVal(key string) (val string) {
	val, err := client.Get(ctx, key).Result()
	if err != nil {
		fmt.Println(err)
	}
	return val
}
