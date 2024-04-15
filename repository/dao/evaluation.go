package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

var ErrorRecordNotFind = gorm.ErrRecordNotFound

type EvaluationDAO interface {
	Insert(ctx context.Context, evaluation Evaluation) (int64, error)
	FindEvaluation(ctx context.Context, publisherId int64, courseId int64) (Evaluation, error)
}

type GORMEvaluationDAO struct {
	db *gorm.DB
}

func NewGORMEvaluationDAO(db *gorm.DB) EvaluationDAO {
	return &GORMEvaluationDAO{db: db}
}

func (dao *GORMEvaluationDAO) FindEvaluation(ctx context.Context, publisherId int64, courseId int64) (Evaluation, error) {
	var e Evaluation
	err := dao.db.WithContext(ctx).
		Where("publisher_id = ? and course_id = ?", publisherId, courseId).
		First(&e).Error
	return e, err
}

func (dao *GORMEvaluationDAO) Insert(ctx context.Context, evaluation Evaluation) (int64, error) {
	now := time.Now().UnixMilli()
	evaluation.Utime = now
	evaluation.Ctime = now
	err := dao.db.WithContext(ctx).Create(&evaluation).Error
	return evaluation.Id, err
}

type Evaluation struct {
	Id          int64 `gorm:"primaryKey,autoIncrement"`
	PublisherId int64
	CourseId    int64
	StarRating  uint8
	Content     string
	Status      uint8
	Utime       int64
	Ctime       int64
}
