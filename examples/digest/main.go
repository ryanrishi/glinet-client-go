package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"github.com/ryanrishi/glinet-client-go"
	"log"
	"net/http"
)

type request struct {
	Username string
}

type response struct {
	Salt      string
	Algorithm string
	Nonce     string
}

func challenge(req *request, res *response) error {
	log.Print("Got request: ", req)

	res.Salt = "salt"
	res.Algorithm = "algorithm"
	res.Nonce = "nonce"

	return nil
}

func startServer() {
	server := rpc.NewServer()
	server.RegisterCodec(json.NewCodec(), "application/json")
	server.RegisterService(challenge, "")

	router := mux.NewRouter()
	router.Handle("/rpc", server)
	http.ListenAndServe(":8222", router)
	log.Print("Server started on :8222")
	//http.Handle("/rpc", )
	//server.HandleHTTP("/rpc", "/debug/rpc")
	//server.Register(challenge)
	//server.

	//l, err := net.Listen("tcp", ":8222")
	//if err != nil {
	//	log.Fatal("listen error: ", err)
	//}
	//
	//for {
	//	conn, err := l.Accept()
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	//go jsonrpc.ServeConn(conn)
	//
	//	go server.ServeCodec(jsonrpc.NewServerCodec(conn))
	//}
}

func main() {
	//conn, err := .Dial("tcp", "192.168.8.1")
	//if err != nil {
	//	panic(err)
	//}
	//
	//defer conn.Close()
	//
	//c := jsonrpc.NewClient(conn)

	go startServer()

	c := glinet.NewClientWithAddress("localhost:8222")
	challenge, err := c.Digest.Challenge(context.Background(), "admin")
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v\n", err).Error())
		return
	}

	fmt.Printf("Challenge: %v\n", challenge)
}
