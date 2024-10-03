package source

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockReaderAt struct {
	readAtFunc func(p []byte, off int64) (n int, err error)
}

func (m *mockReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	return m.readAtFunc(p, off)
}

func TestPrefetcher(t *testing.T) {
	t.Run("prefetch success", func(t *testing.T) {
		var readAtCalls int
		mockReader := &mockReaderAt{
			readAtFunc: func(p []byte, off int64) (n int, err error) {
				readAtCalls++
				return 0, nil
			},
		}

		prefetcher := NewPrefetcher(context.Background(), mockReader, 100*ChunkSize)
		err := prefetcher.Start()
		assert.NoError(t, err, "unexpected error")
		assert.Equal(t, 100, readAtCalls, "expected 100 ReadAt calls")
	})

	t.Run("context cancel", func(t *testing.T) {
		mockReader := &mockReaderAt{
			readAtFunc: func(p []byte, off int64) (n int, err error) {
				return 0, nil
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		prefetcher := NewPrefetcher(ctx, mockReader, 100*ChunkSize)

		cancel()

		err := prefetcher.Start()
		assert.ErrorIs(t, err, context.Canceled, "expected context.Canceled error")
	})
}