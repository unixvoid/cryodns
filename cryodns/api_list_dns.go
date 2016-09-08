package main

import (
	"fmt"
	"net/http"

	"gopkg.in/redis.v4"
)

func listEntries(w http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	hosts, err := redisClient.SInter("index:dns").Result()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	fmt.Fprintln(w, hosts)
}
