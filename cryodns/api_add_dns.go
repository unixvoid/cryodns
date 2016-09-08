package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/unixvoid/glogger"
	"golang.org/x/crypto/sha3"
	"gopkg.in/redis.v4"
)

func addDNS(w http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	r.ParseForm()

	dnsType := strings.ToLower(strings.TrimSpace(r.FormValue("dnstype")))
	domain := strings.TrimSpace(r.FormValue("domain"))
	domainValue := strings.TrimSpace(r.FormValue("value"))
	clientSec := strings.TrimSpace(r.FormValue("sec"))

	// make sure domain and value are set
	if (len(clientSec) == 0) || (len(domain) == 0) || (len(domainValue) == 0) {
		glogger.Debug.Println("domain or value not set, exiting..")
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// auth client
		clientSecHash := sha3.Sum512([]byte(clientSec))
		storedSecHash, err := redisClient.Get("sec").Result()
		if err != nil {
			// sec not set,
			// throw 400
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if fmt.Sprintf("%x", clientSecHash) == storedSecHash {
			// authed
			if len(dnsType) == 0 {
				// default to aname entry
				dnsType = "a"
			} else {
				// if dnstype is set, make sure it is something we support
				switch dnsType {
				case
					"a",
					// coming soon
					"aaaa",
					"cname":
					break
				default:
					glogger.Debug.Printf("unsupported dnstype '%s', exiting..\n", dnsType)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}
			if dnsType == "cname" {
				// if we are dealing with a CNAME entry fully qualify it
				if string(domainValue[len(domainValue)-1]) != "." {
					domainValue = fmt.Sprintf("%s.", domainValue)
				}
			}
			// fully qualify the domain name if it is not already:
			if string(domain[len(domain)-1]) != "." {
				domain = fmt.Sprintf("%s.", domain)
			}

			glogger.Debug.Printf("adding domain entry: dns:%s:%s :: %s", dnsType, domain, domainValue)

			// add dns entry dns:<dns_type>:<domain> <domain_value>
			redisClient.Set(fmt.Sprintf("dns:%s:%s", dnsType, domain), domainValue, 0).Err()

			// add dns entry to the list of custom dns names
			redisClient.SAdd("index:dns", fmt.Sprintf("%s:%s", dnsType, domain))

			// return confirmation header to client
			w.Header().Set("x-register", "registered")
			w.WriteHeader(http.StatusOK)
		} else {
			// not authed
			w.WriteHeader(http.StatusForbidden)
		}
	}
}
