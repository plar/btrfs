package subvolume

import (
	"errors"
	"fmt"

	"github.com/plar/btrfs"
	"github.com/plar/btrfs/ioctl"
)

type subvolFindNew struct {
	lastGen uint64
	dest    string

	executor func(c *subvolFindNew) error
}

func (c *subvolFindNew) Destination(dest string) btrfs.SubvolFindNew {
	c.dest = dest
	return c
}

func (c *subvolFindNew) LastGen(lastGen uint64) btrfs.SubvolFindNew {
	c.lastGen = lastGen
	return c
}

func (c *subvolFindNew) context() string {
	return fmt.Sprintf("dest='%s', lastGen=%d", c.dest, c.lastGen)
}

func (c *subvolFindNew) error(err error) *btrfs.BtrfsError {
	return &btrfs.BtrfsError{Func: string(btrfs.CmdSubvolFindNew), Context: c.context(), Err: err}
}

func (c *subvolFindNew) Execute() error {
	return c.executor(c)
}

// btrfs ioctl executor
func ioctlFindNewExecute(c *subvolFindNew) error {
	if len(c.dest) == 0 {
		return fmt.Errorf("Subvolume is required")
	}

	_ /*genId*/, err := ioctl.SubvolFindNew(c.dest, c.lastGen)
	if err != nil {
		return err
	}

	return nil
}

// btrfs cli executor
func cliFindNewExecute(c *subvolFindNew) error {
	return c.error(errors.New("Unimplemented"))
}

// commands
func ioctlFindNew() interface{} {
	return &subvolFindNew{executor: ioctlFindNewExecute}
}

func cliFindNew() interface{} {
	return &subvolFindNew{executor: cliFindNewExecute}
}
