package subvolume

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/plar/btrfs"
	"github.com/plar/btrfs/ioctl"
)

type cmdSubvolDelete struct {
	dest string

	executor func(c *cmdSubvolDelete) error
}

func (c *cmdSubvolDelete) Destination(dest string) btrfs.CmdSubvolDelete {
	c.dest = dest
	return c
}

func (c *cmdSubvolDelete) context() string {
	return fmt.Sprintf("dest='%s'", c.dest)
}

func (c *cmdSubvolDelete) error(err error) *btrfs.BtrfsError {
	return &btrfs.BtrfsError{Func: string(btrfs.CmdSubvolumeDelete), Context: c.context(), Err: err}
}

func (c *cmdSubvolDelete) Execute() error {
	return c.executor(c)
}

// btrfs ioctl executor
func ioctlDeleteExecute(c *cmdSubvolDelete) error {
	if len(c.dest) == 0 {
		return fmt.Errorf("Subvolume is required")
	}

	path := filepath.Dir(c.dest)
	name := filepath.Base(c.dest)

	err := ioctl.SubvolDelete(path, name)
	if err != nil {
		return err
	}

	return nil
}

// btrfs cli executor
func cliDeleteExecute(c *cmdSubvolDelete) error {
	return c.error(errors.New("Unimplemented"))
}

// commands
func ioctlDelete() interface{} {
	return &cmdSubvolDelete{executor: ioctlDeleteExecute}
}

func cliDelete() interface{} {
	return &cmdSubvolDelete{executor: cliDeleteExecute}
}
