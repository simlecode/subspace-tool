package dao

import (
	"context"
	"reflect"
	"strconv"

	"github.com/itering/subscan/util"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/simlecode/subspace-tool/models"
)

const (
	// RedisMetadataKey      = "metadata"
	FillAlreadyBlockNum   = "fill_already_blockNum"
	FillFinalizedBlockNum = "fill_finalized_blockNum"

	MetadataBlockNum          = "blockNum"
	MetadataFinalizedBlockNum = "finalized_blockNum"
)

func (d *Dao) SetMetadata(c context.Context, metadata map[string]interface{}) error {
	// conn, _ := d.redis.GetContext(c)
	// defer conn.Close()
	// args := redis.Args{}.Add(RedisMetadataKey)
	if len(metadata) == 0 {
		return errors.New("ERR: nil metadata")
	}

	for k, v := range metadata {
		if reflect.ValueOf(v).Kind() == reflect.Int {
			if err := d.db.Save(&models.KeyValue{
				Key:   k,
				Value: util.IntToString(v.(int)),
			}).Error; err != nil {
				return err
			}
		} else {
			if err := d.db.Save(&models.KeyValue{
				Key:   k,
				Value: v.(string),
			}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *Dao) IncrMetadata(c context.Context, filed string, incrNum int) error {
	if incrNum == 0 {
		return nil
	}
	// conn, _ := d.redis.GetContext(c)
	// defer conn.Close()
	// _, err = conn.Do("HINCRBY", RedisMetadataKey, filed, incrNum)
	// return

	var kv models.KeyValue
	err := d.db.First(&kv, "`key` = ?", filed).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return d.db.Save(&models.KeyValue{
				Key:   filed,
				Value: util.IntToString(incrNum),
			}).Error
		}
		return err
	}
	kv.Value = util.IntToString(util.StringToInt(kv.Value) + incrNum)
	return d.db.Save(&kv).Error
}

func (d *Dao) GetMetadata(c context.Context) (map[string]string, error) {
	// conn, _ := d.redis.GetContext(c)
	// defer conn.Close()
	// ms, err = redis.StringMap(conn.Do("HGETALL", RedisMetadataKey))
	// return

	var kv []models.KeyValue
	err := d.db.Find(&kv).Error
	if err != nil {
		return nil, err
	}
	ms := make(map[string]string)
	for _, v := range kv {
		ms[v.Key] = v.Value
	}
	return ms, nil
}

func (d *Dao) GetBestBlockNum(c context.Context) (uint64, error) {
	// conn, _ := d.redis.GetContext(c)
	// defer conn.Close()
	// return redis.Uint64(conn.Do("HGET", RedisMetadataKey, "blockNum"))
	var kv models.KeyValue
	err := d.db.First(&kv, "`key` = ?", MetadataBlockNum).Error
	if err != nil {
		return 0, err
	}
	if len(kv.Value) == 0 {
		return 0, nil
	}
	return strconv.ParseUint(kv.Value, 10, 64)
}

func (d *Dao) GetFinalizedBlockNum(c context.Context) (uint64, error) {
	// conn, _ := d.redis.GetContext(c)
	// defer conn.Close()
	// return redis.Uint64(conn.Do("HGET", RedisMetadataKey, "finalized_blockNum"))

	var kv models.KeyValue
	err := d.db.Where("`key` = ?", MetadataFinalizedBlockNum).Take(&kv).Error
	if err != nil {
		return 0, err
	}
	if len(kv.Value) == 0 {
		return 0, nil
	}
	return strconv.ParseUint(kv.Value, 10, 64)
}
