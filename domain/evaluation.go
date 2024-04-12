package domain

import "time"

type Evaluation struct {
	Id          int64
	PublisherId int64
	CourseId    int64
	StarRating  uint8
	Content     string
	Status      EvaluationStatus
	Utime       time.Time
	Ctime       time.Time
}

type EvaluationStatus uint8

const (
	EvaluationStatusUnknown = iota
	EvaluationStatusPublished
	EvaluationStatusPrivate
	EvaluationStatusFolded
)

func (s EvaluationStatus) Uint8() uint8 {
	return uint8(s)
}
