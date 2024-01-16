package models

import (
	"context"

	"github.com/simlecode/subspace-tool/types"
	"gorm.io/gorm"
)

type eventDetail struct {
	ID            string `gorm:"column:id;type:varchar(256);primary_key"`
	Name          string `gorm:"column:name;type:varchar(64)"`
	BlockHight    int64  `gorm:"column:block_height;index"`
	PublicKey     string `gorm:"column:public_key;type:varchar(128);index"`
	ParentHash    string `gorm:"column:parent_hash;type:varchar(64)"`
	RewardAddress string `gorm:"column:reward_address;type:varchar(128);index"`
}

func fromEventDetail(src *types.EventDetail) (*eventDetail, error) {
	out := &eventDetail{
		ID:            src.ID,
		Name:          src.Name,
		BlockHight:    src.EventArgs.Height,
		PublicKey:     src.EventArgs.PublicKey,
		ParentHash:    src.EventArgs.ParentHash,
		RewardAddress: src.EventArgs.RewardAddress,
	}

	return out, nil
}

func toEventDetail(src *eventDetail) *types.EventDetail {
	return &types.EventDetail{
		ID:        src.ID,
		Name:      src.Name,
		EventArgs: types.EventArgs{Height: src.BlockHight, PublicKey: src.PublicKey, RewardAddress: src.RewardAddress, ParentHash: src.ParentHash},
	}
}

func (ed *eventDetail) TableName() string {
	return "event_details"
}

var _ EventDetailRepo = (*eventDetailRepo)(nil)

type eventDetailRepo struct {
	*gorm.DB
}

func newEventDetailRepo(db *gorm.DB) *eventDetailRepo {
	return &eventDetailRepo{DB: db}
}

func (er *eventDetailRepo) SaveEventDetail(ctx context.Context, eventDetail *types.EventDetail) error {
	detail, err := fromEventDetail(eventDetail)
	if err != nil {
		return err
	}

	return er.DB.WithContext(ctx).Save(detail).Error
}

func (er *eventDetailRepo) ByBlockHeight(ctx context.Context, blockHeight int) (*types.EventDetail, error) {
	var d eventDetail
	if err := er.WithContext(ctx).Where("block_height = ?", blockHeight).Take(&d).Error; err != nil {
		return nil, err
	}

	return toEventDetail(&d), nil
}

func (er *eventDetailRepo) ByID(ctx context.Context, eventID string) (*types.EventDetail, error) {
	var d eventDetail
	if err := er.WithContext(ctx).Where("id = ?", eventID).Take(&d).Error; err != nil {
		return nil, err
	}

	return toEventDetail(&d), nil
}

func (er *eventDetailRepo) List(ctx context.Context) ([]*types.EventDetail, error) {
	var eds []eventDetail
	if err := er.WithContext(ctx).Find(&eds).Error; err != nil {
		return nil, err
	}
	out := make([]*types.EventDetail, 0, len(eds))
	for _, e := range eds {
		out = append(out, toEventDetail(&e))
	}

	return out, nil
}
