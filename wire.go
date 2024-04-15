//go:build wireinject

package main

import (
	"github.com/MuxiKeStack/be-evaluation/grpc"
	"github.com/MuxiKeStack/be-evaluation/ioc"
	"github.com/MuxiKeStack/be-evaluation/pkg/grpcx"
	"github.com/MuxiKeStack/be-evaluation/repository"
	"github.com/MuxiKeStack/be-evaluation/repository/dao"
	"github.com/MuxiKeStack/be-evaluation/service"
	"github.com/google/wire"
)

func InitGRPCServer() grpcx.Server {
	wire.Build(
		ioc.InitGRPCxKratosServer,
		grpc.NewEvaluationServiceServer,
		service.NewEvaluationService,
		repository.NewEvaluationRepository,
		dao.NewGORMEvaluationDAO,
		ioc.InitDB,
		ioc.InitEtcdClient,
		ioc.InitLogger,
	)
	return grpcx.Server(nil)
}
