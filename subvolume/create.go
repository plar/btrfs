package subvolume

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"os"

	"github.com/plar/btrfs"
	"github.com/plar/btrfs/ioctl"
)

// create command
type Create interface {
	btrfs.Executor

	QuotaGroups(qgroups ...string) Create
	Destination(dest string) Create
	Name(name string) Create
}

type createContext struct {
	qgroups []string
	dest    string
	name    string

	executor func(c *createContext) error
}

func (c *createContext) QuotaGroups(qgroups ...string) Create {
	for _, qgroup := range qgroups {
		c.qgroups = append(c.qgroups, qgroup)
	}
	return c
}

func (c *createContext) Destination(dest string) Create {
	c.dest = dest
	return c
}

func (c *createContext) Name(name string) Create {
	c.name = name
	return c
}

func (c *createContext) context() string {
	return fmt.Sprintf("qgroups=%v, dest='%s', name='%s'", c.qgroups, c.dest, c.name)
}

func (c *createContext) error(err error) *btrfs.BtrfsError {
	return &btrfs.BtrfsError{Func: "Subvolume Create", Context: c.context(), Err: err}
}

func (c *createContext) validate() error {
	c.name = strings.TrimSpace(c.name)
	if !btrfs.TestSubvolumeName(c.name) {
		return c.error(fmt.Errorf("incorrect subvolume name '%s'", c.name))
	}

	if len(c.name) > btrfs.BtrfsVolNameMax {
		return c.error(fmt.Errorf("subvolume name too long '%s', max length is %d", c.name, btrfs.BtrfsVolNameMax))
	}

	fi, err := os.Stat(filepath.Join(c.dest, c.name))
	if err == nil && fi.IsDir() {
		return c.error(fmt.Errorf("'%s' exists", c.dest))
	}

	return nil
}

func (c *createContext) Execute() error {
	return c.executor(c)
}

// btrfs ioctl executor
func ioctlExecute(c *createContext) error {
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

func IoctlNewCreate() Create {
	return &createContext{executor: ioctlExecute}
}

func cliExecute(c *createContext) error {
	err := c.validate()
	if err != nil {
		return err
	}

	return c.error(errors.New("Unimplemented"))
}

func CliNewCreate() Create {
	return &createContext{executor: cliExecute}
}
