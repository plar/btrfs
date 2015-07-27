package subvolume

import (
	"errors"
	"fmt"

	"github.com/plar/btrfs"
	"github.com/plar/btrfs/ioctl"
)

type cmdSubvolFindNew struct {
	lastGen uint64
	dest    string

	executor func(c *cmdSubvolFindNew) error
}

func (c *cmdSubvolFindNew) Destination(dest string) btrfs.CmdSubvolFindNew {
	c.dest = dest
	return c
}

func (c *cmdSubvolFindNew) LastGen(lastGen uint64) btrfs.CmdSubvolFindNew {
	c.lastGen = lastGen
	return c
}

func (c *cmdSubvolFindNew) context() string {
	return fmt.Sprintf("dest='%s', lastGen=%d", c.dest, c.lastGen)
}

func (c *cmdSubvolFindNew) error(err error) *btrfs.BtrfsError {
	return &btrfs.BtrfsError{Func: string(btrfs.CmdSubvolumeFindNew), Context: c.context(), Err: err}
}

func (c *cmdSubvolFindNew) Execute() error {
	return c.executor(c)
}

// btrfs ioctl executor
func ioctlFindNewExecute(c *cmdSubvolFindNew) error {
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
func cliFindNewExecute(c *cmdSubvolFindNew) error {
	return c.error(errors.New("Unimplemented"))
}

// commands
func ioctlFindNew() interface{} {
	return &cmdSubvolFindNew{executor: ioctlFindNewExecute}
}

func cliFindNew() interface{} {
	return &cmdSubvolFindNew{executor: cliFindNewExecute}
}
