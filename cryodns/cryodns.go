package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/miekg/dns"
	"github.com/unixvoid/glogger"
	"gopkg.in/gcfg.v1"
	"gopkg.in/redis.v4"
)

type Config struct {
	Cryo struct {
		Loglevel string
		DNSPort  int
		APIPort  int
		Ttl      uint32
	}
	Upstream struct {
		Server string
	}
	Redis struct {
		Host     string
		Password string
	}
}

var (
	config = Config{}
)

func main() {
	readConf()
	initLogger(config.Cryo.Loglevel)
	redisClient, redisErr := initRedisConnection()
	if redisErr != nil {
		glogger.Error.Println("redis connection cannot be made.")
		glogger.Error.Println("cryodns will continue to function in passthrough mode only")
	} else {
		glogger.Debug.Println("connection to redis succeeded.")
	}

	// format the string to be :port
	fPort := fmt.Sprint(":", config.Cryo.DNSPort)

	udpServer := &dns.Server{Addr: fPort, Net: "udp"}
	tcpServer := &dns.Server{Addr: fPort, Net: "tcp"}
	glogger.Info.Println("started server on", config.Cryo.DNSPort)
	go apiListener()
	dns.HandleFunc(".", func(w dns.ResponseWriter, req *dns.Msg) {

		switch req.Question[0].Qtype {
		case 1:
			glogger.Debug.Println("'A' request recieved, continuing")
			//go anameresolve(w, req, redisClient)
			go anameresolve(w, req, redisClient)
			break
		case 5:
			glogger.Debug.Println("Routing 'CNAME' request")
			go cnameresolve(w, req, redisClient)
			break
		case 28:
			glogger.Debug.Println("Routing 'AAAA' request")
			go aaaaresolve(w, req, redisClient)
			break
		default:
			glogger.Debug.Println("Not 'A' request")
			break
		}

	})

	go func() {
		glogger.Error.Println(udpServer.ListenAndServe())
	}()
	glogger.Error.Println(tcpServer.ListenAndServe())
}

func readConf() {
	// init config file
	err := gcfg.ReadFileInto(&config, "config.gcfg")
	if err != nil {
		panic(fmt.Sprintf("Could not load config.gcfg, error: %s\n", err))
	}
}

func initLogger(logLevel string) {
	// init logger
	if logLevel == "debug" {
		glogger.LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	} else if logLevel == "cluster" {
		glogger.LogInit(os.Stdout, os.Stdout, ioutil.Discard, os.Stderr)
	} else if logLevel == "info" {
		glogger.LogInit(os.Stdout, ioutil.Discard, ioutil.Discard, os.Stderr)
	} else {
		glogger.LogInit(ioutil.Discard, ioutil.Discard, ioutil.Discard, os.Stderr)
	}
}

func initRedisConnection() (*redis.Client, error) {
	// init redis connection
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Host,
		Password: config.Redis.Password,
		DB:       0,
	})

	_, redisErr := redisClient.Ping().Result()
	return redisClient, redisErr
}

func upstreamQuery(w dns.ResponseWriter, req *dns.Msg) *dns.Msg {
	transport := "udp"
	if _, ok := w.RemoteAddr().(*net.TCPAddr); ok {
		transport = "tcp"
	}
	c := &dns.Client{Net: transport}
	resp, _, err := c.Exchange(req, config.Upstream.Server)

	if err != nil {
		glogger.Debug.Println(err)
		dns.HandleFailed(w, req)
	}
	return resp

	// call main builder to craft and send the response
	//mainBuilder(w, req, resp, clusterString, redisClient)
}

func anameresolve(w dns.ResponseWriter, req *dns.Msg, redisClient *redis.Client) {
	hostname := req.Question[0].Name
	glogger.Cluster.Println(hostname)

	// check redis for entry
	res, err := checkRecord(hostname, "a", redisClient)
	if err != nil {
		// we dont have it in local records, send upstream
		glogger.Debug.Println("entry not found in records, sending upstream")
		req = upstreamQuery(w, req)
		w.WriteMsg(req)
		return
	}

	// craft response
	rr := new(dns.A)
	rr.Hdr = dns.RR_Header{Name: hostname, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: config.Cryo.Ttl}
	addr := strings.TrimSuffix(res, "\n")
	rr.A = net.ParseIP(addr)

	// craft reply
	rep := new(dns.Msg)
	rep.SetReply(req)
	rep.Answer = append(rep.Answer, rr)

	// send it
	w.WriteMsg(rep)
	return
}

func cnameresolve(w dns.ResponseWriter, req *dns.Msg, redisClient *redis.Client) {
	hostname := req.Question[0].Name
	glogger.Cluster.Println(hostname)

	// check redis for entry
	res, err := checkRecord(hostname, "cname", redisClient)
	if err != nil {
		// we dont have it in local records, send upstream
		glogger.Debug.Println("entry not found in records, sending upstream")
		req = upstreamQuery(w, req)
		w.WriteMsg(req)
		return
	}

	// craft response
	rr := new(dns.CNAME)
	rr.Hdr = dns.RR_Header{Name: hostname, Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: config.Cryo.Ttl}
	// return default cname
	rr.Target = res

	// craft reply
	rep := new(dns.Msg)
	rep.SetReply(req)
	rep.Answer = append(rep.Answer, rr)

	// send it
	w.WriteMsg(rep)
	return
}

func aaaaresolve(w dns.ResponseWriter, req *dns.Msg, redisClient *redis.Client) {
	hostname := req.Question[0].Name
	glogger.Cluster.Println(hostname)

	// check redis for entry
	res, err := checkRecord(hostname, "aaaa", redisClient)
	if err != nil {
		// we dont have it in local records, send upstream
		glogger.Debug.Println("entry not found in records, sending upstream")
		req = upstreamQuery(w, req)
		w.WriteMsg(req)
		return
	}

	// craft response
	rr := new(dns.AAAA)
	rr.Hdr = dns.RR_Header{Name: hostname, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: config.Cryo.Ttl}
	rr.AAAA = net.ParseIP(res)

	// craft reply
	rep := new(dns.Msg)
	rep.SetReply(req)
	rep.Answer = append(rep.Answer, rr)

	// send it
	w.WriteMsg(rep)
	return
}

func checkRecord(lookup, requestType string, redisClient *redis.Client) (string, error) {
	glogger.Debug.Println("querying", lookup)
	res, redisErr := redisClient.Get(fmt.Sprintf("dns:%s:%s", requestType, lookup)).Result()
	glogger.Debug.Println("returned value is:", res)
	return res, redisErr
}
