package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/unixvoid/glogger"
	"golang.org/x/crypto/sha3"

	"gopkg.in/redis.v4"
)

func removeDNS(w http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	r.ParseForm()

	rmType := strings.ToLower(strings.TrimSpace(r.FormValue("dnstype")))
	rmDomain := strings.TrimSpace(r.FormValue("domain"))
	clientSec := strings.TrimSpace(r.FormValue("sec"))

	if (len(clientSec) == 0) || (len(rmDomain) == 0) {
		glogger.Debug.Println("domain or sec not set, exiting..")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// check auth
	clientSecHash := sha3.Sum512([]byte(clientSec))
	storedSecHash, err := redisClient.Get("sec").Result()
	if err != nil {
		// sec not set, throw 400
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if fmt.Sprintf("%x", clientSecHash) == storedSecHash {
		// authed

		// fully qualify domain if not done already
		if string(rmDomain[len(rmDomain)-1]) != "." {
			rmDomain = fmt.Sprintf("%s.", rmDomain)
		}

		if len(rmType) == 0 {
			// if type not set, nix them all
			glogger.Debug.Printf("removing all dns types for %s", rmDomain)
			redisClient.Del(fmt.Sprintf("dns:a:%s", rmDomain))
			redisClient.Del(fmt.Sprintf("dns:aaaa:%s", rmDomain))
			redisClient.Del(fmt.Sprintf("dns:cname:%s", rmDomain))
			// remove all from index
			redisClient.SRem(fmt.Sprintf("index:dns"), fmt.Sprintf("a:%s", rmDomain))
			redisClient.SRem(fmt.Sprintf("index:dns"), fmt.Sprintf("aaaa:%s", rmDomain))
			redisClient.SRem(fmt.Sprintf("index:dns"), fmt.Sprintf("cname:%s", rmDomain))
		} else {
			// just remove the specific type
			glogger.Debug.Printf("removing %s entry for %s", rmType, rmDomain)
			redisClient.Del(fmt.Sprintf("dns:%s:%s", rmType, rmDomain))
			redisClient.SRem(fmt.Sprintf("index:dns"), fmt.Sprintf("a:%s", rmDomain))

			// remove dns entry from the custom list
			redisClient.SRem("index:dns", fmt.Sprintf("%s:%s", rmType, rmDomain))
		}
		w.WriteHeader(http.StatusOK)
		return
	} else {
		// not authed, throw 403
		w.WriteHeader(http.StatusForbidden)
		return
	}
}
