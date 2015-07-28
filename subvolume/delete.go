package subvolume

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/plar/btrfs"
	"github.com/plar/btrfs/ioctl"
)

type subvolDelete struct {
	dest string

	executor func(c *subvolDelete) error
}

func (c *subvolDelete) Destination(dest string) btrfs.SubvolDelete {
	c.dest = dest
	return c
}

func (c *subvolDelete) context() string {
	return fmt.Sprintf("dest='%s'", c.dest)
}

func (c *subvolDelete) error(err error) *btrfs.BtrfsError {
	return &btrfs.BtrfsError{Func: string(btrfs.CmdSubvolDelete), Context: c.context(), Err: err}
}

func (c *subvolDelete) Execute() error {
	return c.executor(c)
}

// btrfs ioctl executor
func ioctlDeleteExecute(c *subvolDelete) error {
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
func cliDeleteExecute(c *subvolDelete) error {
	return c.error(errors.New("Unimplemented"))
}

// commands
func ioctlDelete() interface{} {
	return &subvolDelete{executor: ioctlDeleteExecute}
}

func cliDelete() interface{} {
	return &subvolDelete{executor: cliDeleteExecute}
}
