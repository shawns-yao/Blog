//go:build unix

package handler

import "golang.org/x/sys/unix"

// getDiskInfo 通过 Statfs 获取指定路径所在文件系统的容量信息。
func getDiskInfo(path string) DiskInfo {
	var stat unix.Statfs_t
	if err := unix.Statfs(path, &stat); err != nil {
		return DiskInfo{Path: path}
	}

	// Bsize 为文件系统的块大小，部分文件系统可能返回 0，此时容量字段将全为 0
	bsize := uint64(stat.Bsize)
	all := stat.Blocks * bsize
	free := stat.Bfree * bsize

	return DiskInfo{
		Path: path,
		All:  all,
		Used: all - free,
		Free: free,
	}
}
