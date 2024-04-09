//go:build wireinject

package main

import (
	"github.com/MuxiKeStack/be-evaluation/grpc"
	"github.com/MuxiKeStack/be-evaluation/ioc"
	"github.com/MuxiKeStack/be-evaluation/pkg/grpcx"
	"github.com/google/wire"
)

func InitGRPCServer() grpcx.Server {
	wire.Build(
		ioc.InitGRPCxKratosServer,
		grpc.NewEvaluationServiceServer,
		ioc.InitEtcdClient,
		ioc.InitLogger,
	)
	return grpcx.Server(nil)
}
