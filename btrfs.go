package btrfs

import "fmt"

type ApiType int
type Command string
type CommandFactory func() interface{}

const (
	IOCTL ApiType = iota
	CLI
)

const (
	CmdSubvolumeCreate   Command = "subvolume create"
	CmdSubvolumeSnapshot Command = "subvolume snapshot"
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
	Create() CmdSubvolCreate
	Snapshot() CmdSubvolSnapshot
}

type CmdSubvolCreate interface {
	Executor

	QuotaGroups(qgroups ...string) CmdSubvolCreate
	Destination(dest string) CmdSubvolCreate
	Name(name string) CmdSubvolCreate
}

type CmdSubvolSnapshot interface {
	Executor

	QuotaGroups(qgroups ...string) CmdSubvolSnapshot
	ReadOnly() CmdSubvolSnapshot
	Source(src string) CmdSubvolSnapshot
	Destination(dest string) CmdSubvolSnapshot
	Name(name string) CmdSubvolSnapshot
}

type Delete interface {
	Executor
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

func (s *subvolume) Create() CmdSubvolCreate {
	cmd, ok := factory(s.apiType, CmdSubvolumeCreate).(CmdSubvolCreate)
	if !ok {
		panic("Expected btrfs.CmdSubvolCreate interface")
	}
	return cmd
}

func (s *subvolume) Snapshot() CmdSubvolSnapshot {
	cmd, ok := factory(s.apiType, CmdSubvolumeSnapshot).(CmdSubvolSnapshot)
	if !ok {
		panic("Expected btrfs.CmdSubvolSnapshot interface")
	}
	return cmd
}

// func (s *subvolume) Snapshot() Snapshot {
// 	return api[s.apiType][SubvolumeSnapshot]().(Snapshot)
// }

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
