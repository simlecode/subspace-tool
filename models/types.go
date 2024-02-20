package models

type Space struct {
	ID        int   `gorm:"column:id;primary_key"`
	Timestamp int64 `gorm:"column:timestamp;index"`
	Pledged   int64 `gorm:"column:pledged;index"`
}

func (s *Space) TableName() string {
	return "spaces"
}
