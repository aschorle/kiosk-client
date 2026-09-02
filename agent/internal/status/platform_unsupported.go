//go:build !linux

package status

import "runtime"

func platformKernel() string { return "" }

func platformArchitecture() string { return runtime.GOARCH }

func platformDiskSpace(string) (uint64, uint64) { return 0, 0 }
