package utils

import (
	"fmt"
	"syscall"
)

// CheckDiskSpace checks if there's enough disk space at the given path
// requiredBytes specifies the minimum required free space in bytes
func CheckDiskSpace(path string, requiredBytes uint64) error {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return fmt.Errorf("failed to get disk space info: %w", err)
	}

	// Available blocks * block size = available bytes
	available := stat.Bavail * uint64(stat.Bsize)

	if available < requiredBytes {
		return fmt.Errorf("insufficient disk space: %d bytes available, %d bytes required",
			available, requiredBytes)
	}

	return nil
}

// GetAvailableDiskSpace returns the available disk space in bytes at the given path
func GetAvailableDiskSpace(path string) (uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, fmt.Errorf("failed to get disk space info: %w", err)
	}

	available := stat.Bavail * uint64(stat.Bsize)
	return available, nil
}

// FormatBytes formats bytes into human-readable format (KB, MB, GB, etc.)
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}
