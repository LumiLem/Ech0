package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/lin-snow/ech0/internal/cache"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/echo"
	"github.com/lin-snow/ech0/internal/transaction"
)

type EchoRepository struct {
	db    func() *gorm.DB
	cache cache.ICache[string, any]
}

func NewEchoRepository(dbProvider func() *gorm.DB, cache cache.ICache[string, any]) EchoRepositoryInterface {
	return &EchoRepository{db: dbProvider, cache: cache}
}

// getDB 从上下文中获取事务
func (echoRepository *EchoRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(transaction.TxKey).(*gorm.DB); ok {
		return tx
	}
	return echoRepository.db()
}

// CreateEcho 创建新的 Echo
func (echoRepository *EchoRepository) CreateEcho(ctx context.Context, echo *model.Echo) error {
	echo.Content = strings.TrimSpace(echo.Content)

	result := echoRepository.getDB(ctx).Create(echo)
	if result.Error != nil {
		return result.Error
	}

	// 清除相关缓存
	ClearEchoPageCache(echoRepository.cache)
	echoRepository.cache.Delete(GetTodayEchosCacheKey(true))  // 删除今天的 Echo 缓存（管理员视图）
	echoRepository.cache.Delete(GetTodayEchosCacheKey(false)) // 删除今天的 Echo 缓存（非管理员视图）

	return nil
}

// GetEchosByPage 获取分页的 Echo 列表
func (echoRepository *EchoRepository) GetEchosByPage(
	page, pageSize int,
	search string,
	showPrivate bool,
) ([]model.Echo, int64) {
	// 查找缓存
	cacheKey := GetEchoPageCacheKey(page, pageSize, search, showPrivate)
	if cachedResult, err := echoRepository.cache.Get(cacheKey); err == nil {
		// 缓存命中，直接返回
		// 类型断言
		cachedResultTyped, ok := cachedResult.(commonModel.PageQueryResult[[]model.Echo])
		if ok {
			return cachedResultTyped.Items, cachedResultTyped.Total
		}
	}

	// 如果缓存未命中，进行数据库查询

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 查询数据库
	var echos []model.Echo
	var total int64

	query := echoRepository.db().Model(&model.Echo{})

	// 如果 search 不为空，添加模糊查询条件
	if search != "" {
		searchPattern := "%" + search + "%" // 模糊匹配模式
		query = query.Where("content LIKE ?", searchPattern)
	}

	// 如果不是管理员，过滤私密Echo
	if !showPrivate {
		query = query.Where("private = ?", false)
	}

	// 获取总数并进行分页查询
	query.Count(&total).
		Preload("Media").
		Preload("Tags").
		Joins("User").
		Limit(pageSize).
		Offset(offset).
		Order("created_at DESC").
		Find(&echos)

	// 保存到缓存
	echoKeyList = append(echoKeyList, cacheKey) // 记录缓存键
	echoRepository.cache.Set(cacheKey, commonModel.PageQueryResult[[]model.Echo]{
		Items: echos,
		Total: total,
	}, 1)

	// 返回结果
	return echos, total
}

// GetEchosById 根据 ID 获取 Echo
func (echoRepository *EchoRepository) GetEchosById(id uint) (*model.Echo, error) {
	// 查询缓存
	cacheKey := GetEchoByIDCacheKey(id)
	if cachedEcho, err := echoRepository.cache.Get(cacheKey); err == nil {
		// 缓存命中，直接返回
		if echo, ok := cachedEcho.(*model.Echo); ok {
			return echo, nil
		}
	}

	// 缓存未命中，查询数据库
	// 使用 Preload 预加载关联的 Media，使用 Joins 关联 User
	var echo model.Echo
	result := echoRepository.db().Preload("Media").Preload("Tags").Joins("User").First(&echo, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // 如果未找到记录，则返回 nil
		}
		return nil, result.Error // 其他错误返回
	}

	// 保存到缓存
	echoRepository.cache.Set(cacheKey, &echo, 1)

	return &echo, nil
}

// DeleteEchoById 删除 Echo
func (echoRepository *EchoRepository) DeleteEchoById(ctx context.Context, id uint) error {
	var echo model.Echo
	// 删除外键media
	echoRepository.getDB(ctx).Where("message_id = ?", id).Delete(&model.Media{})

	result := echoRepository.getDB(ctx).Delete(&echo, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // 如果没有找到记录
	}

	// 清除缓存
	echoRepository.cache.Delete(GetEchoByIDCacheKey(id))      // 删除具体 Echo 的缓存
	echoRepository.cache.Delete(GetTodayEchosCacheKey(true))  // 删除今天的 Echo 缓存（管理员视图）
	echoRepository.cache.Delete(GetTodayEchosCacheKey(false)) // 删除今天的 Echo 缓存（非管理员视图）

	// 清除相关缓存
	ClearEchoPageCache(echoRepository.cache)

	return nil
}

// GetTodayEchos 获取今天的 Echo 列表
func (echoRepository *EchoRepository) GetTodayEchos(showPrivate bool) []model.Echo {
	// 查找缓存
	if cachedTodayEchos, err := echoRepository.cache.Get(GetTodayEchosCacheKey(showPrivate)); err == nil {
		// 缓存命中，直接返回
		if todayEchos, ok := cachedTodayEchos.([]model.Echo); ok {
			return todayEchos
		}
	}

	// 查询数据库
	var echos []model.Echo

	// 获取当天开始和结束时间
	today := time.Now()
	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := echoRepository.db().Model(&model.Echo{})
	// 如果不是管理员，过滤私密Echo
	if !showPrivate {
		query = query.Where("private = ?", false)
	}

	// 添加当天的时间过滤
	query = query.Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay)

	// 获取总数并进行分页查询
	query.
		Preload("Media").
		Preload("Tags").
		Joins("User").
		Order("created_at DESC").
		Find(&echos)

	// 保存到缓存
	echoRepository.cache.Set(GetTodayEchosCacheKey(showPrivate), echos, 1)

	// 返回结果
	return echos
}

// UpdateEcho 更新 Echo
func (echoRepository *EchoRepository) UpdateEcho(ctx context.Context, echo *model.Echo) error {
	// 清空缓存
	ClearEchoPageCache(echoRepository.cache)
	echoRepository.cache.Delete(GetEchoByIDCacheKey(echo.ID)) // 删除具体 Echo 的缓存
	echoRepository.cache.Delete(GetTodayEchosCacheKey(true))  // 删除今天的 Echo 缓存（管理员视图）
	echoRepository.cache.Delete(GetTodayEchosCacheKey(false)) // 删除今天的 Echo 缓存（非管理员视图）

	// 1. 获取现有媒体列表
	var existingMedia []model.Media
	if err := echoRepository.getDB(ctx).Where("message_id = ?", echo.ID).Find(&existingMedia).Error; err != nil {
		return err
	}

	// 2. 构建现有媒体的 URL -> Media 映射
	existingMediaMap := make(map[string]*model.Media)
	for i := range existingMedia {
		existingMediaMap[existingMedia[i].MediaURL] = &existingMedia[i]
	}

	// 3. 构建新媒体的 URL 集合
	newMediaURLs := make(map[string]bool)
	for _, m := range echo.Media {
		if m.MediaURL != "" {
			newMediaURLs[m.MediaURL] = true
		}
	}

	// 4. 删除不再需要的媒体（在新列表中不存在的）
	for url, media := range existingMediaMap {
		if !newMediaURLs[url] {
			if err := echoRepository.getDB(ctx).Delete(media).Error; err != nil {
				return err
			}
		}
	}

	// 5. 更新 Echo 内容
	if err := echoRepository.getDB(ctx).Model(&model.Echo{}).
		Where("id = ?", echo.ID).
		Updates(map[string]interface{}{
			"content":        echo.Content,
			"private":        echo.Private,
			"layout":         echo.Layout,
			"extension":      echo.Extension,
			"extension_type": echo.ExtensionType,
		}).Error; err != nil {
		return err
	}

	// 6. 检查是否只是顺序变化（媒体集合相同，但顺序不同）
	orderChanged := false
	if len(echo.Media) == len(existingMedia) {
		// 检查 URL 集合是否完全相同
		allExist := true
		for _, newMedia := range echo.Media {
			if _, exists := existingMediaMap[newMedia.MediaURL]; !exists {
				allExist = false
				break
			}
		}

		// 如果集合相同，检查顺序是否不同
		if allExist {
			for i := range echo.Media {
				if echo.Media[i].MediaURL != existingMedia[i].MediaURL {
					orderChanged = true
					break
				}
			}
		}
	}

	// 7. 处理媒体
	if orderChanged {
		// 顺序变化：交换数据而不是删除重建，保持原有 ID 不变
		//
		// 原逻辑（已注释，供备用）：
		// // 顺序变化：删除所有并按新顺序重新插入
		// if err := echoRepository.getDB(ctx).Where("message_id = ?", echo.ID).Delete(&model.Media{}).Error; err != nil {
		// 	return err
		// }
		// // 构建实况照片关联映射：图片URL -> 视频URL
		// livePhotoMap := make(map[string]string)
		// for _, m := range echo.Media {
		// 	if existing, ok := existingMediaMap[m.MediaURL]; ok {
		// 		if existing.LiveVideoID != nil && *existing.LiveVideoID > 0 {
		// 			for _, oldMedia := range existingMedia {
		// 				if oldMedia.ID == *existing.LiveVideoID {
		// 					livePhotoMap[m.MediaURL] = oldMedia.MediaURL
		// 					break
		// 				}
		// 			}
		// 		}
		// 	}
		// }
		// // 插入所有媒体
		// newMediaMap := make(map[string]*model.Media)
		// for i := range echo.Media {
		// 	echo.Media[i].MessageID = echo.ID
		// 	echo.Media[i].ID = 0
		// 	echo.Media[i].LiveVideoID = nil
		// 	if echo.Media[i].MediaURL != "" {
		// 		if err := echoRepository.getDB(ctx).Create(&echo.Media[i]).Error; err != nil {
		// 			return err
		// 		}
		// 		newMediaMap[echo.Media[i].MediaURL] = &echo.Media[i]
		// 	}
		// }
		// // 更新实况照片关联
		// for imageURL, videoURL := range livePhotoMap {
		// 	if imageMedia, ok := newMediaMap[imageURL]; ok {
		// 		if videoMedia, ok := newMediaMap[videoURL]; ok {
		// 			imageMedia.LiveVideoID = &videoMedia.ID
		// 			if err := echoRepository.getDB(ctx).Model(&model.Media{}).
		// 				Where("id = ?", imageMedia.ID).
		// 				Update("live_video_id", videoMedia.ID).Error; err != nil {
		// 				return err
		// 			}
		// 		}
		// 	}
		// }

		// 新逻辑：交换数据，保持 ID 不变
		// 1. 构建新顺序的 URL 列表
		newOrder := make([]string, 0, len(echo.Media))
		for _, m := range echo.Media {
			if m.MediaURL != "" {
				newOrder = append(newOrder, m.MediaURL)
			}
		}

		// 2. 构建旧顺序的 URL 列表和 ID 列表
		oldOrder := make([]string, 0, len(existingMedia))
		oldIDs := make([]uint, 0, len(existingMedia))
		for _, m := range existingMedia {
			oldOrder = append(oldOrder, m.MediaURL)
			oldIDs = append(oldIDs, m.ID)
		}

		// 3. 保存实况照片关联信息（URL -> 视频URL）
		livePhotoRelations := make(map[string]string)
		for _, m := range existingMedia {
			if m.LiveVideoID != nil && *m.LiveVideoID > 0 {
				for _, oldMedia := range existingMedia {
					if oldMedia.ID == *m.LiveVideoID {
						livePhotoRelations[m.MediaURL] = oldMedia.MediaURL
						break
					}
				}
			}
		}

		// 4. 按新顺序更新每个媒体记录的数据（保持 ID 不变）
		for i, newURL := range newOrder {
			if i >= len(oldIDs) {
				break
			}
			targetID := oldIDs[i]

			// 获取新 URL 对应的原始媒体数据
			sourceMedia, ok := existingMediaMap[newURL]
			if !ok {
				continue
			}

			// 更新目标 ID 的记录，使用源媒体的数据
			updateData := map[string]interface{}{
				"media_url":    sourceMedia.MediaURL,
				"media_type":   sourceMedia.MediaType,
				"media_source": sourceMedia.MediaSource,
				"object_key":   sourceMedia.ObjectKey,
				"width":        sourceMedia.Width,
				"height":       sourceMedia.Height,
			}

			if err := echoRepository.getDB(ctx).Model(&model.Media{}).
				Where("id = ?", targetID).
				Updates(updateData).Error; err != nil {
				return err
			}
		}

		// 5. 重新建立实况照片关联（基于新的 URL 位置）
		// 先清除所有关联
		if err := echoRepository.getDB(ctx).Model(&model.Media{}).
			Where("message_id = ?", echo.ID).
			Update("live_video_id", nil).Error; err != nil {
			return err
		}

		// 重新查询更新后的媒体列表
		var updatedMedia []model.Media
		if err := echoRepository.getDB(ctx).Where("message_id = ?", echo.ID).Find(&updatedMedia).Error; err != nil {
			return err
		}

		// 构建新的 URL -> ID 映射
		newURLToID := make(map[string]uint)
		for _, m := range updatedMedia {
			newURLToID[m.MediaURL] = m.ID
		}

		// 恢复实况照片关联
		for imageURL, videoURL := range livePhotoRelations {
			imageID, imageOK := newURLToID[imageURL]
			videoID, videoOK := newURLToID[videoURL]
			if imageOK && videoOK {
				if err := echoRepository.getDB(ctx).Model(&model.Media{}).
					Where("id = ?", imageID).
					Update("live_video_id", videoID).Error; err != nil {
					return err
				}
			}
		}
	} else {
		// 保留现有的，新增新的（原逻辑）
		for i := range echo.Media {
			echo.Media[i].MessageID = echo.ID
			if existing, ok := existingMediaMap[echo.Media[i].MediaURL]; ok {
				// 媒体已存在，保留原有 ID 和关联信息
				echo.Media[i].ID = existing.ID
				echo.Media[i].LiveVideoID = existing.LiveVideoID
			} else if echo.Media[i].MediaURL != "" {
				// 新媒体，插入数据库
				if err := echoRepository.getDB(ctx).Create(&echo.Media[i]).Error; err != nil {
					return err
				}
			}
		}
	}

	// 8. 更新标签关联关系
	if err := echoRepository.getDB(ctx).Model(echo).Association("Tags").Replace(echo.Tags); err != nil {
		return err
	}

	return nil
}

// LikeEcho 点赞 Echo
func (echoRepository *EchoRepository) LikeEcho(ctx context.Context, id uint) error {
	// 检查是否存在（可选，防止无效点赞）
	var exists bool
	if err := echoRepository.getDB(ctx).
		Model(&model.Echo{}).
		Select("count(*) > 0").
		Where("id = ?", id).
		Find(&exists).Error; err != nil {
		return err
	}
	if !exists {
		return errors.New(commonModel.ECHO_NOT_FOUND)
	}

	// 原子自增点赞数
	if err := echoRepository.getDB(ctx).
		Model(&model.Echo{}).
		Where("id = ?", id).
		UpdateColumn("fav_count", gorm.Expr("fav_count + ?", 1)).Error; err != nil {
		return err
	}

	// 清除相关缓存
	ClearEchoPageCache(echoRepository.cache)
	echoRepository.cache.Delete(GetEchoByIDCacheKey(id))      // 删除具体 Echo 的缓存
	echoRepository.cache.Delete(GetTodayEchosCacheKey(true))  // 删除今天的 Echo 缓存（管理员视图）
	echoRepository.cache.Delete(GetTodayEchosCacheKey(false)) // 删除今天的 Echo 缓存（非管理员视图）

	return nil
}

// GetAllTags 获取所有标签
func (echoRepository *EchoRepository) GetAllTags() ([]model.Tag, error) {
	var tags []model.Tag
	result := echoRepository.db().Order("usage_count DESC, created_at DESC").Find(&tags)
	if result.Error != nil {
		return nil, result.Error
	}
	return tags, nil
}

// DeleteTagById 删除标签
func (echoRepository *EchoRepository) DeleteTagById(ctx context.Context, id uint) error {
	var tag model.Tag

	// 删除关联的 EchoTag 关系
	if err := echoRepository.getDB(ctx).Where("tag_id = ?", id).Delete(&model.EchoTag{}).Error; err != nil {
		return err
	}

	// 删除标签
	result := echoRepository.getDB(ctx).Delete(&tag, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // 如果没有找到记录
	}

	return nil
}

// GetTagByName 根据名称获取标签
func (echoRepository *EchoRepository) GetTagByName(name string) (*model.Tag, error) {
	var tag model.Tag
	result := echoRepository.db().Where("name = ?", name).First(&tag)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // 如果未找到记录，则返回 nil
		}
		return nil, result.Error // 其他错误返回
	}
	return &tag, nil
}

// GetTagsByNames 根据名称列表获取标签
func (echoRepository *EchoRepository) GetTagsByNames(names []string) ([]*model.Tag, error) {
	var tags []*model.Tag
	result := echoRepository.db().Where("name IN ?", names).Find(&tags)
	if result.Error != nil {
		return nil, result.Error
	}
	return tags, nil
}

// CreateTag 创建标签
func (echoRepository *EchoRepository) CreateTag(ctx context.Context, tag *model.Tag) error {
	tag.Name = strings.TrimSpace(tag.Name)
	if tag.Name == "" {
		return errors.New("标签名称不能为空")
	}

	result := echoRepository.getDB(ctx).Create(tag)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// IncrementTagUsageCount 增加标签的使用计数
func (echoRepository *EchoRepository) IncrementTagUsageCount(ctx context.Context, tagID uint) error {
	return echoRepository.getDB(ctx).Model(&model.Tag{}).
		Where("id = ?", tagID).
		UpdateColumn("usage_count", gorm.Expr("usage_count + ?", 1)).Error
}

// GetEchosByTagId 根据标签ID获取关联的 Echo 列表
func (echoRepository *EchoRepository) GetEchosByTagId(
	tagId uint,
	page, pageSize int,
	search string,
	showPrivate bool,
) ([]model.Echo, int64, error) {
	var (
		echos []model.Echo
		total int64
	)

	applyFilters := func(db *gorm.DB) *gorm.DB {
		db = db.Joins("JOIN echo_tags ON echo_tags.echo_id = echos.id").
			Where("echo_tags.tag_id = ?", tagId)

		if !showPrivate {
			db = db.Where("echos.private = ?", false)
		}

		if search != "" {
			db = db.Where("echos.content LIKE ?", "%"+search+"%")
		}

		return db
	}

	countQuery := applyFilters(echoRepository.db().Model(&model.Echo{}))

	if err := countQuery.Distinct("echos.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize

	var echoIDs []uint
	idsQuery := applyFilters(echoRepository.db().Model(&model.Echo{}))
	if err := idsQuery.
		Distinct("echos.id").
		Order("echos.created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Pluck("echos.id", &echoIDs).Error; err != nil {
		return nil, 0, err
	}

	if len(echoIDs) == 0 {
		return []model.Echo{}, total, nil
	}

	if err := echoRepository.db().
		Where("echos.id IN ?", echoIDs).
		Preload("Media").
		Preload("Tags").
		Joins("User").
		Order("created_at DESC").
		Find(&echos).Error; err != nil {
		return nil, 0, err
	}

	return echos, total, nil
}

// UpdateMediaLiveVideoID 更新媒体的实况照片关联
func (echoRepository *EchoRepository) UpdateMediaLiveVideoID(ctx context.Context, mediaID uint, liveVideoID uint) error {
	return echoRepository.getDB(ctx).Model(&model.Media{}).
		Where("id = ?", mediaID).
		Update("live_video_id", liveVideoID).Error
}

// GetMediaByID 根据 ID 获取媒体
func (echoRepository *EchoRepository) GetMediaByID(id uint) (*model.Media, error) {
	var media model.Media
	result := echoRepository.db().First(&media, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &media, nil
}

// IsLivePhotoVideo 检查视频是否是实况照片的一部分
func (echoRepository *EchoRepository) IsLivePhotoVideo(videoID uint) (bool, error) {
	var count int64
	result := echoRepository.db().Model(&model.Media{}).
		Where("live_video_id = ?", videoID).
		Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}
