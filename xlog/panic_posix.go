// +build !windows

package xlog

import (
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"syscall"
)

const panicFileName = "/panic.log"

// 系统发生panic即会输出到panic文件（recover的panic也会输出）
// windows系统暂时只会创建panic
func recordPanic(logPath string, stdout bool) error {
	if logPath == "" {
		return nil
	}
	err := os.MkdirAll(filepath.Dir(logPath), 0644)
	if err != nil {
		return errors.Wrap(err, "创建目录失败")
	}

	f, err := os.OpenFile(filepath.Join(filepath.Dir(logPath), panicFileName), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return errors.Wrap(err, "创建panic.log失败")
	}

	if stdout {
		return nil //错误输出到控制台
	}
	return errors.WithStack(syscall.Dup2(int(f.Fd()), int(os.Stderr.Fd())))
}
