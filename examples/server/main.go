package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"log"
	"net/http"
)

type Arith int

type Args struct {
	A, B int
}

type Result int

func (t *Arith) Multiply(r *http.Request, args *Args, result *Result) error {
	log.Printf("Multiplying %d with %d\n", args.A, args.B)
	*result = Result(args.A * args.B)
	return nil
}

func main() {
	server := rpc.NewServer()
	server.RegisterCodec(json.NewCodec(), "application/json")
	arith := new(Arith)
	server.RegisterService(arith, "")

	router := mux.NewRouter()
	router.Handle("/rpc", server)
	http.ListenAndServe(":8222", router)
	log.Print("Server started on :8222")
}
