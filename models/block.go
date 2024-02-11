package models

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/simlecode/subspace-tool/types"
	"gorm.io/gorm"
)

type block struct {
	ID             string    `gorm:"column:id;type:varchar(256);primary_key"`
	Author         string    `gorm:"column:author;type:varchar(256);index"`
	Hight          int64     `gorm:"column:height;index"`
	Hash           string    `gorm:"column:hash;type:varchar(256);index"`
	StateRoot      string    `gorm:"column:state_root;type:varchar(256)"`
	Timestamp      time.Time `gorm:"column:timestamp"`
	ExtrinsicsRoot string    `gorm:"column:extrinsics_root;type:varchar(256)"`
	SpecId         string    `gorm:"column:spec_id;type:varchar(256)"`
	ParentHash     string    `gorm:"column:parent_hash;type:varchar(256)"`
	ExtrinsicCount int       `gorm:"column:extrinsic_count;type:int"`
	EventCount     int       `gorm:"column:event_count;type:int"`
}

func fromBlock(src *types.BlockInfo) (*block, error) {
	out := &block{
		ID:             src.ID,
		Author:         src.Author.ID,
		Hash:           src.Hash,
		StateRoot:      src.StateRoot,
		ExtrinsicsRoot: src.ExtrinsicsRoot,
		SpecId:         src.SpecId,
		ParentHash:     src.ParentHash,
		ExtrinsicCount: src.ExtrinsicsCount,
		EventCount:     src.EventsCount,
	}
	height, err := strconv.ParseInt(src.Height, 10, 64)
	if err != nil {
		return nil, err
	}
	out.Hight = height

	out.Timestamp, err = time.Parse(layout, strings.Split(src.Timestamp, ".")[0])
	if err != nil {
		return nil, err
	}

	return out, nil
}

func toBlock(src *block) *types.BlockInfo {
	return &types.BlockInfo{
		ID:              src.ID,
		Author:          types.Author{ID: src.Author},
		Height:          fmt.Sprintf("%v", src.Hight),
		Hash:            src.Hash,
		StateRoot:       src.StateRoot,
		Timestamp:       src.Timestamp.Format(layout),
		ExtrinsicsRoot:  src.ExtrinsicsRoot,
		SpecId:          src.SpecId,
		ParentHash:      src.ParentHash,
		ExtrinsicsCount: src.ExtrinsicCount,
		EventsCount:     src.EventCount,
	}
}

func (b *block) TableName() string {
	return "blocks"
}

var _ BlockRepo = (*blockRepo)(nil)

type blockRepo struct {
	*gorm.DB
}

func newBlockRepo(db *gorm.DB) *blockRepo {
	return &blockRepo{DB: db}
}

func (br *blockRepo) SaveBlock(ctx context.Context, blk *types.BlockInfo) error {
	b, err := fromBlock(blk)
	if err != nil {
		return err
	}

	return br.DB.WithContext(ctx).Save(b).Error
}

func (br *blockRepo) ByBlockHeight(ctx context.Context, blockHeight int) (*types.BlockInfo, error) {
	var blk block
	if err := br.WithContext(ctx).Where("height = ?", blockHeight).Take(&blk).Error; err != nil {
		return nil, err
	}

	return toBlock(&blk), nil
}

func (br *blockRepo) ListBlock(ctx context.Context) ([]*types.BlockInfo, error) {
	var blks []block
	if err := br.WithContext(ctx).Find(&blks).Error; err != nil {
		return nil, err
	}

	var out []*types.BlockInfo
	for _, blk := range blks {
		out = append(out, toBlock(&blk))
	}
	return out, nil
}
