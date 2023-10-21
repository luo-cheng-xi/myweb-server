package rpc

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
	"log"
	"time"
)

// CustomResolver 实现了 resolver.Resolver 接口
type CustomResolver struct {
	target resolver.Target
	cc     resolver.ClientConn
	client *clientv3.Client
	prefix string
	quit   chan struct{}
}

// NewCustomResolver 创建一个自定义名称解析器
func NewCustomResolver(target resolver.Target, cc resolver.ClientConn, client *clientv3.Client) resolver.Resolver {
	r := &CustomResolver{
		target: target,
		cc:     cc,
		client: client,
		prefix: target.Endpoint(), // 使用服务名称前缀作为 etcd 键前缀
		quit:   make(chan struct{}),
	}
	go r.watchUpdates()
	return r
}

// ResolveNow 实现了重新解析的逻辑
func (r *CustomResolver) ResolveNow(o resolver.ResolveNowOptions) {
	// 在此方法中，你可以手动触发名称解析的重新解析逻辑
	// 这里可以留空，因为我们在 watchUpdates 中已经实现了自动的前缀查询
}

// Close 实现了名称解析器的关闭逻辑
func (r *CustomResolver) Close() {
	close(r.quit)
}

// watchUpdates 监视前缀键的变化，实现前缀查询
func (r *CustomResolver) watchUpdates() {
	for {
		select {
		case <-r.quit:
			return
		default:
			// 使用 etcd 的前缀查询功能获取所有以 prefix 开头的键
			resp, err := r.client.Get(context.Background(), r.prefix, clientv3.WithPrefix())
			if err != nil {
				log.Printf("Failed to get etcd keys: %v", err)
				return
			}

			// 提取服务地址
			var addresses []resolver.Address
			for _, kv := range resp.Kvs {
				addresses = append(addresses, resolver.Address{Addr: string(kv.Value)})
			}

			// 更新 gRPC 客户端连接状态
			r.cc.UpdateState(resolver.State{Addresses: addresses})

			// 定期触发前缀查询，以获取最新的服务地址
			time.Sleep(10 * time.Second)
		}
	}
}

// CustomResolverBuilder 实现了 resolver.Builder 接口
type CustomResolverBuilder struct {
	EtcdClient *clientv3.Client
}

// Build 实现了创建自定义名称解析器的逻辑
func (b *CustomResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	return NewCustomResolver(target, cc, b.EtcdClient), nil
}

// Scheme 返回名称解析器的方案
func (b *CustomResolverBuilder) Scheme() string {
	return "custom"
}

func init() {
	// 注册名称解析器构建器
	resolver.Register(&CustomResolverBuilder{})
}
