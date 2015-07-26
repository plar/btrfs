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

type cmdSubvolCreate struct {
	qgroups []string
	dest    string

	executor func(c *cmdSubvolCreate) error
}

func (c *cmdSubvolCreate) QuotaGroups(qgroups ...string) btrfs.CmdSubvolCreate {
	for _, qgroup := range qgroups {
		c.qgroups = append(c.qgroups, qgroup)
	}
	return c
}

func (c *cmdSubvolCreate) Destination(dest string) btrfs.CmdSubvolCreate {
	c.dest = dest
	return c
}

func (c *cmdSubvolCreate) context() string {
	return fmt.Sprintf("qgroups=%v, dest='%s'", c.qgroups, c.dest)
}

func (c *cmdSubvolCreate) error(err error) *btrfs.BtrfsError {
	return &btrfs.BtrfsError{Func: string(btrfs.CmdSubvolumeCreate), Context: c.context(), Err: err}
}

func (c *cmdSubvolCreate) validate() error {
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

func (c *cmdSubvolCreate) Execute() error {
	err := c.executor(c)
	if err != nil {
		return c.error(err)
	}
	return nil
}

// btrfs ioctl executor
func ioctlCreateExecute(c *cmdSubvolCreate) error {
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
func cliCreateExecute(c *cmdSubvolCreate) error {
	err := c.validate()
	if err != nil {
		return err
	}

	return errors.New("Unimplemented")
}

// commands
func ioctlCreate() interface{} {
	return &cmdSubvolCreate{executor: ioctlCreateExecute}
}

func cliCreate() interface{} {
	return &cmdSubvolCreate{executor: cliCreateExecute}
}
