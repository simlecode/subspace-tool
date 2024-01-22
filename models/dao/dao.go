package dao

import (
	"context"

	"github.com/jinzhu/gorm"
)

var (
	DaemonAction = []string{"substrate"}
)

// dao
type Dao struct {
	db *gorm.DB
}

var _ IDao = (*Dao)(nil)

// New new a dao and return.
func New(ctx context.Context, mysqlDsn string) (*Dao, *DbStorage, error) {
	db, err := newDb(mysqlDsn)
	if err != nil {
		return nil, nil, err
	}

	dao := &Dao{
		db: db,
	}
	dao.Migration(ctx)
	storage := &DbStorage{db: db}

	return dao, storage, nil
}

// Close close the resource.
func (d *Dao) Close() {
	_ = d.db.Close()
}
