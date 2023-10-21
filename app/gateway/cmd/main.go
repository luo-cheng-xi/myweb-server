package main

import (
	"github.com/gin-gonic/gin"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
	"log"
	"myweb/app/gateway/rpc"
	"net/http"
	"time"
)

func main() {
	// 创建gin.Engine
	r := gin.Default()

	// 设置中间件
	SetMiddleWare(r)

	// 初始化etcd客户端
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:12379"},
	})
	if err != nil {
		log.Fatalf("failed to create etcd client %v", err)
	}
	defer etcdClient.Close()

	resolver.Register(&rpc.CustomResolverBuilder{
		EtcdClient: etcdClient,
	})

	InitRouter(r)
	s := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err = s.ListenAndServe()
	if err != nil {
		log.Print(err)
	}
}
