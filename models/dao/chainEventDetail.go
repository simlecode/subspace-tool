package dao

import (
	"context"
	"fmt"

	"github.com/itering/subscan/model"
)

type EventDetail struct {
	ID string `gorm:"column:id;type:varchar(256);primary_key"`
	// event id
	Name          string `gorm:"column:name;type:varchar(64);index"`
	BlockHeight   int    `gorm:"column:block_height;index"`
	PublicKey     string `gorm:"column:public_key;type:varchar(128);index"`
	ParentHash    string `gorm:"column:parent_hash;type:varchar(128)"`
	RewardAddress string `gorm:"column:reward_address;type:varchar(128);index"`
}

var SplitTableBlockNum = model.SplitTableBlockNum

func (c EventDetail) TableName() string {
	if c.BlockHeight/SplitTableBlockNum == 0 {
		return "event_details"
	}
	return fmt.Sprintf("event_details_%d", c.BlockHeight/SplitTableBlockNum)
}

// var _ EventDetailRepo = (*eventDetailRepo)(nil)

// type eventDetailRepo struct {
// 	*gorm.DB
// }

// func newEventDetailRepo(db *gorm.DB) *eventDetailRepo {
// 	return &eventDetailRepo{DB: db}
// }

func (d *Dao) CreateEventDetail(txn *GormDB, eventDetail *EventDetail) error {
	if txn != nil {
		query := txn.Save(eventDetail)
		return d.checkDBError(query.Error)
	}
	return d.db.Save(eventDetail).Error
}

func (d *Dao) SaveEventDetail(ctx context.Context, eventDetail *EventDetail) error {
	return d.db.Save(eventDetail).Error
}

func (d *Dao) ByBlockHeight(ctx context.Context, blockHeight int) ([]*EventDetail, error) {
	var eds []*EventDetail
	if err := d.db.Model(&EventDetail{BlockHeight: blockHeight}).Where("block_height = ?", blockHeight).Take(&eds).Error; err != nil {
		return nil, err
	}

	return eds, nil
}

func (d *Dao) ByID(ctx context.Context, eventID string) (*EventDetail, error) {
	var ed EventDetail
	if err := d.db.Where("id = ?", eventID).Take(&ed).Error; err != nil {
		return nil, err
	}

	return &ed, nil
}

func (d *Dao) List(ctx context.Context) ([]*EventDetail, error) {
	var eds []*EventDetail
	if err := d.db.Find(&eds).Error; err != nil {
		return nil, err
	}

	return eds, nil
}
