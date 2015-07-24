package subvolume

// snapshot command
type Snapshot interface {
	ReadOnly() Snapshot
	Source(src string) Snapshot
	Destination(dest string) Snapshot
	Name(name string) Snapshot

	Execute() error
}

func NewSnapshot() Snapshot {
	return &snapshot{}
}

type snapshot struct {
	readOnly bool
	src      string
	dest     string
	name     string
}

func (s *snapshot) ReadOnly() Snapshot {
	s.readOnly = true

	return s
}

func (s *snapshot) Source(src string) Snapshot {
	s.src = src

	return s
}

func (s *snapshot) Destination(dest string) Snapshot {
	s.dest = dest

	return s
}

func (s *snapshot) Name(name string) Snapshot {
	s.name = name

	return s
}

func (s *snapshot) Execute() error {
	return nil
}
