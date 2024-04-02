//go:build wireinject

package main

import (
	"github.com/MuxiKeStack/be-user/grpc"
	"github.com/MuxiKeStack/be-user/ioc"
	"github.com/MuxiKeStack/be-user/pkg/grpcx"
	"github.com/MuxiKeStack/be-user/repository"
	"github.com/MuxiKeStack/be-user/repository/dao"
	"github.com/MuxiKeStack/be-user/service"
	"github.com/google/wire"
)

func InitGRPCServer() grpcx.Server {
	wire.Build(
		ioc.InitGRPCxKratosServer,
		grpc.NewUserServiceServer,
		service.NewUserService,
		service.NewCCNUService,
		repository.NewUserRepository,
		dao.NewGORMUserDAO,
		// 第三方
		ioc.InitEtcdClient,
		ioc.InitDB,
		ioc.InitLogger,
	)

	return grpcx.Server(nil)
}
