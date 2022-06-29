package landlock_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/landlock-lsm/go-landlock/landlock"
	ll "github.com/landlock-lsm/go-landlock/landlock/syscall"
	"golang.org/x/sys/unix"
)

func RequireLandlockABI(t *testing.T, want int) {
	t.Helper()

	if v, err := ll.LandlockGetABIVersion(); err != nil || v < want {
		t.Skipf("Requires Landlock >= V%v, got V%v (err=%v)", want, v, err)
	}
}

func TestPathDoesNotExist(t *testing.T) {
	RequireLandlockABI(t, 1)

	doesNotExistPath := filepath.Join(t.TempDir(), "does_not_exist")

	err := landlock.V1.RestrictPaths(
		landlock.RODirs(doesNotExistPath),
	)
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected 'not exist' error, got: %v", err)
	}
}

func TestRestrictingPlainFileWithDirectoryFlags(t *testing.T) {
	RequireLandlockABI(t, 1)

	err := landlock.V1.RestrictPaths(
		landlock.RODirs("/etc/passwd"),
	)
	if !errors.Is(err, unix.EINVAL) {
		t.Errorf("expected 'invalid argument' error, got: %v", err)
	}
}

func TestEmptyAccessRights(t *testing.T) {
	RequireLandlockABI(t, 1)

	err := landlock.V1.RestrictPaths(
		landlock.PathAccess(0, "/etc/passwd"),
	)
	if !errors.Is(err, unix.ENOMSG) {
		t.Errorf("expected ENOMSG, got: %v", err)
	}
	want := "empty access rights"
	if !strings.Contains(err.Error(), want) {
		t.Errorf("expected error message with %q, got: %v", want, err)
	}
}

func TestOverlyBroadPathOpt(t *testing.T) {
	RequireLandlockABI(t, 1)

	handled := landlock.AccessFSSet(0b011)
	excempt := landlock.AccessFSSet(0b111) // superset of handled!
	err := landlock.MustConfig(handled).RestrictPaths(
		landlock.PathAccess(excempt, "/tmp"),
	)
	if !errors.Is(err, unix.EINVAL) {
		t.Errorf("expected 'invalid argument' error, got: %v", err)
	}
}
