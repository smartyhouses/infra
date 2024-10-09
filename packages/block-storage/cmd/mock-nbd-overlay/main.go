package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/e2b-dev/infra/packages/block-storage/pkg/nbd"
	"golang.org/x/sync/errgroup"
)

type testDevice struct {
	f *os.File
}

func (t *testDevice) BlockSize() int64 {
	return 4096
}

func (t *testDevice) ReadRaw(off int64, size int64) ([]byte, func(), error) {
	b := make([]byte, size)

	n, err := t.f.ReadAt(b, off)

	return b[:n], func() {}, err
}

func (t *testDevice) Size() (int64, error) {
	fmt.Println("getting size")

	fi, err := t.f.Stat()
	if err != nil {
		return 0, err
	}

	return fi.Size(), nil
}

func (t *testDevice) Close() error {
	fmt.Println("closing")

	return t.f.Close()
}

func (t *testDevice) ReadAt(b []byte, off int64) (int, error) {
	fmt.Printf("read at %d, size %d\n", off, len(b))

	return t.f.ReadAt(b, off)
}

func (t *testDevice) WriteAt(b []byte, off int64) (int, error) {
	fmt.Printf("write at %d, size %d\n", off, len(b))

	return t.f.WriteAt(b, off)
}

func (t *testDevice) Sync() error {
	fmt.Println("syncing")

	return t.f.Sync()
}

func NewTestDevice(path string) (*testDevice, error) {
	f, err := os.OpenFile(path, os.O_RDWR, 0o666)
	if err != nil {
		return nil, err
	}

	return &testDevice{f: f}, nil
}

func main() {
	fmt.Println("creating nbd device pool")

	pool, err := nbd.NewNbdDevicePool()
	if err != nil {
		fmt.Println("error creating nbd device pool", err)
		return
	}

	fmt.Println("creating test device")

	// dd if=/dev/zero of=test.ext4 bs=4096K count=500
	// mkfs.ext4 test.ext4
	device, err := NewTestDevice(".test/test.ext4")
	if err != nil {
		fmt.Println("error creating test device", err)
		return
	}

	defer device.Close()

	fmt.Println("creating temp file")

	socketPath := "/tmp/nbd.sock"
	defer os.Remove(socketPath)

	fmt.Println("creating nbd")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	n, err := nbd.NewNbd(ctx, device, pool, socketPath)
	if err != nil {
		fmt.Println("error creating nbd", err)
		return
	}

	e, ctx := errgroup.WithContext(ctx)

	e.Go(func() error {
		fmt.Println("starting server")
		err := n.StartServer()
		if err != nil {
			return err
		}

		return nil
	})

	time.Sleep(1 * time.Second)

	e.Go(func() error {
		fmt.Println("starting client")
		err := n.StartClient()
		if err != nil {
			return err
		}

		return nil
	})

	defer n.Close()

	fmt.Printf("nbd path: %s\n", n.Path)

	if err := e.Wait(); err != nil {
		fmt.Println("error waiting for server and client", err)
	}
}
