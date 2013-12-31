package linux_backend

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/vito/garden/backend"
)

type Container interface {
	Snapshot(io.Writer) error
	Cleanup()

	backend.Container
}

type ContainerPool interface {
	Setup() error
	Create(backend.ContainerSpec) (Container, error)
	Restore(io.Reader) (Container, error)
	Destroy(Container) error
}

type LinuxBackend struct {
	containerPool ContainerPool
	snapshotsPath string

	containers map[string]Container

	sync.RWMutex
}

type UnknownHandleError struct {
	Handle string
}

func (e UnknownHandleError) Error() string {
	return "unknown handle: " + e.Handle
}

type FailedToSnapshotError struct {
	OriginalError error
}

func (e FailedToSnapshotError) Error() string {
	return fmt.Sprintf("failed to save snapshot: %s", e.OriginalError)
}

func New(containerPool ContainerPool, snapshotsPath string) *LinuxBackend {
	return &LinuxBackend{
		containerPool: containerPool,
		snapshotsPath: snapshotsPath,

		containers: make(map[string]Container),
	}
}

func (b *LinuxBackend) Setup() error {
	return b.containerPool.Setup()
}

func (b *LinuxBackend) Start() error {
	if b.snapshotsPath != "" {
		err := os.MkdirAll(b.snapshotsPath, 0755)
		if err != nil {
			return err
		}

		err = b.restoreSnapshots()
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *LinuxBackend) Create(spec backend.ContainerSpec) (backend.Container, error) {
	container, err := b.containerPool.Create(spec)
	if err != nil {
		return nil, err
	}

	err = container.Start()
	if err != nil {
		return nil, err
	}

	b.Lock()

	b.containers[container.Handle()] = container

	b.Unlock()

	return container, nil
}

func (b *LinuxBackend) Destroy(handle string) error {
	container, found := b.containers[handle]
	if !found {
		return UnknownHandleError{handle}
	}

	err := b.containerPool.Destroy(container)
	if err != nil {
		return err
	}

	b.Lock()

	delete(b.containers, container.Handle())

	b.Unlock()

	return nil
}

func (b *LinuxBackend) Containers() (containers []backend.Container, err error) {
	b.RLock()
	defer b.RUnlock()

	for _, container := range b.containers {
		containers = append(containers, container)
	}

	return containers, nil
}

func (b *LinuxBackend) Lookup(handle string) (backend.Container, error) {
	b.RLock()
	defer b.RUnlock()

	container, found := b.containers[handle]
	if !found {
		return nil, UnknownHandleError{handle}
	}

	return container, nil
}

func (b *LinuxBackend) Stop() {
	b.RLock()
	defer b.RUnlock()

	for _, container := range b.containers {
		container.Cleanup()
		b.saveSnapshot(container)
	}
}

func (b *LinuxBackend) restoreSnapshots() error {
	entries, err := ioutil.ReadDir(b.snapshotsPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		snapshot := path.Join(b.snapshotsPath, entry.Name())

		file, err := os.Open(snapshot)
		if err != nil {
			return err
		}

		_, err = b.restore(file)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *LinuxBackend) saveSnapshot(container Container) error {
	if b.snapshotsPath == "" {
		return nil
	}

	tmpfile, err := ioutil.TempFile(os.TempDir(), "snapshot-"+container.ID())
	if err != nil {
		return &FailedToSnapshotError{err}
	}

	err = container.Snapshot(tmpfile)
	if err != nil {
		return &FailedToSnapshotError{err}
	}

	snapshotPath := path.Join(b.snapshotsPath, container.ID())

	err = os.Rename(tmpfile.Name(), snapshotPath)
	if err != nil {
		return &FailedToSnapshotError{err}
	}

	return nil
}

func (b *LinuxBackend) restore(snapshot io.Reader) (backend.Container, error) {
	container, err := b.containerPool.Restore(snapshot)
	if err != nil {
		return nil, err
	}

	b.Lock()

	b.containers[container.Handle()] = container

	b.Unlock()

	return container, nil
}
