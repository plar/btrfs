package btrfs

import (
	"fmt"

	"github.com/satori/go.uuid"
)

type ApiType int
type Command string
type CommandFactory func() interface{}

const (
	IOCTL ApiType = iota
	CLI
)

const (
	CmdSubvolCreate   Command = "subvolume create"
	CmdSubvolSnapshot Command = "subvolume snapshot"
	CmdSubvolFindNew  Command = "subvolume find-new"
	CmdSubvolDelete   Command = "subvolume delete"
	CmdSubvolList     Command = "subvolume list"
)

const (
	BtrfsVolNameMax = 255
)

func (at ApiType) String() string {
	switch at {
	case IOCTL:
		return "IOCTL"
	case CLI:
		return "CLI"
	default:
		return fmt.Sprintf("%d", int(at))
	}
}

// All Public API Methods should return BtrfsError
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

//
type API interface {
	Subvolume() Subvolume
}

type Subvolume interface {
	Create() SubvolCreate
	Snapshot() SubvolSnapshot
	Delete() SubvolDelete
	List() SubvolList
}

type SubvolCreate interface {
	Executor

	QuotaGroups(qgroups ...string) SubvolCreate
	Destination(dest string) SubvolCreate
}

type SubvolSnapshot interface {
	Executor

	QuotaGroups(qgroups ...string) SubvolSnapshot
	ReadOnly() SubvolSnapshot
	Source(src string) SubvolSnapshot
	Destination(dest string) SubvolSnapshot
}

type SubvolFindNew interface {
	Executor

	Destination(dest string) SubvolFindNew
	LastGen(uint64) SubvolFindNew
}

type SubvolDelete interface {
	Executor

	Destination(dest string) SubvolDelete
}

type SubvolInfo struct {
	Path             string
	ParentID         uint64
	ID               uint64
	OriginGeneration uint64
	Generation       uint64
	ParentUUID       uuid.UUID
	UUID             uuid.UUID
	IsSnapshot       bool
	IsReadOnly       bool

	Childred []SubvolInfo
}

type SubvolList interface {
	Path(path string) SubvolList

	FilterGeneration(filter string) SubvolList
	FilterOriginGeneration(filter string) SubvolList
	Sort(order string) SubvolList

	Execute() ([]SubvolInfo, error)
}

type api struct {
	apiType ApiType
}

func (a *api) Subvolume() Subvolume {
	return &subvolume{apiType: a.apiType}
}

type subvolume struct {
	apiType ApiType
}

func (s *subvolume) Create() SubvolCreate {
	cmd, ok := factory(s.apiType, CmdSubvolCreate).(SubvolCreate)
	if !ok {
		panic("Expected btrfs.SubvolCreate interface")
	}
	return cmd
}

func (s *subvolume) Snapshot() SubvolSnapshot {
	cmd, ok := factory(s.apiType, CmdSubvolSnapshot).(SubvolSnapshot)
	if !ok {
		panic("Expected btrfs.SubvolSnapshot interface")
	}
	return cmd
}

func (s *subvolume) Delete() SubvolDelete {
	cmd, ok := factory(s.apiType, CmdSubvolDelete).(SubvolDelete)
	if !ok {
		panic("Expected btrfs.SubvolDelete interface")
	}
	return cmd
}

func (s *subvolume) List() SubvolList {
	cmd, ok := factory(s.apiType, CmdSubvolList).(SubvolList)
	if !ok {
		panic("Expected btrfs.SubvolList interface")
	}
	return cmd
}

func NewIoctl() API {
	return &api{apiType: IOCTL}
}

func NewCli() API {
	return &api{apiType: CLI}
}

//

var commands map[ApiType]map[Command]CommandFactory

func init() {
	commands = make(map[ApiType]map[Command]CommandFactory)
}

func RegisterAPI(apiType ApiType, cmd Command, factory CommandFactory) {
	if _, exists := commands[apiType]; !exists {
		commands[apiType] = make(map[Command]CommandFactory)
	}
	commands[apiType][cmd] = factory
}

func factory(apiType ApiType, cmd Command) interface{} {
	if _, exists := commands[apiType]; !exists {
		panic(fmt.Sprintf("Unsupported API type: %s", apiType))
	}

	if command, exists := commands[apiType][cmd]; !exists {
		panic(fmt.Sprintf("%s: Unsupported command '%s'", apiType, cmd))
	} else {
		return command()
	}
}
