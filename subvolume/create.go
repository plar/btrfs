package subvolume

import (
	"errors"
	"fmt"
	"path/filepath"

	"os"

	"github.com/plar/btrfs"
	"github.com/plar/btrfs/ioctl"
	"github.com/plar/btrfs/validators"
)

type subvolCreate struct {
	qgroups []string
	dest    string

	executor func(c *subvolCreate) error
}

func (c *subvolCreate) QuotaGroups(qgroups ...string) btrfs.SubvolCreate {
	for _, qgroup := range qgroups {
		c.qgroups = append(c.qgroups, qgroup)
	}
	return c
}

func (c *subvolCreate) Destination(dest string) btrfs.SubvolCreate {
	c.dest = dest
	return c
}

func (c *subvolCreate) context() string {
	return fmt.Sprintf("qgroups=%v, dest='%s'", c.qgroups, c.dest)
}

func (c *subvolCreate) error(err error) *btrfs.BtrfsError {
	return &btrfs.BtrfsError{Func: string(btrfs.CmdSubvolCreate), Context: c.context(), Err: err}
}

func (c *subvolCreate) validate() error {
	if len(c.dest) == 0 {
		return errors.New("destination is empty")
	}

	name := filepath.Base(c.dest)
	err := validators.ValidSubvolumeName(name)
	if err != nil {
		return err
	}

	fi, err := os.Stat(c.dest)
	if err == nil && fi.IsDir() {
		return fmt.Errorf("'%s' exists", c.dest)
	}

	return nil
}

func (c *subvolCreate) Execute() error {
	err := c.executor(c)
	if err != nil {
		return c.error(err)
	}
	return nil
}

// btrfs ioctl executor
func ioctlCreateExecute(c *subvolCreate) error {
	err := c.validate()
	if err != nil {
		return err
	}

	var dest, name string
	dest = filepath.Dir(c.dest)
	name = filepath.Base(c.dest)

	err = ioctl.SubvolCreate(dest, name)
	if err != nil {
		return err
	}

	return nil
}

// btrfs cli executor
func cliCreateExecute(c *subvolCreate) error {
	err := c.validate()
	if err != nil {
		return err
	}

	return errors.New("Unimplemented")
}

// commands
func ioctlCreate() interface{} {
	return &subvolCreate{executor: ioctlCreateExecute}
}

func cliCreate() interface{} {
	return &subvolCreate{executor: cliCreateExecute}
}
