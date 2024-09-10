package template

import (
	"context"
	"fmt"
	"io"

	blockStorage "github.com/e2b-dev/infra/packages/block-storage/pkg/source"
	"github.com/e2b-dev/infra/packages/shared/pkg/template"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

type TemplateStorage struct {
	bucket *storage.BucketHandle
}

func NewTemplateStorage(ctx context.Context, client *storage.Client, bucket string) *TemplateStorage {
	b := client.Bucket(bucket)

	return &TemplateStorage{
		bucket: b,
	}
}

func (t *TemplateStorage) Remove(ctx context.Context, templateID string) error {
	objects := t.bucket.Objects(ctx, &storage.Query{
		Prefix: templateID + "/",
	})

	for {
		object, err := objects.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("error when iterating over template objects: %w", err)
		}

		err = t.bucket.Object(object.Name).Delete(ctx)
		if err != nil {
			return fmt.Errorf("error when deleting template object: %w", err)
		}
	}

	return nil
}

func (t *TemplateStorage) NewTemplateBuild(ctx context.Context, templateID, buildID string) (*TemplateBuild, error) {
	return &TemplateBuild{
		bucket: t.bucket,
		paths:  template.NewTemplateFiles(templateID, buildID),
	}, nil
}

type TemplateBuild struct {
	bucket *storage.BucketHandle
	paths  *template.TemplateFiles
}

func (t *TemplateBuild) Remove(ctx context.Context) error {
	objects := t.bucket.Objects(ctx, &storage.Query{
		Prefix: t.paths.BuildDir() + "/",
	})

	for {
		object, err := objects.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("error when iterating over template build objects: %w", err)
		}

		err = t.bucket.Object(object.Name).Delete(ctx)
		if err != nil {
			return fmt.Errorf("error when deleting template build object: %w", err)
		}
	}

	return nil
}

func (t *TemplateBuild) UploadMemfile(ctx context.Context, memfile io.Reader) error {
	object := blockStorage.NewGCSObjectFromBucket(ctx, t.bucket, t.paths.MemfilePath())

	_, err := object.ReadFrom(memfile)
	if err != nil {
		return fmt.Errorf("error when uploading memfile: %w", err)
	}

	return nil
}

func (t *TemplateBuild) UploadRootfs(ctx context.Context, rootfs io.Reader) error {
	object := blockStorage.NewGCSObjectFromBucket(ctx, t.bucket, t.paths.RootfsPath())

	_, err := object.ReadFrom(rootfs)
	if err != nil {
		return fmt.Errorf("error when uploading rootfs: %w", err)
	}

	return nil
}

func (t *TemplateBuild) UploadSnapfile(ctx context.Context, snapfile io.Reader) error {
	object := blockStorage.NewGCSObjectFromBucket(ctx, t.bucket, t.paths.SnapfilePath())

	_, err := object.ReadFrom(snapfile)
	if err != nil {
		return fmt.Errorf("error when uploading snapfile: %w", err)
	}

	return nil
}
