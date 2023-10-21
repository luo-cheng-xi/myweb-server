package main

import (
	"context"
	"flag"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"log"
	"myweb/app/api/pb"
	"myweb/app/user/internal/server"
	"net"
)

// 运行在50051端口
var (
	host = flag.String("host", "127.0.0.1", "The server host")
	port = flag.String("port", "50051", "The server port")
)

// user server main函数
func main() {
	// 启动gRPC服务
	flag.Parse()
	lis, err := net.Listen("tcp", net.JoinHostPort(*host, *port))
	if err != nil {
		log.Printf("host %v port %v", *host, *port)
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

	// 注册gRPC服务到etcd服务中
	// 创建etcd客户端
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:12379"},
	})
	if err != nil {
		log.Fatalf("Failed to create etcd client: %v", err)
	}
	defer etcdClient.Close()

	// 创建租约
	leaseGrant, err := etcdClient.Grant(context.Background(), 10) // 10秒租约
	if err != nil {
		log.Fatalf("Failed to grant lease: %v", err)
	}

	serviceAddr := net.JoinHostPort(*host, *port)
	serviceKey := "/services/myweb/user/" + serviceAddr
	_, err = etcdClient.Put(context.Background(), serviceKey, serviceAddr, clientv3.WithLease(leaseGrant.ID))
	if err != nil {
		log.Fatalf("Failed to register service to etcd %v", err)
	}

	//防止租约过期，定期续约
	keepAlive, err := etcdClient.KeepAlive(context.Background(), leaseGrant.ID)
	if err != nil {
		return
	}
	go func() {
		for {
			select {
			case _, ok := <-keepAlive:
				if !ok {
					fmt.Println("keep alive channel closed")
				}
				//fmt.Printf("Received keep-alive response: ID=%d ,TTL=%d\n", ka.ID, ka.TTL)
			}
		}
	}()
	// 开启服务
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
}
