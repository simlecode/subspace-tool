package models

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/simlecode/subspace-tool/types"
	"gorm.io/gorm"
)

const layout = "2006-01-02T15:04:05"

type extrinsic struct {
	ID           string    `gorm:"column:id;type:varchar(256);primary_key"`
	Name         string    `gorm:"column:name;type:varchar(256)"`
	IndexInBlock int       `gorm:"column:index_in_block;type:int"`
	BlockHight   int       `gorm:"column:block_height;type:int;index"`
	Timestamp    time.Time `gorm:"column:timestamp"`
	Hash         string    `gorm:"column:hash;type:varchar(256);index"`
	Success      bool      `gorm:"column:success;type:bool"`
	Cursor       string    `gorm:"column:cursor;type:varchar(128)"`
}

func fromExtrinsic(src *types.Event) (*extrinsic, error) {
	e := &extrinsic{
		ID:           src.Node.ID,
		Name:         src.Node.Name,
		IndexInBlock: src.Node.IndexInBlock,
		Hash:         src.Node.Hash,
		Success:      src.Node.Success,
		Cursor:       src.Cursor,
	}
	height, err := strconv.ParseInt(src.Node.Block.Height, 10, 64)
	if err != nil {
		return nil, err
	}
	e.BlockHight = int(height)

	e.Timestamp, err = time.Parse(layout, strings.Split(src.Node.Block.Timestamp, ".")[0])
	if err != nil {
		return nil, err
	}

	return e, nil
}

func toExtrinsic(e *extrinsic) *types.Event {
	return &types.Event{
		Node: types.Node{
			ID:           e.ID,
			Name:         e.Name,
			IndexInBlock: e.IndexInBlock,
			Block: types.Block{
				Timestamp: e.Timestamp.Format(layout),
				Height:    strconv.Itoa(e.BlockHight),
			},
			Hash:    e.Hash,
			Success: e.Success,
		},
		Cursor: e.Cursor,
	}
}

func (e *extrinsic) TableName() string {
	return "extrinsics"
}

var _ ExtrinsicRepo = (*extrinsicRepo)(nil)

type extrinsicRepo struct {
	*gorm.DB
}

func newExtrinsicRepo(db *gorm.DB) *extrinsicRepo {
	return &extrinsicRepo{DB: db}
}

func (er *extrinsicRepo) SaveExtrinsic(ctx context.Context, event *types.Event) error {
	e, err := fromExtrinsic(event)
	if err != nil {
		return err
	}

	return er.DB.WithContext(ctx).Save(e).Error
}

func (er *extrinsicRepo) ByBlockHeight(ctx context.Context, blockHeight int) ([]*types.Event, error) {
	var extrinsics []*extrinsic
	if err := er.WithContext(ctx).Where("block_height = ?", blockHeight).Find(&extrinsics).Error; err != nil {
		return nil, err
	}
	out := make([]*types.Event, 0, len(extrinsics))
	for _, e := range extrinsics {
		out = append(out, toExtrinsic(e))
	}

	return out, nil
}

func (er *extrinsicRepo) List(ctx context.Context, limit int) ([]*types.Event, error) {
	var extrinsics []*extrinsic
	if err := er.WithContext(ctx).Limit(limit).Order("block_height desc").Find(&extrinsics).Error; err != nil {
		return nil, err
	}
	out := make([]*types.Event, 0, len(extrinsics))
	for _, e := range extrinsics {
		out = append(out, toExtrinsic(e))
	}

	return out, nil
}
