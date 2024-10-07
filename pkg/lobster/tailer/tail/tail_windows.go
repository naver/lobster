//go:build windows
// +build windows

package tail

import (
	"os"

	"github.com/naver/lobster/pkg/lobster/tailer/tail/winfile"
)

func OpenFile(name string) (file *os.File, err error) {
	return winfile.OpenFile(name, os.O_RDONLY, 0)
}
