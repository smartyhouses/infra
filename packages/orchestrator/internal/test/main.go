package test

import (
	"context"
	"fmt"
	"time"

	"github.com/e2b-dev/infra/packages/orchestrator/internal/dns"
	sandboxStorage "github.com/e2b-dev/infra/packages/orchestrator/internal/sandbox/storage"
	templateStorage "github.com/e2b-dev/infra/packages/shared/pkg/storage"

	"cloud.google.com/go/storage"
)

func Run(envID, buildID, instanceID string, keepAlive, count *int) {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	// Start of mock build for testing
	dns := dns.New()
	go dns.Start("127.0.0.4:53")

	client, err := storage.NewClient(ctx, storage.WithJSONReads())
	if err != nil {
		errMsg := fmt.Errorf("failed to create GCS client: %v", err)
		panic(errMsg)
	}

	templateCache := sandboxStorage.NewTemplateDataCache(ctx, client, templateStorage.BucketName)

	MockInstance(ctx, envID, buildID, instanceID+"-1", dns, templateCache, time.Duration(*keepAlive)*time.Second)
	MockInstance(ctx, envID, buildID, instanceID+"-2", dns, templateCache, time.Duration(*keepAlive)*time.Second)
}
