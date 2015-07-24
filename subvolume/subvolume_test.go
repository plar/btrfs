package subvolume

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var rootDir, mount string

func TestSubVolumeCreate(t *testing.T) {
	cmd := IoctlNewCreate()
	assert.NotNil(t, cmd)

	err := cmd.QuotaGroups("1", "2", "3").Destination(mount).Name("volume1").Execute()
	assert.NoError(t, err)

	ctx := cmd.(*createContext)
	assert.Equal(t, ctx.qgroups, []string{"1", "2", "3"})
	assert.Equal(t, ctx.dest, mount)
	assert.Equal(t, ctx.name, "volume1")
}

func testSubVolumeSnapshot(t *testing.T) {
	cmd := NewSnapshot()
	assert.NotNil(t, cmd)

	err := cmd.ReadOnly().Source("source").Destination("/tmp/").Name("volume1").Execute()
	assert.NoError(t, err)

	ctx := cmd.(*snapshot)
	assert.Equal(t, ctx.readOnly, true)
	assert.Equal(t, ctx.src, "source")
	assert.Equal(t, ctx.dest, "/tmp/")
	assert.Equal(t, ctx.name, "volume1")
}

func run(cmd string, args ...string) error {
	log.Printf("Run %s %s", cmd, args)
	_, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func setup() {
	var err error

	rootDir, err = ioutil.TempDir("/var/tmp", "btrfs-test-")
	if err != nil {
		log.Fatalf("Cannot create tmp directory, err=%s", err)
	}

	mount = path.Join(rootDir, "btrfs")
	if err := os.MkdirAll(mount, 0700); err != nil {
		log.Fatalf("ERROR: MkdirAll %s, err=%s", mount, err)
	}

	imageFileName := filepath.Join(rootDir, "btrfs.img")
	ioutil.WriteFile(imageFileName, []byte(""), 0700)
	os.Truncate(imageFileName, 1024*1024*1024) // 1GB

	if err := run("mkfs.btrfs", imageFileName); err != nil {
		log.Fatalf("ERROR: mkfs.btrfs %s, err=%s", imageFileName, err)
	}

	if err := run("mount", imageFileName, mount); err != nil {
		log.Fatalf("ERROR: mount %s %s, err=%s", imageFileName, mount, err)
	}
}

func teardown() {
	if err := run("umount", mount); err != nil {
		log.Fatal("ERROR: umount, err=%s", err)
	}
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}
