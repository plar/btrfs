package subvolume

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"os"

	"github.com/plar/btrfs"
	"github.com/plar/btrfs/ioctl"
	"github.com/plar/btrfs/validators"
)

type cmdSubvolCreate struct {
	qgroups []string
	dest    string
	name    string

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

func (c *cmdSubvolCreate) Name(name string) btrfs.CmdSubvolCreate {
	c.name = name
	return c
}

func (c *cmdSubvolCreate) context() string {
	return fmt.Sprintf("qgroups=%v, dest='%s', name='%s'", c.qgroups, c.dest, c.name)
}

func (c *cmdSubvolCreate) error(err error) *btrfs.BtrfsError {
	return &btrfs.BtrfsError{Func: string(btrfs.CmdSubvolumeCreate), Context: c.context(), Err: err}
}

func (c *cmdSubvolCreate) validate() error {
	c.name = strings.TrimSpace(c.name)
	if !validators.IsSubvolumeName(c.name) {
		return fmt.Errorf("incorrect subvolume name '%s'", c.name)
	}

	if len(c.name) > btrfs.BtrfsVolNameMax {
		return fmt.Errorf("subvolume name too long '%s', max length is %d", c.name, btrfs.BtrfsVolNameMax)
	}

	dest := filepath.Join(c.dest, c.name)
	fi, err := os.Stat(dest)
	if err == nil && fi.IsDir() {
		return fmt.Errorf("'%s' exists", dest)
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

	err = ioctl.SubvolCreate(c.dest, c.name)
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
