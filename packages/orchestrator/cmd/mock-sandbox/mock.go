package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"time"

	nbd "github.com/e2b-dev/infra/packages/block-storage/pkg/nbd"
	"github.com/e2b-dev/infra/packages/orchestrator/internal/consul"
	"github.com/e2b-dev/infra/packages/orchestrator/internal/dns"
	"github.com/e2b-dev/infra/packages/orchestrator/internal/sandbox"
	sandboxStorage "github.com/e2b-dev/infra/packages/orchestrator/internal/sandbox/storage"
	snapshotStorage "github.com/e2b-dev/infra/packages/orchestrator/internal/sandbox/storage"
	"github.com/e2b-dev/infra/packages/shared/pkg/grpc/orchestrator"
	templateStorage "github.com/e2b-dev/infra/packages/shared/pkg/storage"
	"github.com/e2b-dev/infra/packages/shared/pkg/telemetry"

	"cloud.google.com/go/storage"
	"go.opentelemetry.io/otel"
)

func main() {
	templateId := flag.String("template", "", "template id")
	buildId := flag.String("build", "", "build id")
	sandboxId := flag.String("sandbox", "", "sandbox id")
	keepAlive := flag.Int("alive", 0, "keep alive")
	count := flag.Int("count", 1, "number of serially spawned sandboxes")

	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(*keepAlive)+time.Second*20)
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

	for i := 0; i < *count; i++ {
		mockSandbox(ctx, *templateId, *buildId, *sandboxId+"-"+strconv.Itoa(i), dns, templateCache, time.Duration(*keepAlive)*time.Second)
	}
}

func mockSandbox(
	ctx context.Context,
	templateId,
	buildId,
	sandboxId string,
	dns *dns.DNS,
	templateCache *snapshotStorage.TemplateDataCache,
	keepAlive time.Duration,
) {
	tracer := otel.Tracer(fmt.Sprintf("sandbox-%s", sandboxId))
	childCtx, _ := tracer.Start(ctx, "mock-sandbox")

	nbdDevicePool, err := nbd.NewNbdDevicePool()
	if err != nil {
		panic(err)
	}

	consulClient, err := consul.New(childCtx)

	networkPool := make(chan sandbox.IPSlot, 1)

	select {
	case <-ctx.Done():
		return
	default:
		ips, err := sandbox.NewSlot(ctx, tracer, consulClient)
		if err != nil {
			fmt.Printf("failed to create network: %v\n", err)

			return
		}

		err = ips.CreateNetwork(ctx, tracer)
		if err != nil {
			ips.Release(ctx, tracer, consulClient)

			fmt.Printf("failed to create network: %v\n", err)

			return
		}

		networkPool <- *ips
	}

	start := time.Now()

	sbx, err := sandbox.NewSandbox(
		childCtx,
		tracer,
		consulClient,
		dns,
		networkPool,
		templateCache,
		nbdDevicePool,
		&orchestrator.SandboxConfig{
			TemplateId:         templateId,
			FirecrackerVersion: "v1.7.0-dev_8bb88311",
			KernelVersion:      "vmlinux-5.10.186",
			TeamId:             "test-team",
			BuildId:            buildId,
			HugePages:          true,
			MaxSandboxLength:   1,
			SandboxId:          sandboxId,
		},
		"trace-test-1",
		time.Now(),
		time.Now(),
	)
	if err != nil {
		errMsg := fmt.Errorf("failed to create sandbox: %v", err)
		telemetry.ReportError(ctx, errMsg)
		return
	}

	duration := time.Since(start)

	fmt.Printf("[Sandbox is running] - started in %dms (without network)\n", duration.Milliseconds())

	time.Sleep(keepAlive)

	defer sbx.CleanupAfterFCStop(childCtx, tracer, consulClient, dns, sandboxId)

	sbx.Stop(childCtx, tracer)
}
