package models

import (
	"context"
	"strconv"

	"github.com/simlecode/subspace-tool/types"
	"gorm.io/gorm"
)

type event struct {
	ID           string    `gorm:"column:id;type:varchar(256);primary_key"`
	Name         string    `gorm:"column:name;type:varchar(256)"`
	Phase        string    `gorm:"column:phase;type:varchar(256)"`
	IndexInBlock int       `gorm:"column:index_in_block;type:int"`
	BlockHight   int       `gorm:"column:block_height;type:int;index"`
	BlockID      string    `gorm:"column:block_id;type:varchar(256)"`
	Extrinsic    Extrinsic `gorm:"embedded;embeddedPrefix:extrinsic_"`
}

type Extrinsic struct {
	IndexInBlock int `gorm:"column:index_in_block;type:int"`
}

func fromEvent(e *types.Event) (*event, error) {
	out := &event{
		ID:           e.Node.ID,
		Name:         e.Node.Name,
		Phase:        e.Node.Phase,
		IndexInBlock: e.Node.IndexInBlock,
		BlockID:      e.Node.Block.ID,
		Extrinsic: Extrinsic{
			IndexInBlock: e.Node.Extrinsic.IndexInBlock,
		},
	}
	height, err := strconv.ParseInt(e.Node.Block.Height, 10, 64)
	if err != nil {
		return nil, err
	}
	out.BlockHight = int(height)

	return out, nil
}

func toEvent(e *event) *types.Event {
	return &types.Event{
		Node: types.Node{
			ID:           e.ID,
			Name:         e.Name,
			Phase:        e.Phase,
			IndexInBlock: e.IndexInBlock,
			Block: types.Block{
				ID:     e.BlockID,
				Height: strconv.Itoa(e.BlockHight),
			},
			Extrinsic: types.Extrinsic{
				IndexInBlock: e.Extrinsic.IndexInBlock,
				Block: types.Block{
					ID:     e.BlockID,
					Height: strconv.Itoa(e.BlockHight),
				},
			},
		},
	}
}

func (e *event) TableName() string {
	return "events"
}

var _ EventRepo = (*eventRepo)(nil)

type eventRepo struct {
	*gorm.DB
}

func newEventRepo(db *gorm.DB) *eventRepo {
	return &eventRepo{DB: db}
}

func (er *eventRepo) SaveEvent(ctx context.Context, event *types.Event) error {
	e, err := fromEvent(event)
	if err != nil {
		return err
	}

	return er.DB.WithContext(ctx).Save(e).Error
}

func (er *eventRepo) ByBlockHeight(ctx context.Context, blockHeight int) ([]*types.Event, error) {
	var events []*event
	if err := er.WithContext(ctx).Where("block_height = ?", blockHeight).Find(&events).Error; err != nil {
		return nil, err
	}
	out := make([]*types.Event, 0, len(events))
	for _, e := range events {
		out = append(out, toEvent(e))
	}

	return out, nil
}
