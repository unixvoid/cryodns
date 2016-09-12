package main

import (
	"fmt"
	"net/http"

	"github.com/unixvoid/glogger"
	"golang.org/x/crypto/sha3"
	"gopkg.in/redis.v4"
)

func rotate(w http.ResponseWriter, r *http.Request, client *redis.Client) {
	sec := r.FormValue("sec")
	if len(sec) == 0 {
		glogger.Debug.Println("sec not set")
		w.WriteHeader(http.StatusBadRequest)
		return

	}
	secHash := sha3.Sum512([]byte(sec))

	// check if instance is already registered
	storedSecHash, err := client.Get("sec").Result()
	if err != redis.Nil {
		// sec exists, check auth
		if fmt.Sprintf("%x", secHash) == storedSecHash {
			// client is authed
			newSec := randStr(config.Cryo.SecTokenSize)
			newSecHash := sha3.Sum512([]byte(newSec))

			// upload sec key to server
			err := client.Set("sec", fmt.Sprintf("%x", newSecHash), 0).Err()
			if err != nil {
				// cannot update sec key
				glogger.Error.Println("error in setting sec key in redis during rotate")
				glogger.Error.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				//return security token to client
				w.Header().Set("secToken", newSec)
				fmt.Fprintf(w, "%s", newSec)
			}
		} else {
			// client auth failed
			glogger.Debug.Println("client auth failed")
			w.WriteHeader(http.StatusForbidden)
			return

		}
	} else {
		// instance has not been registered yet
		glogger.Debug.Println("sec not set while rotating")
		w.WriteHeader(http.StatusBadRequest)
		return

	}
}
