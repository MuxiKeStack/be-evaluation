package grpc

import (
	"context"
	evaluationv1 "github.com/MuxiKeStack/be-api/gen/proto/evaluation/v1"
	"github.com/MuxiKeStack/be-evaluation/domain"
	"github.com/MuxiKeStack/be-evaluation/service"
	"google.golang.org/grpc"
)

type EvaluationServiceServer struct {
	evaluationv1.UnimplementedEvaluationServiceServer
	svc service.EvaluationService
}

func NewEvaluationServiceServer(svc service.EvaluationService) *EvaluationServiceServer {
	return &EvaluationServiceServer{svc: svc}
}

func (s *EvaluationServiceServer) Register(server grpc.ServiceRegistrar) {
	evaluationv1.RegisterEvaluationServiceServer(server, s)
}

func (s *EvaluationServiceServer) Evaluated(ctx context.Context, request *evaluationv1.EvaluatedRequest) (*evaluationv1.EvaluatedResponse, error) {
	evaluated, err := s.svc.Evaluated(ctx, request.GetPublisherId(), request.GetCourseId())
	return &evaluationv1.EvaluatedResponse{
		Evaluated: evaluated,
	}, err
}

func (s *EvaluationServiceServer) Publish(ctx context.Context, request *evaluationv1.PublishRequest) (*evaluationv1.PublishResponse, error) {
	id, err := s.svc.Publish(ctx, convertDomain(request.GetEvaluation()))
	return &evaluationv1.PublishResponse{EvaluationId: id}, err
}

func convertDomain(e *evaluationv1.Evaluation) domain.Evaluation {
	return domain.Evaluation{
		Id:          e.Id,
		PublisherId: e.PublisherId,
		CourseId:    e.CourseId,
		StarRating:  uint8(e.StarRating),
		Content:     e.Content,
	}
}
