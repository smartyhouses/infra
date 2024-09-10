package storage

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	blockStorage "github.com/e2b-dev/infra/packages/block-storage/pkg"
	"github.com/e2b-dev/infra/packages/shared/pkg/template"

	"cloud.google.com/go/storage"
	"github.com/jellydator/ttlcache/v3"
)

const (
	templateDataExpiration = time.Hour * 25
	pageSize               = 2 << 11
	hugepageSize           = 2 << 20
)

type TemplateData struct {
	Memfile  *blockStorage.BlockStorage
	Snapfile *SimpleFile

	ensureOpen func() (*TemplateData, error)
}

func (t *TemplateData) Close() error {
	return t.Memfile.Close()
}

func newTemplateData(ctx context.Context, bucket *storage.BucketHandle, templateId, buildId string, hugePages bool) *TemplateData {
	h := &TemplateData{}

	h.ensureOpen = sync.OnceValues(func() (*TemplateData, error) {
		paths := template.NewTemplateFiles(templateId, buildId)

		err := os.MkdirAll(paths.CacheDir(), os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", paths.CacheDir(), err)
		}

		h.Snapfile = NewSimpleFile(ctx, bucket, paths.StorageSnapfilePath(), paths.CacheSnapfilePath())

		go h.Snapfile.Ensure()

		memfileObject := blockStorage.NewBucketObject(
			ctx,
			bucket,
			paths.StorageMemfilePath(),
		)

		var blockSize int64
		if hugePages {
			blockSize = hugepageSize
		} else {
			blockSize = pageSize
		}

		memfileStorage, err := blockStorage.New(
			ctx,
			memfileObject,
			paths.CacheMemfilePath(),
			blockSize,
		)

		h.Memfile = memfileStorage

		return h, err
	})

	return h
}

type TemplateDataCache struct {
	cache  *ttlcache.Cache[string, *TemplateData]
	bucket *storage.BucketHandle
	ctx    context.Context
}

func (t *TemplateDataCache) GetTemplateData(templateID, buildID string, hugePages bool) (*TemplateData, error) {
	id := fmt.Sprintf("%s-%s", templateID, buildID)

	templateData, _ := t.cache.GetOrSet(
		id,
		newTemplateData(t.ctx, t.bucket, templateID, buildID, hugePages),
		ttlcache.WithTTL[string, *TemplateData](templateDataExpiration),
	)

	mp, err := templateData.Value().ensureOpen()
	if err != nil {
		t.cache.Delete(id)

		return nil, fmt.Errorf("failed to create template data cache %s: %w", id, err)
	}

	return mp, nil
}

func NewTemplateDataCache(ctx context.Context, client *storage.Client, bucket string) *TemplateDataCache {
	b := client.Bucket(bucket)

	cache := ttlcache.New(
		ttlcache.WithTTL[string, *TemplateData](templateDataExpiration),
	)

	cache.OnEviction(func(ctx context.Context, reason ttlcache.EvictionReason, item *ttlcache.Item[string, *TemplateData]) {
		data := item.Value()

		err := data.Close()
		if err != nil {
			fmt.Printf("failed to cleanup template data for item %s: %v", item.Key(), err)
		}
	})

	go cache.Start()

	return &TemplateDataCache{
		bucket: b,
		cache:  cache,
		ctx:    ctx,
	}
}
