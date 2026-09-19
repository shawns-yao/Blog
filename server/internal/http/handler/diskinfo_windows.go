//go:build windows

package handler

import (
	"path/filepath"

	"golang.org/x/sys/windows"
)

// getDiskInfo 通过 GetDiskFreeSpaceEx 获取指定路径所在卷的容量信息。
func getDiskInfo(path string) DiskInfo {
	abs, err := filepath.Abs(path)
	if err != nil {
		return DiskInfo{Path: path}
	}
	// GetDiskFreeSpaceEx 需要卷根目录，例如 "C:\"
	volume := filepath.VolumeName(abs)
	if volume == "" {
		return DiskInfo{Path: path}
	}
	root, err := windows.UTF16PtrFromString(volume + `\`)
	if err != nil {
		return DiskInfo{Path: path}
	}

	var freeToCaller, total, free uint64
	if err := windows.GetDiskFreeSpaceEx(root, &freeToCaller, &total, &free); err != nil {
		return DiskInfo{Path: path}
	}

	return DiskInfo{
		Path: path,
		All:  total,
		Used: total - free,
		Free: free,
	}
}
