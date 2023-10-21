package rpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"log"
	"myweb/app/api/pb"
	"time"
)

type UserConn struct {
	Conn *grpc.ClientConn
}

// NewUserConn 创建一个UserConn对象
func NewUserConn() *UserConn {
	// 连接
	userConn, err := grpc.Dial(
		"custom:////services/myweb/user",                         // etcd服务前缀
		grpc.WithTransportCredentials(insecure.NewCredentials()), //暂时不使用安全证书
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    10 * time.Second, // 定期发送心跳以保持连接
			Timeout: 5 * time.Second,  // 超时时间,
		}),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`)) //配置轮询策略
	if err != nil {
		log.Fatalf("create client error %#v", err)
	}
	return &UserConn{
		userConn,
	}
}

func (u UserConn) NewUserClient() pb.UserClient {
	return pb.NewUserClient(u.Conn)
}
