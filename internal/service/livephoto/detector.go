package livephoto

import (
	"path/filepath"
	"strings"

	model "github.com/lin-snow/ech0/internal/model/echo"
)

// ProcessLivePhotoPairs 处理实况照片关联
// 根据 pairIDs 建立图片和视频的关联
// 参数 media 是媒体数组，pairIDs 是与 media 数组对应的 live_pair_id 列表
// 返回: map[imageIndex]videoIndex - 图片索引到视频索引的映射
func ProcessLivePhotoPairs(media []model.Media, pairIDs []string) map[int]int {
	pairs := make(map[int]int)

	// 按 LivePairID 分组
	pairGroups := make(map[string][]int) // pairID -> media indexes
	for i, pairID := range pairIDs {
		if pairID != "" && i < len(media) {
			pairGroups[pairID] = append(pairGroups[pairID], i)
		}
	}

	// 在每组中找到图片和视频，建立关联
	for _, indexes := range pairGroups {
		var imageIdx, videoIdx int = -1, -1
		for _, idx := range indexes {
			if media[idx].MediaType == model.MediaTypeImage {
				imageIdx = idx
			} else if media[idx].MediaType == model.MediaTypeVideo {
				videoIdx = idx
			}
		}
		if imageIdx >= 0 && videoIdx >= 0 {
			pairs[imageIdx] = videoIdx
		}
	}

	return pairs
}

// GetBaseName 获取文件基础名（不含扩展名和路径）
// 用于前端检测同名文件对时的辅助函数
func GetBaseName(urlOrPath string) string {
	// 获取文件名（去除路径）
	filename := filepath.Base(urlOrPath)

	// 处理 URL 中可能的查询参数
	if idx := strings.Index(filename, "?"); idx != -1 {
		filename = filename[:idx]
	}

	// 去除扩展名
	ext := filepath.Ext(filename)
	if ext != "" {
		filename = filename[:len(filename)-len(ext)]
	}

	return strings.ToLower(filename)
}
