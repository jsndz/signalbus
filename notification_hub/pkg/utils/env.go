package utils

import "os"

func GetEnv(key string) string {
	val := os.Getenv(key)
	return val
}