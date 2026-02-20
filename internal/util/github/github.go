package util

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v80/github"
	"golang.org/x/mod/semver"
)

var latestVersionCache struct {
	mu        sync.Mutex
	version   string
	expiresAt time.Time
}

// GetLatestVersion 获取最新版本（从 custom 版仓库）
func GetLatestVersion() (string, error) {
	// 规范化 semver 标签（支持 -custom.X 格式）
	normalizeSemver := func(tag string) string {
		t := strings.TrimSpace(tag)
		if t == "" {
			return ""
		}
		if !strings.HasPrefix(t, "v") {
			t = "v" + t
		}
		t = semver.Canonical(t)
		if t == "" {
			return ""
		}
		return t
	}

	// 获取最新版本
	now := time.Now().UTC()
	latestVersionCache.mu.Lock()
	if latestVersionCache.version != "" && now.Before(latestVersionCache.expiresAt) {
		v := latestVersionCache.version
		latestVersionCache.mu.Unlock()
		return v, nil
	}
	latestVersionCache.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := github.NewClient(nil)
	// 检测 custom 版仓库的更新
	rel, _, err := client.Repositories.GetLatestRelease(ctx, "LumiLem", "Ech0")
	if err != nil {
		return "", fmt.Errorf("get latest release failed: %w", err)
	}

	tag := strings.TrimSpace(rel.GetTagName())
	best := normalizeSemver(tag)
	if best == "" {
		return "", fmt.Errorf("invalid semver tag from latest release: %q", tag)
	}

	// 保持与 commonModel.FullVersion 一致：返回不带 v 的版本号
	result := strings.TrimPrefix(best, "v")

	latestVersionCache.mu.Lock()
	latestVersionCache.version = result
	latestVersionCache.expiresAt = time.Now().UTC().Add(30 * time.Minute)
	latestVersionCache.mu.Unlock()

	return result, nil
}
