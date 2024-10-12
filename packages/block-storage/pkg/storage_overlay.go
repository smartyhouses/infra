package block_storage

import (
	"fmt"
	"io"

	"github.com/e2b-dev/infra/packages/block-storage/pkg/cache"
	"github.com/e2b-dev/infra/packages/block-storage/pkg/overlay"
)

type BlockStorageOverlay struct {
	overlay   *overlay.Overlay
	cache     *cache.MmapCache
	size      int64
	blockSize int64
}

func newBlockStorageOverlay(base io.ReaderAt, cachePath string, size, blockSize int64) (*BlockStorageOverlay, error) {
	cache, err := cache.NewMmapCache(size, blockSize, cachePath)
	if err != nil {
		return nil, fmt.Errorf("error creating cache: %w", err)
	}

	overlay := overlay.New(base, cache, true)

	return &BlockStorageOverlay{
		blockSize: blockSize,
		overlay:   overlay,
		size:      size,
		cache:     cache,
	}, nil
}

func (o *BlockStorageOverlay) ReadAt(p []byte, off int64) (n int, err error) {
	return o.overlay.ReadAt(p, off)
}

func (o *BlockStorageOverlay) WriteAt(p []byte, off int64) (n int, err error) {
	return o.overlay.WriteAt(p, off)
}

func (o *BlockStorageOverlay) Size() (int64, error) {
	return o.size, nil
}

func (o *BlockStorageOverlay) BlockSize() int64 {
	return o.blockSize
}

func (o *BlockStorageOverlay) Sync() error {
	// We don't sync overlay cache because we are not using it for anything.
	// We might add this later, but right now it's not needed.
	return nil
}

func (o *BlockStorageOverlay) Close() error {
	return o.cache.Close()
}

func (o *BlockStorageOverlay) ReadRaw(offset, length int64) ([]byte, func(), error) {
	return o.overlay.ReadRaw(offset, length)
}
