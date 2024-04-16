package domain

import (
	coursev1 "github.com/MuxiKeStack/be-api/gen/proto/course/v1"
	evaluationv1 "github.com/MuxiKeStack/be-api/gen/proto/evaluation/v1"
	"time"
)

type Evaluation struct {
	Id             int64
	PublisherId    int64
	CourseId       int64
	CourseProperty coursev1.CourseProperty
	StarRating     uint8
	Content        string
	Status         evaluationv1.EvaluationStatus
	Utime          time.Time
	Ctime          time.Time
}
