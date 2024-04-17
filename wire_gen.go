// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/MuxiKeStack/be-evaluation/grpc"
	"github.com/MuxiKeStack/be-evaluation/ioc"
	"github.com/MuxiKeStack/be-evaluation/pkg/grpcx"
	"github.com/MuxiKeStack/be-evaluation/repository"
	"github.com/MuxiKeStack/be-evaluation/repository/dao"
	"github.com/MuxiKeStack/be-evaluation/service"
)

// Injectors from wire.go:

func InitGRPCServer() grpcx.Server {
	logger := ioc.InitLogger()
	db := ioc.InitDB(logger)
	evaluationDAO := dao.NewGORMEvaluationDAO(db)
	evaluationRepository := repository.NewEvaluationRepository(evaluationDAO)
	client := ioc.InitEtcdClient()
	courseServiceClient := ioc.InitCourseClient(client)
	evaluationService := service.NewEvaluationService(evaluationRepository, courseServiceClient)
	evaluationServiceServer := grpc.NewEvaluationServiceServer(evaluationService)
	server := ioc.InitGRPCxKratosServer(evaluationServiceServer, client, logger)
	return server
}
