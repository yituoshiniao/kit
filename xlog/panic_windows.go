// +build windows

package xlog

// windows系统暂时不处理
func recordPanic(logPath string, stdout bool) error {
	return nil
}
