package repository

import (
	"context"
	"github.com/MuxiKeStack/be-evaluation/domain"
	"github.com/MuxiKeStack/be-evaluation/repository/dao"
)

type EvaluationRepository interface {
	Create(ctx context.Context, evaluation domain.Evaluation) (int64, error)
	Evaluated(ctx context.Context, publisherId int64, courseId int64) (bool, error)
}

type evaluationRepository struct {
	dao dao.EvaluationDAO
}

func NewEvaluationRepository(dao dao.EvaluationDAO) EvaluationRepository {
	return &evaluationRepository{dao: dao}
}

func (repo *evaluationRepository) Evaluated(ctx context.Context, publisherId int64, courseId int64) (bool, error) {
	_, err := repo.dao.FindEvaluation(ctx, publisherId, courseId)
	switch {
	case err == nil:
		return true, nil
	case err == dao.ErrorRecordNotFind:
		return false, nil
	default:
		return false, err
	}
}

func (repo *evaluationRepository) Create(ctx context.Context, evaluation domain.Evaluation) (int64, error) {
	return repo.dao.Insert(ctx, repo.toEntity(evaluation))
}

func (repo *evaluationRepository) toEntity(e domain.Evaluation) dao.Evaluation {
	return dao.Evaluation{
		Id:          e.Id,
		PublisherId: e.PublisherId,
		CourseId:    e.CourseId,
		StarRating:  e.StarRating,
		Content:     e.Content,
		Status:      e.Status.Uint8(),
	}
}
