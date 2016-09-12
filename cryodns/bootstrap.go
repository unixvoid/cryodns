package main

import (
	"crypto/rand"
	"fmt"

	"github.com/unixvoid/glogger"
	"golang.org/x/crypto/sha3"
	"gopkg.in/redis.v4"
)

func randStr(strSize int) string {
	dictionary := config.Cryo.SecDictionary

	var bytes = make([]byte, strSize)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}

	return string(bytes)
}

func bootstrapCheck(client *redis.Client) {
	// check if instance is already registered
	_, err := client.Get("sec").Result()
	if err != redis.Nil {
		glogger.Debug.Println("instance already registered while bootstrapping")
		return
	} else {
		// instance is not registered, generate key
		sec := randStr(config.Cryo.SecTokenSize)
		secHash := sha3.Sum512([]byte(sec))

		// upload sec key to server
		err := client.Set("sec", fmt.Sprintf("%x", secHash), 0).Err()
		if err != nil {
			// cannot update sec key
			glogger.Error.Println("error in setting sec key in redis while bootstrapping")
			glogger.Error.Println(err)
			return
		} else {
			//return security token
			glogger.Info.Println("environment bootstrapped with: " + sec)
		}
	}
}
