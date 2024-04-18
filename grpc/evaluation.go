package grpc

import (
	"context"
	evaluationv1 "github.com/MuxiKeStack/be-api/gen/proto/evaluation/v1"
	"github.com/MuxiKeStack/be-evaluation/domain"
	"github.com/MuxiKeStack/be-evaluation/service"
	"github.com/ecodeclub/ekit/slice"
	"google.golang.org/grpc"
	"math"
)

type EvaluationServiceServer struct {
	evaluationv1.UnimplementedEvaluationServiceServer
	svc service.EvaluationService
}

func (s *EvaluationServiceServer) CompositeScoreCourse(ctx context.Context,
	request *evaluationv1.CompositeScoreCourseRequest) (*evaluationv1.CompositeScoreCourseResponse, error) {
	score, err := s.svc.CompositeScoreCourse(ctx, request.GetCourseId())
	return &evaluationv1.CompositeScoreCourseResponse{Score: score}, err
}

func (s *EvaluationServiceServer) VisiblePublishersCourse(ctx context.Context,
	request *evaluationv1.VisiblePublishersCourseRequest) (*evaluationv1.VisiblePublishersCourseResponse, error) {
	publishers, err := s.svc.VisiblePublishersCourse(ctx, request.GetCourseId())
	return &evaluationv1.VisiblePublishersCourseResponse{
		Publishers: publishers,
	}, err
}

func (s *EvaluationServiceServer) Detail(ctx context.Context, request *evaluationv1.DetailRequest) (*evaluationv1.DetailResponse, error) {
	evaluation, err := s.svc.Detail(ctx, request.GetEvaluationId())
	return &evaluationv1.DetailResponse{Evaluation: convertToV(evaluation)}, err
}

func (s *EvaluationServiceServer) CountCourseInvisible(ctx context.Context,
	request *evaluationv1.CountCourseInvisibleRequest) (*evaluationv1.CountCourseInvisibleResponse, error) {
	count, err := s.svc.CountCourseInvisible(ctx, request.GetCourseId())
	return &evaluationv1.CountCourseInvisibleResponse{Count: count}, err
}

func (s *EvaluationServiceServer) CountMine(ctx context.Context,
	request *evaluationv1.CountMineRequest) (*evaluationv1.CountMineResponse, error) {
	count, err := s.svc.CountMine(ctx, request.GetUid(), request.GetStatus())
	return &evaluationv1.CountMineResponse{Count: count}, err
}

func (s *EvaluationServiceServer) ListCourse(ctx context.Context, request *evaluationv1.ListCourseRequest) (*evaluationv1.ListCourseResponse, error) {
	var curEvaluationId int64
	if request.GetCurEvaluationId() == 0 {
		curEvaluationId = math.MaxInt64
	} else {
		curEvaluationId = request.GetCurEvaluationId()
	}
	list, err := s.svc.ListCourse(ctx, curEvaluationId, request.GetLimit(), request.GetCourseId())
	return &evaluationv1.ListCourseResponse{
		Evaluations: slice.Map(list, func(idx int, src domain.Evaluation) *evaluationv1.Evaluation {
			return convertToV(src)
		}),
	}, err
}

func (s *EvaluationServiceServer) ListMine(ctx context.Context, request *evaluationv1.ListMineRequest) (*evaluationv1.ListMineResponse, error) {
	var curEvaluationId int64
	if request.GetCurEvaluationId() == 0 {
		curEvaluationId = math.MaxInt64
	} else {
		curEvaluationId = request.GetCurEvaluationId()
	}
	list, err := s.svc.ListMine(ctx, curEvaluationId, request.GetLimit(), request.GetUid(), request.GetStatus())
	return &evaluationv1.ListMineResponse{
		Evaluations: slice.Map(list, func(idx int, src domain.Evaluation) *evaluationv1.Evaluation {
			return convertToV(src)
		}),
	}, err
}

func (s *EvaluationServiceServer) ListRecent(ctx context.Context,
	request *evaluationv1.ListRecentRequest) (*evaluationv1.ListRecentResponse, error) {
	var curEvaluationId int64
	if request.GetCurEvaluationId() == 0 {
		curEvaluationId = math.MaxInt64
	} else {
		curEvaluationId = request.GetCurEvaluationId()
	}
	list, err := s.svc.ListRecent(ctx, curEvaluationId, request.GetLimit(), request.GetProperty())
	return &evaluationv1.ListRecentResponse{
		Evaluations: slice.Map(list, func(idx int, src domain.Evaluation) *evaluationv1.Evaluation {
			return convertToV(src)
		}),
	}, err
}

func NewEvaluationServiceServer(svc service.EvaluationService) *EvaluationServiceServer {
	return &EvaluationServiceServer{svc: svc}
}

func (s *EvaluationServiceServer) Register(server grpc.ServiceRegistrar) {
	evaluationv1.RegisterEvaluationServiceServer(server, s)
}

func (s *EvaluationServiceServer) Evaluated(ctx context.Context,
	request *evaluationv1.EvaluatedRequest) (*evaluationv1.EvaluatedResponse, error) {
	evaluated, err := s.svc.Evaluated(ctx, request.GetPublisherId(), request.GetCourseId())
	return &evaluationv1.EvaluatedResponse{
		Evaluated: evaluated,
	}, err
}

func (s *EvaluationServiceServer) Save(ctx context.Context,
	request *evaluationv1.SaveRequest) (*evaluationv1.SaveResponse, error) {
	id, err := s.svc.Save(ctx, convertDomain(request.GetEvaluation()))
	return &evaluationv1.SaveResponse{EvaluationId: id}, err
}

// UpdateStatus TODO 如果是权限不够这个错误要抛上去，需要在这里判断RecordNotFound并在proto里面定义然后往上抛
func (s *EvaluationServiceServer) UpdateStatus(ctx context.Context,
	request *evaluationv1.UpdateStatusRequest) (*evaluationv1.UpdateStatusResponse, error) {
	err := s.svc.UpdateStatus(ctx, request.GetEvaluationId(), request.GetStatus(), request.GetUid())
	return &evaluationv1.UpdateStatusResponse{}, err
}

func convertDomain(e *evaluationv1.Evaluation) domain.Evaluation {
	return domain.Evaluation{
		Id:          e.Id,
		PublisherId: e.PublisherId,
		CourseId:    e.CourseId,
		StarRating:  uint8(e.StarRating),
		Content:     e.Content,
		Status:      e.Status,
	}
}

func convertToV(e domain.Evaluation) *evaluationv1.Evaluation {
	return &evaluationv1.Evaluation{
		Id:          e.Id,
		PublisherId: e.PublisherId,
		CourseId:    e.CourseId,
		StarRating:  uint32(e.StarRating),
		Content:     e.Content,
		Status:      e.Status,
		Utime:       e.Utime.UnixMilli(),
		Ctime:       e.Ctime.UnixMilli(),
	}
}
