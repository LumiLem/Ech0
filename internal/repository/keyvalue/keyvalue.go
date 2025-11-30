package keyvalue

import (
	"context"

	"gorm.io/gorm"

	"github.com/lin-snow/ech0/internal/cache"
	model "github.com/lin-snow/ech0/internal/model/common"
	"github.com/lin-snow/ech0/internal/transaction"
)

type KeyValueRepository struct {
	db    func() *gorm.DB
	cache cache.ICache[string, any]
}

func NewKeyValueRepository(dbProvider func() *gorm.DB, cache cache.ICache[string, any]) KeyValueRepositoryInterface {
	return &KeyValueRepository{
		db:    dbProvider,
		cache: cache,
	}
}

// getDB 从上下文中获取事务
func (keyvalueRepository *KeyValueRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(transaction.TxKey).(*gorm.DB); ok {
		return tx
	}
	return keyvalueRepository.db()
}

// GetKeyValue 根据键获取值
func (keyvalueRepository *KeyValueRepository) GetKeyValue(key string) (interface{}, error) {
	// 先查缓存
	if cachedValue, err := keyvalueRepository.cache.Get(key); err == nil {
		// 缓存命中，类型断言
		if value, ok := cachedValue.(string); ok {
			return value, nil
		}
	}

	// 缓存未命中，查询数据库
	var kv model.KeyValue
	if err := keyvalueRepository.db().Where("key = ?", key).First(&kv).Error; err != nil {
		return nil, err
	}

	// 将查询到的值放入缓存
	keyvalueRepository.cache.Set(key, kv.Value, 1)

	return kv.Value, nil
}

// AddKeyValue 添加键值对
func (keyvalueRepository *KeyValueRepository) AddKeyValue(ctx context.Context, key string, value interface{}) error {
	// 清除相关缓存
	keyvalueRepository.cache.Delete(key) // 删除该键的缓存

	if err := keyvalueRepository.getDB(ctx).Create(&model.KeyValue{
		Key:   key,
		Value: value.(string),
	}).Error; err != nil {
		return err
	}

	// 添加新的缓存
	keyvalueRepository.cache.Set(key, value, 1)

	return nil
}

// DeleteKeyValue 删除键值对
func (keyvalueRepository *KeyValueRepository) DeleteKeyValue(ctx context.Context, key string) error {
	// 删除缓存
	keyvalueRepository.cache.Delete(key) // 删除该键的缓存

	if err := keyvalueRepository.getDB(ctx).Where("key = ?", key).Delete(&model.KeyValue{}).Error; err != nil {
		return err
	}

	return nil
}

// UpdateKeyValue 更新键值对
func (keyvalueRepository *KeyValueRepository) UpdateKeyValue(ctx context.Context, key string, value interface{}) error {
	// 更新缓存
	keyvalueRepository.cache.Delete(key) // 删除该键的缓存

	if err := keyvalueRepository.getDB(ctx).Model(&model.KeyValue{}).Where("key = ?", key).Update("value", value.(string)).Error; err != nil {
		return err
	}

	// 添加新的缓存
	keyvalueRepository.cache.Set(key, value, 1)

	return nil
}

// AddOrUpdateKeyValue 添加或更新键值对
func (keyvalueRepository *KeyValueRepository) AddOrUpdateKeyValue(
	ctx context.Context,
	key string,
	value interface{},
) error {
	// 先尝试更新
	result := keyvalueRepository.getDB(ctx).
		Model(&model.KeyValue{}).
		Where("key = ?", key).
		Update("value", value.(string))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		// 如果没有行被更新，说明该键不存在，执行添加操作
		if err := keyvalueRepository.getDB(ctx).Create(&model.KeyValue{
			Key:   key,
			Value: value.(string),
		}).Error; err != nil {
			return err
		}
	}

	// 更新缓存
	keyvalueRepository.cache.Delete(key) // 删除该键的缓存
	keyvalueRepository.cache.Set(key, value, 1)

	return nil
}
