package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func main() {
	r := gin.Default()
	SetMiddleWare(r)
	SetRpc(r)
	InitRouter(r)
	s := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := s.ListenAndServe()
	if err != nil {
		log.Print(err)
	}
}
