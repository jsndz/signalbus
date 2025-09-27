package database

import (
	"github.com/redis/go-redis/v9"
)


func InitRedis(dsn string) *redis.Client{
    rdb := redis.NewClient(&redis.Options{
        Addr:     dsn,
        Password: "", 
        DB:       0,
    })
	return rdb

}