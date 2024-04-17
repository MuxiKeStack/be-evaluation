package service

import (
	"context"
	"errors"
	coursev1 "github.com/MuxiKeStack/be-api/gen/proto/course/v1"
	evaluationv1 "github.com/MuxiKeStack/be-api/gen/proto/evaluation/v1"
	"github.com/MuxiKeStack/be-evaluation/domain"
	"github.com/MuxiKeStack/be-evaluation/repository"
)

var ErrPermissionDenied = errors.New("没有权限")

type EvaluationService interface {
	Evaluated(ctx context.Context, publisherId int64, courseId int64) (bool, error)
	Save(ctx context.Context, evaluation domain.Evaluation) (int64, error)
	UpdateStatus(ctx context.Context, evaluationId int64, status evaluationv1.EvaluationStatus, uid int64) error
	ListRecent(ctx context.Context, curEvaluationId int64, limit int64, property coursev1.CourseProperty) ([]domain.Evaluation, error)
	ListCourse(ctx context.Context, curEvaluationId int64, limit int64, courseId int64) ([]domain.Evaluation, error)
	ListMine(ctx context.Context, curEvaluationId int64, limit int64, uid int64, status evaluationv1.EvaluationStatus) ([]domain.Evaluation, error)
	CountCourseInvisible(ctx context.Context, courseId int64) (int64, error)
	CountMine(ctx context.Context, uid int64, status evaluationv1.EvaluationStatus) (int64, error)
	Detail(ctx context.Context, evaluationId int64) (domain.Evaluation, error)
	VisiblePublishersCourse(ctx context.Context, courseId int64) ([]int64, error)
}

type evaluationService struct {
	repo         repository.EvaluationRepository
	courseClient coursev1.CourseServiceClient
}

func NewEvaluationService(repo repository.EvaluationRepository, courseClient coursev1.CourseServiceClient) EvaluationService {
	return &evaluationService{repo: repo, courseClient: courseClient}
}

func (s *evaluationService) VisiblePublishersCourse(ctx context.Context, courseId int64) ([]int64, error) {
	return s.repo.GetPublishersByCourseIdStatus(ctx, courseId, evaluationv1.EvaluationStatus_Public)
}

func (s *evaluationService) Detail(ctx context.Context, evaluationId int64) (domain.Evaluation, error) {
	return s.repo.GetDetailById(ctx, evaluationId)
}

func (s *evaluationService) CountMine(ctx context.Context, uid int64, status evaluationv1.EvaluationStatus) (int64, error) {
	return s.repo.GetCountMine(ctx, uid, status)
}

func (s *evaluationService) CountCourseInvisible(ctx context.Context, courseId int64) (int64, error) {
	return s.repo.GetCountCourseInvisible(ctx, courseId)
}

func (s *evaluationService) ListCourse(ctx context.Context, curEvaluationId int64, limit int64, courseId int64) ([]domain.Evaluation, error) {
	return s.repo.GetListCourse(ctx, curEvaluationId, limit, courseId)
}

func (s *evaluationService) ListMine(ctx context.Context, curEvaluationId int64, limit int64, uid int64,
	status evaluationv1.EvaluationStatus) ([]domain.Evaluation, error) {
	return s.repo.GetListMine(ctx, curEvaluationId, limit, uid, status)
}

func (s *evaluationService) ListRecent(ctx context.Context, curEvaluationId int64, limit int64,
	property coursev1.CourseProperty) ([]domain.Evaluation, error) {
	return s.repo.GetListRecent(ctx, curEvaluationId, limit, property)
}

func (s *evaluationService) UpdateStatus(ctx context.Context, evaluationId int64, status evaluationv1.EvaluationStatus, uid int64) error {
	return s.repo.UpdateStatus(ctx, evaluationId, status, uid)
}

func (s *evaluationService) Save(ctx context.Context, evaluation domain.Evaluation) (int64, error) {
	// 聚合property
	res, err := s.courseClient.GetDetailById(ctx, &coursev1.GetDetailByIdRequest{
		CourseId: evaluation.CourseId,
	})
	if err != nil {
		return 0, err
	}
	evaluation.CourseProperty = res.GetCourse().GetProperty()
	// 下面是一个upsert语义
	if evaluation.Id > 0 {
		err = s.repo.Update(ctx, evaluation)
		return evaluation.Id, err
	}
	return s.repo.Create(ctx, evaluation)
}

func (s *evaluationService) Evaluated(ctx context.Context, publisherId int64, courseId int64) (bool, error) {
	return s.repo.Evaluated(ctx, publisherId, courseId)
}
