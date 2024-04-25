package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var ErrorRecordNotFind = gorm.ErrRecordNotFound

type EvaluationDAO interface {
	FindEvaluation(ctx context.Context, publisherId int64, courseId int64) (Evaluation, error)
	UpdateStatus(ctx context.Context, evaluationId int64, status uint32, uid int64) (OldEvaluation, error)
	// 更新课评并返回旧的星级
	UpdateById(ctx context.Context, evaluation Evaluation) (OldEvaluation, error)
	Insert(ctx context.Context, evaluation Evaluation) (int64, error)
	GetListRecent(ctx context.Context, curEvaluationId int64, limit int64, property int32) ([]Evaluation, error)
	GetListCourse(ctx context.Context, curEvaluationId int64, limit int64, courseId int64) ([]Evaluation, error)
	GetListMine(ctx context.Context, curEvaluationId int64, limit int64, uid int64, status int32) ([]Evaluation, error)
	GetCountCourseInvisible(ctx context.Context, courseId int64) (int64, error)
	GetCountMine(ctx context.Context, uid int64, status int32) (int64, error)
	GetDetailById(ctx context.Context, evaluationId int64) (Evaluation, error)
	GetPublishersByCourseIdStatus(ctx context.Context, courseId int64, status int32) ([]int64, error)
	GetCompositeScoreByCourseId(ctx context.Context, courseId int64) (CompositeScore, error)
}

const (
	EvaluationStatusPublic  = 0
	EvaluationStatusPrivate = 1
	EvaluationStatusFolded  = 2
)

type GORMEvaluationDAO struct {
	db *gorm.DB
}

func (dao *GORMEvaluationDAO) GetCompositeScoreByCourseId(ctx context.Context, courseId int64) (CompositeScore, error) {
	// 这是聚合的写法
	//var averageRating float64
	//err := dao.db.WithContext(ctx).
	//	Model(&Evaluation{}).
	//	Select("COALESCE(AVG(CAST(star_rating as double)), 0) as average_rating").
	//	Where("course_id = ? and status = ?", courseId, EvaluationStatusPublic).
	//	First(&averageRating).Error
	var cs CompositeScore
	err := dao.db.WithContext(ctx).
		Where("course_id = ?", courseId).
		First(&cs).Error
	return cs, err
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
	err := dao.db.WithContext(ctx).
		Model(&Evaluation{}).
		Where("publisher_id = ? and status = ?", uid, status).
		Count(&count).Error
	return count, err
}

func (dao *GORMEvaluationDAO) GetCountCourseInvisible(ctx context.Context, courseId int64) (int64, error) {
	var count int64
	err := dao.db.WithContext(ctx).
		Model(&Evaluation{}).
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
	if property != CoursePropertyAny {
		query = query.Where("property = ?", property)
	}
	query = query.Where("status = ? and id < ?", EvaluationStatusPublic, curEvaluationId)
	err := query.Limit(int(limit)).Order("utime desc").Find(&evaluations).Error
	return evaluations, err
}

type OldEvaluation struct {
	CourseId   int64
	StarRating uint8
	Status     int32
}

func (dao *GORMEvaluationDAO) UpdateById(ctx context.Context, evaluation Evaluation) (OldEvaluation, error) {
	now := time.Now().UnixMilli()
	var oe OldEvaluation
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先获取原有评价的星级，并锁定该行直到事务结束，这里是一个检查，然后做某事的场景
		err := tx.Model(&Evaluation{}).
			Clauses(clause.Locking{Strength: "UPDATE"}). // 添加行级锁
			Select("course_id, star_rating, status").
			Where("id = ? AND publisher_id = ?", evaluation.Id, evaluation.PublisherId).
			First(&oe).Error
		if err != nil {
			return err
		}

		res := tx.Model(&Evaluation{}).
			Where("id = ? AND publisher_id = ?", evaluation.Id, evaluation.PublisherId).
			Updates(map[string]any{
				"star_rating": evaluation.StarRating,
				"content":     evaluation.Content,
				"status":      evaluation.Status,
				"utime":       now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("更新数据失败")
		}
		switch {
		case oe.Status == EvaluationStatusPrivate && evaluation.Status == EvaluationStatusPrivate:
			return nil
		case oe.Status == EvaluationStatusPublic && evaluation.Status == EvaluationStatusPublic:
			if oe.StarRating != evaluation.StarRating { // 只在评分改变时更新得分
				sql := `
				UPDATE composite_scores
				SET 
    				score = ((score * rater_cnt - ? + ?) / rater_cnt)
				WHERE course_id = ?
				`
				return tx.Exec(sql, float64(oe.StarRating), float64(evaluation.StarRating), oe.CourseId).Error
			}
			return nil
		case oe.Status == EvaluationStatusPrivate && evaluation.Status == EvaluationStatusPublic:
			sql := `
        	UPDATE composite_scores
        	SET score = ((score * rater_cnt + ?) / (rater_cnt + 1)),
        	    rater_cnt = rater_cnt + 1
        	WHERE course_id = ?;
    		`
			return tx.Exec(sql, float64(evaluation.StarRating), oe.CourseId).Error
		case oe.Status == EvaluationStatusPublic && evaluation.Status == EvaluationStatusPrivate:
			sql := `
			UPDATE composite_scores
			SET 
    			score = CASE 
                			WHEN rater_cnt = 1 THEN 0
                			ELSE ((score * rater_cnt - ?) / (rater_cnt - 1))
            			END,
    			rater_cnt = CASE 
                    			WHEN rater_cnt = 1 THEN 0
                    			ELSE rater_cnt - 1
                			END
			WHERE course_id = ?;
    		`
			return tx.Exec(sql, float64(oe.StarRating), oe.CourseId).Error
		default:
			return errors.New("非法的课评新旧状态变迁")
		}
	})
	if err != nil {
		return OldEvaluation{}, err
	}
	return oe, nil
}

func (dao *GORMEvaluationDAO) UpdateStatus(ctx context.Context, evaluationId int64, status uint32, uid int64) (OldEvaluation, error) {
	var oe OldEvaluation
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先获取原有评价的状态，并锁定该行直到事务结束
		err := tx.Model(&Evaluation{}).
			Clauses(clause.Locking{Strength: "UPDATE"}). // 添加行级锁
			Select("course_id, star_rating, status").
			Where("id = ? AND publisher_id = ?", evaluationId, uid).
			First(&oe).Error
		if err != nil {
			return err
		}

		// 更新状态
		res := tx.Model(&Evaluation{}).
			Where("id = ? AND publisher_id = ? AND status != ?", evaluationId, uid, EvaluationStatusFolded).
			Updates(map[string]any{
				"utime":  time.Now().UnixMilli(),
				"status": status,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("更新数据失败")
		}

		// 更新综分，根据评价状态的变化
		switch {
		case oe.Status == EvaluationStatusPrivate && status == EvaluationStatusPublic:
			// 新增语义
			sql := `
			UPDATE composite_scores
			SET score = ((score * rater_cnt + ?) / (rater_cnt + 1)),
			    rater_cnt = rater_cnt + 1
			WHERE course_id = ?;
			`
			// 执行 SQL 更新操作
			return tx.Exec(sql, float64(oe.StarRating), oe.CourseId).Error
		case oe.Status == EvaluationStatusPublic && status == EvaluationStatusPrivate:
			// 删除语义
			sql := `
			UPDATE composite_scores
			SET 
    			score = CASE 
                			WHEN rater_cnt = 1 THEN 0
                			ELSE ((score * rater_cnt - ?) / (rater_cnt - 1))
            			END,
    			rater_cnt = CASE 
                    			WHEN rater_cnt = 1 THEN 0
                    			ELSE rater_cnt - 1
                			END
			WHERE course_id = ?;
    		`
			// 执行 SQL 更新操作
			return tx.Exec(sql, float64(oe.StarRating), oe.CourseId).Error
		default:
			return nil
		}
	})
	if err != nil {
		return OldEvaluation{}, err
	}
	return oe, nil
}

func (dao *GORMEvaluationDAO) Insert(ctx context.Context, evaluation Evaluation) (int64, error) {
	now := time.Now().UnixMilli()
	evaluation.Ctime = now
	evaluation.Utime = now

	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建评价记录
		err := tx.Create(&evaluation).Error
		if err != nil {
			return err
		}
		if evaluation.Status == EvaluationStatusPrivate {
			// 非公开的课评，不计入评分
			return nil
		}
		// 使用 upsert 来更新或插入分数
		sql := `
		INSERT INTO composite_scores (course_id, score, rater_cnt)
		VALUES (?, ?, 1)
		ON DUPLICATE KEY UPDATE 
		    score = ((score * rater_cnt + VALUES(score)) / (rater_cnt + 1)),
		    rater_cnt = rater_cnt + 1
		`
		// 执行 SQL 更新操作
		return tx.Exec(sql, evaluation.CourseId, float64(evaluation.StarRating)).Error
	})

	if err != nil {
		return 0, err
	}
	return evaluation.Id, nil
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

// TODO 设计索引，优化查询
type Evaluation struct {
	Id             int64 `gorm:"primaryKey,autoIncrement"`
	PublisherId    int64 `gorm:"uniqueIndex:publisherId_courseId"`
	CourseId       int64 `gorm:"uniqueIndex:publisherId_courseId;index:courseId_status"`
	CourseProperty int32 // 冗余一个课程性质，用于查询
	StarRating     uint8
	Content        string
	Status         int32 `gorm:"index:courseId_status"`
	Utime          int64
	Ctime          int64
}

// 维护一个综合得分
type CompositeScore struct {
	Id       int64 `gorm:"primaryKey,autoIncrement"`
	CourseId int64 `gorm:"uniqueIndex"`
	Score    float64
	RaterCnt int64
}
