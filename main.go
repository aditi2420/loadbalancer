package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type simpleServer struct {
	addr       string
	reverProxy *httputil.ReverseProxy
}

func newSimpleServer(add string) *simpleServer {
	adr, err := url.Parse(add)
	if err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}
	return &simpleServer{
		addr:       add,
		reverProxy: httputil.NewSingleHostReverseProxy(adr),
	}
}

type server interface {
	Address() string
	IsAlive() bool
	serve(resp http.ResponseWriter, rq *http.Request)
}

func (s *simpleServer) Address() string { return s.addr }
func (s *simpleServer) IsAlive() bool   { return true }
func (s *simpleServer) serve(rs http.ResponseWriter, rq * http.Request){
	s.reverProxy.ServeHTTP(rs,rq)
}

type LoadBalncer struct {
	Port       string
	Roundrobin int
	Servers    []server
}

func NewLoadBalancer (port string, server []server) *LoadBalncer{
	return &LoadBalncer{
		Port: port,
		Roundrobin: 0,
		Servers: server,
		
	}
}

func (lb *LoadBalncer) getNextAvailableServer() server {
	server := lb.Servers[lb.Roundrobin%len(lb.Servers)]
	for !server.IsAlive() {
		lb.Roundrobin++
		server = lb.Servers[lb.Roundrobin%len(lb.Servers)]
	}
	lb.Roundrobin++

	return server
}

func (lb *LoadBalncer)serveProxy(rs http.ResponseWriter,rq *http.Request){
	targetServer := lb.getNextAvailableServer()
	fmt.Println("forwardidng reques to ",targetServer.Address())

	targetServer.serve(rs,rq)

}

func main(){
	servers := []server{
		newSimpleServer("https://www.facebook.com"),
		newSimpleServer("https://www.bing.com"),
		newSimpleServer("https://www.duckduckgo.com"),
	}

	lb := NewLoadBalancer("8008",servers)
	handleRedirect := func(rs http.ResponseWriter,rq *http.Request){
		lb.serveProxy(rs,rq)
	}


	http.HandleFunc("/",handleRedirect)
	fmt.Printf("serving requests at 'localhost:%s'\n", lb.Port)
	http.ListenAndServe(":"+lb.Port, nil)

}