package keyvalue

import "context"

type KeyValueRepositoryInterface interface {
	// GetKeyValue 根据键获取值
	GetKeyValue(key string) (interface{}, error)

	// AddKeyValue 添加键值对
	AddKeyValue(ctx context.Context, key string, value interface{}) error

	// DeleteKeyValue 删除键值对
	DeleteKeyValue(ctx context.Context, key string) error

	// UpdateKeyValue 更新键值对
	UpdateKeyValue(ctx context.Context, key string, value interface{}) error

	// AddOrUpdateKeyValue 添加或更新键值对
	AddOrUpdateKeyValue(ctx context.Context, key string, value interface{}) error
}
