package btrfs

import (
	"fmt"
	"strings"
)

const (
	BtrfsVolNameMax = 255
)

type BtrfsError struct {
	Func    string
	Err     error
	Context string
}

func (e *BtrfsError) Error() string {
	return fmt.Sprintf("ERROR %s: %s, args=(%s)", e.Func, e.Err, e.Context)
}

type Executor interface {
	Execute() error
}

type Btrfs interface {
}

func TestSubvolumeName(name string) bool {
	return len(name) > 0 && strings.Index(name, "/") == -1 && name != "." && name != ".."
}
