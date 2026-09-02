//go:build linux

package status

import "syscall"

func platformKernel() string {
	var uts syscall.Utsname
	if err := syscall.Uname(&uts); err != nil {
		return ""
	}
	return charsToString(uts.Release[:])
}

func platformArchitecture() string {
	var uts syscall.Utsname
	if err := syscall.Uname(&uts); err == nil {
		return charsToString(uts.Machine[:])
	}
	return ""
}

func platformDiskSpace(path string) (uint64, uint64) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0
	}
	blockSize := uint64(stat.Bsize)
	return stat.Blocks * blockSize, stat.Bavail * blockSize
}
