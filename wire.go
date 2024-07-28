//go:build wireinject

package main

import (
	"github.com/MuxiKeStack/be-evaluation/grpc"
	"github.com/MuxiKeStack/be-evaluation/ioc"
	"github.com/MuxiKeStack/be-evaluation/pkg/grpcx"
	"github.com/MuxiKeStack/be-evaluation/repository"
	"github.com/MuxiKeStack/be-evaluation/repository/cache"
	"github.com/MuxiKeStack/be-evaluation/repository/dao"
	"github.com/MuxiKeStack/be-evaluation/service"
	"github.com/google/wire"
)

func InitGRPCServer() grpcx.Server {
	wire.Build(
		ioc.InitGRPCxKratosServer,
		grpc.NewEvaluationServiceServer,
		service.NewEvaluationService,
		ioc.InitCourseClient,
		repository.NewEvaluationRepository,
		cache.NewRedisEvaluationCache,
		dao.NewGORMEvaluationDAO,
		ioc.InitRedis,
		ioc.InitDB,
		ioc.InitLimiter,
		ioc.InitEtcdClient,
		ioc.InitLogger,
	)
	return grpcx.Server(nil)
}
