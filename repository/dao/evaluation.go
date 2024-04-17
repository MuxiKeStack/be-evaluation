package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"time"
)

var ErrorRecordNotFind = gorm.ErrRecordNotFound

type EvaluationDAO interface {
	FindEvaluation(ctx context.Context, publisherId int64, courseId int64) (Evaluation, error)
	UpdateStatus(ctx context.Context, evaluationId int64, status uint32, uid int64) error
	UpdateById(ctx context.Context, evaluation Evaluation) error
	Insert(ctx context.Context, evaluation Evaluation) (int64, error)
	GetListRecent(ctx context.Context, curEvaluationId int64, limit int64, property int32) ([]Evaluation, error)
	GetListCourse(ctx context.Context, curEvaluationId int64, limit int64, courseId int64) ([]Evaluation, error)
	GetListMine(ctx context.Context, curEvaluationId int64, limit int64, uid int64, status int32) ([]Evaluation, error)
	GetCountCourseInvisible(ctx context.Context, courseId int64) (int64, error)
	GetCountMine(ctx context.Context, uid int64, status int32) (int64, error)
	GetDetailById(ctx context.Context, evaluationId int64) (Evaluation, error)
	GetPublishersByCourseIdStatus(ctx context.Context, courseId int64, status int32) ([]int64, error)
}

const (
	EvaluationStatusPublic = 0
	EvaluationStatusFolded = 2
)

type GORMEvaluationDAO struct {
	db *gorm.DB
}

func (dao *GORMEvaluationDAO) GetPublishersByCourseIdStatus(ctx context.Context, courseId int64, status int32) ([]int64, error) {
	var publishers []int64
	err := dao.db.WithContext(ctx).
		Select("publisher_id").
		Model(&Evaluation{}).
		Where("course_id = ? and status = ?", courseId, status).
		Find(&publishers).Error
	return publishers, err
}

func (dao *GORMEvaluationDAO) GetDetailById(ctx context.Context, evaluationId int64) (Evaluation, error) {
	var evaluation Evaluation
	err := dao.db.WithContext(ctx).
		Where("id = ?", evaluationId).
		First(&evaluation).Error
	return evaluation, err
}

func (dao *GORMEvaluationDAO) GetCountMine(ctx context.Context, uid int64, status int32) (int64, error) {
	var count int64
	err := dao.db.WithContext(ctx).Model(&Evaluation{}).
		Where("publisher_id = ? and status = ?", uid, status).
		Count(&count).Error
	return count, err
}

func (dao *GORMEvaluationDAO) GetCountCourseInvisible(ctx context.Context, courseId int64) (int64, error) {
	var count int64
	err := dao.db.WithContext(ctx).Model(&Evaluation{}).
		Where("course_id = ? and status != ?", courseId, EvaluationStatusPublic).
		Count(&count).Error
	return count, err
}

func (dao *GORMEvaluationDAO) GetListCourse(ctx context.Context, curEvaluationId int64, limit int64,
	courseId int64) ([]Evaluation, error) {
	var evaluations []Evaluation
	err := dao.db.WithContext(ctx).
		Where("course_id = ? and status = ? and id < ?", courseId, EvaluationStatusPublic, curEvaluationId).
		Order("utime desc").
		Limit(int(limit)).Find(&evaluations).Error
	return evaluations, err
}

func (dao *GORMEvaluationDAO) GetListMine(ctx context.Context, curEvaluationId int64, limit int64, uid int64,
	status int32) ([]Evaluation, error) {
	var evaluations []Evaluation
	err := dao.db.WithContext(ctx).
		Where("publisher_id = ? and status = ? and id < ?", uid, status, curEvaluationId).
		Order("utime desc").
		Limit(int(limit)).Find(&evaluations).Error
	return evaluations, err
}

func (dao *GORMEvaluationDAO) GetListRecent(ctx context.Context, curEvaluationId int64, limit int64, property int32) ([]Evaluation, error) {
	var evaluations []Evaluation
	query := dao.db.WithContext(ctx)
	const CoursePropertyAny = 0
	if property != 0 {
		query = query.Where("property = ?", property)
	}
	query = query.Where("id < ? and status = ?", curEvaluationId, EvaluationStatusPublic)
	err := query.Limit(int(limit)).Order("utime desc").Find(&evaluations).Error
	return evaluations, err
}

func (dao *GORMEvaluationDAO) UpdateById(ctx context.Context, evaluation Evaluation) error {
	res := dao.db.WithContext(ctx).Model(&Evaluation{}).
		Where("id = ? and publisher_id = ?", evaluation.Id, evaluation.PublisherId).
		Updates(map[string]any{
			"star_rating": evaluation.StarRating,
			"content":     evaluation.Content,
			"status":      evaluation.Status,
			"utime":       time.Now().UnixMilli(),
		})
	err := res.Error
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return errors.New("更新数据失败")
	}
	return nil
}

func (dao *GORMEvaluationDAO) UpdateStatus(ctx context.Context, evaluationId int64, status uint32, uid int64) error {
	res := dao.db.WithContext(ctx).Model(&Evaluation{}).
		// 不让用户直接更改被折叠的课评的状态
		Where("id = ? and publisher_id = ? and status != ?", evaluationId, uid, EvaluationStatusFolded).
		Updates(map[string]any{
			"utime":  time.Now().UnixMilli(),
			"status": status,
		})
	err := res.Error
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return errors.New("更新数据失败")
	}
	return nil
}

func (dao *GORMEvaluationDAO) Insert(ctx context.Context, evaluation Evaluation) (int64, error) {
	now := time.Now().UnixMilli()
	evaluation.Ctime = now
	evaluation.Utime = now
	err := dao.db.WithContext(ctx).Create(&evaluation).Error
	return evaluation.Id, err
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

type Evaluation struct {
	Id             int64 `gorm:"primaryKey,autoIncrement"`
	PublisherId    int64 `gorm:"uniqueIndex:publisherId_courseId"`
	CourseId       int64 `gorm:"uniqueIndex:publisherId_courseId"`
	CourseProperty int32 // 冗余一个课程性质，用于查询
	StarRating     uint8
	Content        string
	Status         int32
	Utime          int64
	Ctime          int64
}
