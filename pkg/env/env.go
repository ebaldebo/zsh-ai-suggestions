package env

import (
	"os"
	"strconv"
)

func Get(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func GetInt(key, fallback string) int {
	val := Get(key, fallback)
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return intVal
}
