package models

import (
	"context"
	"fmt"
	"time"

	"github.com/simlecode/subspace-tool/types"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type EventRepo interface {
	SaveEvent(ctx context.Context, event *types.Event) error
	ByBlockHeight(ctx context.Context, blockHeight int) ([]*types.Event, error)
}

type ExtrinsicRepo interface {
	SaveExtrinsic(ctx context.Context, event *types.Event) error
	ByBlockHeight(ctx context.Context, blockHeight int) ([]*types.Event, error)
	List(ctx context.Context, limit int) ([]*types.Event, error)
}

type BlockRepo interface {
	SaveBlock(ctx context.Context, block *types.BlockInfo) error
	ByBlockHeight(ctx context.Context, blockHeight int) (*types.BlockInfo, error)
}

type Repo interface {
	EventRepo() EventRepo
	ExtrinsicRepo() ExtrinsicRepo
	BlockRepo() BlockRepo
}

type mysqlRepo struct {
	*gorm.DB
}

func (r *mysqlRepo) EventRepo() EventRepo {
	return newEventRepo(r.DB)
}

func (r *mysqlRepo) ExtrinsicRepo() ExtrinsicRepo {
	return newExtrinsicRepo(r.DB)
}

func (r *mysqlRepo) BlockRepo() BlockRepo {
	return newBlockRepo(r.DB)
}

func (r *mysqlRepo) AutoMigrate() error {
	return r.DB.AutoMigrate(&event{}, &extrinsic{}, &block{})
}

func OpenMysql(connectionString string, debug bool) (Repo, error) {
	db, err := gorm.Open(mysql.Open(connectionString), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info), // 日志配置
	})
	if err != nil {
		return nil, fmt.Errorf("[db connection failed] Database name: %s %w", connectionString, err)
	}

	db.Set("gorm:table_options", "CHARSET=utf8mb4")
	if debug {
		db = db.Debug()
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(100)
	sqlDB.SetConnMaxLifetime(time.Minute * 10)

	// 使用插件
	// db.Use(&TracePlugin{})
	r := &mysqlRepo{
		db,
	}

	return r, r.AutoMigrate()
}
