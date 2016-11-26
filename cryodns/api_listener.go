package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/unixvoid/glogger"
)

func apiListener() {
	// async listener gets its own redis connection
	redisClient, _ := initRedisConnection()

	// format the string to be :port
	port := fmt.Sprint(":", config.Cryo.APIPort)
	glogger.Info.Println("started API listener on port", config.Cryo.APIPort)

	router := mux.NewRouter()
	router.HandleFunc("/dns", func(w http.ResponseWriter, r *http.Request) {
		listEntries(w, r, redisClient)
	}).Methods("GET")
	router.HandleFunc("/dns", func(w http.ResponseWriter, r *http.Request) {
		addDNS(w, r, redisClient)
	}).Methods("POST")
	router.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		rotate(w, r, redisClient)
	}).Methods("GET")
	router.HandleFunc("/remove", func(w http.ResponseWriter, r *http.Request) {
		removeDNS(w, r, redisClient)
	}).Methods("POST")
	router.HandleFunc("/rotate", func(w http.ResponseWriter, r *http.Request) {
		rotate(w, r, redisClient)
	}).Methods("POST")
	//log.Fatal(http.ListenAndServe(port, router))
	glogger.Error.Println(http.ListenAndServe(port, router))
}
