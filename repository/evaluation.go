package repository

import (
	"context"
	coursev1 "github.com/MuxiKeStack/be-api/gen/proto/course/v1"
	evaluationv1 "github.com/MuxiKeStack/be-api/gen/proto/evaluation/v1"
	"github.com/MuxiKeStack/be-evaluation/domain"
	"github.com/MuxiKeStack/be-evaluation/repository/dao"
	"github.com/ecodeclub/ekit/slice"
	"time"
)

type EvaluationRepository interface {
	Evaluated(ctx context.Context, publisherId int64, courseId int64) (bool, error)
	UpdateStatus(ctx context.Context, evaluationId int64, status evaluationv1.EvaluationStatus, uid int64) error
	Update(ctx context.Context, evaluation domain.Evaluation) error
	Create(ctx context.Context, evaluation domain.Evaluation) (int64, error)
	GetListRecent(ctx context.Context, curEvaluationId int64, limit int64, property coursev1.CourseProperty) ([]domain.Evaluation, error)
	GetListCourse(ctx context.Context, curEvaluationId int64, limit int64, courseId int64) ([]domain.Evaluation, error)
	GetListMine(ctx context.Context, curEvaluationId int64, limit int64, uid int64, status evaluationv1.EvaluationStatus) ([]domain.Evaluation, error)
	GetCountCourseInvisible(ctx context.Context, courseId int64) (int64, error)
	GetCountMine(ctx context.Context, uid int64, status evaluationv1.EvaluationStatus) (int64, error)
	GetDetailById(ctx context.Context, evaluationId int64) (domain.Evaluation, error)
	GetPublishersByCourseIdStatus(ctx context.Context, courseId int64, status evaluationv1.EvaluationStatus) ([]int64, error)
}

type evaluationRepository struct {
	dao dao.EvaluationDAO
}

func (repo *evaluationRepository) GetPublishersByCourseIdStatus(ctx context.Context, courseId int64, status evaluationv1.EvaluationStatus) ([]int64, error) {
	return repo.dao.GetPublishersByCourseIdStatus(ctx, courseId, int32(status))
}

func (repo *evaluationRepository) GetDetailById(ctx context.Context, evaluationId int64) (domain.Evaluation, error) {
	evaluation, err := repo.dao.GetDetailById(ctx, evaluationId)
	return repo.toDomain(evaluation), err
}

func (repo *evaluationRepository) GetCountMine(ctx context.Context, uid int64, status evaluationv1.EvaluationStatus) (int64, error) {
	return repo.dao.GetCountMine(ctx, uid, int32(status))
}

func (repo *evaluationRepository) GetCountCourseInvisible(ctx context.Context, courseId int64) (int64, error) {
	return repo.dao.GetCountCourseInvisible(ctx, courseId)
}

func (repo *evaluationRepository) GetListCourse(ctx context.Context, curEvaluationId int64, limit int64,
	courseId int64) ([]domain.Evaluation, error) {
	evaluations, err := repo.dao.GetListCourse(ctx, curEvaluationId, limit, courseId)
	return slice.Map(evaluations, func(idx int, src dao.Evaluation) domain.Evaluation {
		return repo.toDomain(src)
	}), err
}

func (repo *evaluationRepository) GetListMine(ctx context.Context, curEvaluationId int64, limit int64, uid int64,
	status evaluationv1.EvaluationStatus) ([]domain.Evaluation, error) {
	evaluations, err := repo.dao.GetListMine(ctx, curEvaluationId, limit, uid, int32(status))
	return slice.Map(evaluations, func(idx int, src dao.Evaluation) domain.Evaluation {
		return repo.toDomain(src)
	}), err
}

func (repo *evaluationRepository) GetListRecent(ctx context.Context, curEvaluationId int64, limit int64,
	property coursev1.CourseProperty) ([]domain.Evaluation, error) {
	evaluations, err := repo.dao.GetListRecent(ctx, curEvaluationId, limit, int32(property))
	return slice.Map(evaluations, func(idx int, src dao.Evaluation) domain.Evaluation {
		return repo.toDomain(src)
	}), err
}

func (repo *evaluationRepository) Update(ctx context.Context, evaluation domain.Evaluation) error {
	return repo.dao.UpdateById(ctx, repo.toEntity(evaluation))
}

func (repo *evaluationRepository) Create(ctx context.Context, evaluation domain.Evaluation) (int64, error) {
	return repo.dao.Insert(ctx, repo.toEntity(evaluation))
}

func (repo *evaluationRepository) UpdateStatus(ctx context.Context, evaluationId int64, status evaluationv1.EvaluationStatus, uid int64) error {
	return repo.dao.UpdateStatus(ctx, evaluationId, uint32(status), uid)
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

func (repo *evaluationRepository) toEntity(e domain.Evaluation) dao.Evaluation {
	return dao.Evaluation{
		Id:             e.Id,
		PublisherId:    e.PublisherId,
		CourseId:       e.CourseId,
		CourseProperty: int32(e.CourseProperty),
		StarRating:     e.StarRating,
		Content:        e.Content,
		Status:         int32(e.Status),
	}
}

func (repo *evaluationRepository) toDomain(e dao.Evaluation) domain.Evaluation {
	return domain.Evaluation{
		Id:             e.Id,
		PublisherId:    e.PublisherId,
		CourseId:       e.CourseId,
		CourseProperty: coursev1.CourseProperty(e.CourseProperty),
		StarRating:     e.StarRating,
		Content:        e.Content,
		Status:         evaluationv1.EvaluationStatus(e.Status),
		Utime:          time.UnixMilli(e.Utime),
		Ctime:          time.UnixMilli(e.Ctime),
	}
}
