//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"os"
	"syscall"
)

func IsWritable(path string) (err error) {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path doesn't exist <%s>", path)
	}

	err = nil
	if !info.IsDir() {
		return fmt.Errorf("path isn't a directory <%s>", path)
	}

	// Check if the user bit is enabled in file permission
	if info.Mode().Perm()&(1<<(uint(7))) == 0 {
		return fmt.Errorf("write permission bit is not set on this file for user <%s>", path)
	}

	var stat syscall.Stat_t
	if err = syscall.Stat(path, &stat); err != nil {
		return fmt.Errorf("Unable to get stat for path <%s>", path)
	}

	err = nil
	if uint32(os.Geteuid()) != stat.Uid {
		return fmt.Errorf("user doesn't have permission to write to this directory <%s>", path)
	}
	return nil
}
