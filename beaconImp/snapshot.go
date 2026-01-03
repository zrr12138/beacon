package beaconImp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"beacon/log"
)

const (
	snapshotDir   = "./data/snapshots"
	latestSymlink = "./data/latest"
	maxSnapshots  = 10 // 保留最近10个快照
)

// ========== Snapshot I/O - 快照持久化管理 ==========

// SaveSnapshot 保存当前游戏状态到快照文件
// 注意：调用者需持有读锁（保证状态一致性）
func (B *Beacon) SaveSnapshot() error {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("snapshot_%s.json", timestamp)
	filePath := filepath.Join(snapshotDir, filename)

	// 序列化状态
	data, err := json.MarshalIndent(B.state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	// 写入临时文件
	tmpFile := filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("write tmp file: %w", err)
	}

	// 原子性重命名
	if err := os.Rename(tmpFile, filePath); err != nil {
		return fmt.Errorf("rename snapshot: %w", err)
	}

	// 更新 latest 软链
	if err := updateLatestSymlink(filePath); err != nil {
		log.Warnf("Failed to update latest symlink: %v", err)
	}

	// 清理旧快照
	if err := rotateSnapshots(); err != nil {
		log.Warnf("Failed to rotate snapshots: %v", err)
	}

	log.Infof("Snapshot saved: %s", filename)
	return nil
}

// LoadLatestSnapshot 加载最新快照到内存
func (B *Beacon) LoadLatestSnapshot() error {
	// 检查 latest 软链是否存在
	latestPath, err := os.Readlink(latestSymlink)
	if err != nil {
		if os.IsNotExist(err) {
			log.Info("No existing snapshot found, starting with empty state")
			B.state = NewGameState()
			return nil
		}
		return fmt.Errorf("read latest symlink: %w", err)
	}

	// 读取快照文件
	data, err := os.ReadFile(latestPath)
	if err != nil {
		return fmt.Errorf("read snapshot: %w", err)
	}

	// 反序列化
	state := NewGameState()
	if err := json.Unmarshal(data, state); err != nil {
		return fmt.Errorf("unmarshal snapshot: %w", err)
	}

	B.state = state
	log.Infof("Loaded snapshot from: %s", latestPath)
	return nil
}

// updateLatestSymlink 更新 latest 软链指向最新快照
func updateLatestSymlink(targetPath string) error {
	// 删除旧软链
	os.Remove(latestSymlink)

	// 创建新软链
	return os.Symlink(targetPath, latestSymlink)
}

// rotateSnapshots 清理旧快照，保留最近N个
func rotateSnapshots() error {
	entries, err := os.ReadDir(snapshotDir)
	if err != nil {
		return err
	}

	// 过滤出快照文件
	var snapshots []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			snapshots = append(snapshots, e.Name())
		}
	}

	// 按时间戳排序（文件名已包含时间戳）
	sort.Strings(snapshots)

	// 删除多余的旧快照
	if len(snapshots) > maxSnapshots {
		for i := 0; i < len(snapshots)-maxSnapshots; i++ {
			oldPath := filepath.Join(snapshotDir, snapshots[i])
			if err := os.Remove(oldPath); err != nil {
				log.Warnf("Failed to remove old snapshot %s: %v", snapshots[i], err)
			} else {
				log.Infof("Removed old snapshot: %s", snapshots[i])
			}
		}
	}

	return nil
}
