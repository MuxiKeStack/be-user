package ioc

import (
	"context"
	ccnuv1 "github.com/MuxiKeStack/be-api/gen/proto/ccnu"
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

func InitCCNUClient(etcdClient *etcdv3.Client) ccnuv1.CCNUServiceClient {
	type Config struct {
		Endpoint string `yaml:"endpoint"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.ccnu", &cfg)
	if err != nil {
		panic(err)
	}
	r := etcd.New(etcdClient)
	cc, err := grpc.DialInsecure(context.Background(),
		grpc.WithEndpoint(cfg.Endpoint),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		panic(err)
	}
	client := ccnuv1.NewCCNUServiceClient(cc)
	return client
}
