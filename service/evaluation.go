package service

import (
	"context"
	"github.com/MuxiKeStack/be-evaluation/domain"
	"github.com/MuxiKeStack/be-evaluation/repository"
)

type EvaluationService interface {
	Publish(ctx context.Context, evaluation domain.Evaluation) (int64, error)
	Evaluated(ctx context.Context, publisherId int64, courseId int64) (bool, error)
}

type evaluationService struct {
	repo repository.EvaluationRepository
}

func NewEvaluationService(repo repository.EvaluationRepository) EvaluationService {
	return &evaluationService{repo: repo}
}

func (s *evaluationService) Publish(ctx context.Context, evaluation domain.Evaluation) (int64, error) {
	evaluation.Status = domain.EvaluationStatusPublished
	return s.repo.Create(ctx, evaluation)
}

func (s *evaluationService) Evaluated(ctx context.Context, publisherId int64, courseId int64) (bool, error) {
	return s.repo.Evaluated(ctx, publisherId, courseId)
}
