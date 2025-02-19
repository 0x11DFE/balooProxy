package firewall

import (
	"fmt"
	"goProxy/core/proxy"
	"net"
	"net/http"
	"strings"
	"sync"
)

var (
	Mutex = &sync.Mutex{}
	//store fingerprint requests for ratelimiting
	TcpRequests = map[string]int{}

	//store unknown fingerprints for ratelimiting
	UnkFps = map[string]int{}

	//store bypassing ips for ratelimiting
	AccessIps = map[string]int{}

	//store ips that didnt have verification cookie set for ratelimiting
	AccessIpsCookie = map[string]int{}

	//"cache" encryption result of ips for 2 minutes in order to have less load on the proxy
	CacheIps = map[string]string{}

	//"cache" captcha images to for 2 minutes in order to have less load on the proxy
	CacheImgs = map[string]string{}

	Connections = map[string]string{}
)

func OnStateChange(conn net.Conn, state http.ConnState) {

	ip := strings.Split(conn.RemoteAddr().String(), ":")[0]

	switch state {
	case http.StateNew:
		Mutex.Lock()
		fpReq := TcpRequests[ip]
		successCount := AccessIps[ip]
		challengeCount := AccessIpsCookie[ip]
		TcpRequests[ip] = TcpRequests[ip] + 1
		Mutex.Unlock()

		//We can ratelimit so extremely here because normal browsers will send actual webrequests instead of only establishing connections
		if (fpReq > proxy.FailRequestRatelimit && (successCount < 1 && challengeCount < 1)) || fpReq > 500 {
			defer conn.Close()
			return
		}
	case http.StateHijacked, http.StateClosed:
		//Remove connection from list of fingerprints as it's no longer needed
		Mutex.Lock()
		delete(Connections, fmt.Sprint(conn.RemoteAddr()))
		Mutex.Unlock()
	}
}
