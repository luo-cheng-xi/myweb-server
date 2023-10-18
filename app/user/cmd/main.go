package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"myweb/app/api/pb"
	"myweb/app/user/internal/server"
	"net"
)

// 运行在50051端口
var (
	port = flag.Int("port", 50051, "The server port")
)

// user server main函数
func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %#v", err)
	}

	// 注册grpc服务
	s := grpc.NewServer()
	userServer, err := server.InitUserServer()
	if err != nil {
		log.Fatalf("failed to init user server %#v", err)
	}
	pb.RegisterUserServer(s, userServer)
	log.Printf("server listening at %#v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
}
