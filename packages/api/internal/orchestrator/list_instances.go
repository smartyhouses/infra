package orchestrator

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"

	"github.com/e2b-dev/infra/packages/api/internal/api"
	"github.com/e2b-dev/infra/packages/api/internal/cache/instance"
	"github.com/e2b-dev/infra/packages/api/internal/utils"
)

func (o *Orchestrator) getInstances(ctx context.Context, nodeID string) ([]*instance.InstanceInfo, error) {
	childCtx, childSpan := o.tracer.Start(ctx, "list-instances")
	defer childSpan.End()

	client, err := o.GetClient(nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get GRPC client: %w", err)
	}

	res, err := client.Sandbox.List(childCtx, &empty.Empty{})

	err = utils.UnwrapGRPCError(err)
	if err != nil {
		return nil, fmt.Errorf("failed to list sandboxes: %w", err)
	}

	sandboxes := res.GetSandboxes()

	sandboxesInfo := make([]*instance.InstanceInfo, 0, len(sandboxes))

	for _, sbx := range sandboxes {
		config := sbx.GetConfig()

		if config == nil {
			return nil, fmt.Errorf("sandbox config is nil when listing sandboxes: %#v", sbx)
		}

		teamID, parseErr := uuid.Parse(config.TeamID)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse team ID '%s' for job: %w", config.TeamID, parseErr)
		}

		buildID, parseErr := uuid.Parse(config.BuildID)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse build ID '%s' for job: %w", config.BuildID, err)
		}

		sandboxesInfo = append(sandboxesInfo, &instance.InstanceInfo{
			Instance: api.Sandbox{
				SandboxID:  config.SandboxID,
				TemplateID: config.TemplateID,
				Alias:      config.Alias,
				ClientID:   sbx.ClientID,
			},
			StartTime:         sbx.StartTime.AsTime(),
			EndTime:           sbx.EndTime.AsTime(),
			VCpu:              config.VCpu,
			RamMB:             config.RamMB,
			BuildID:           buildID,
			TeamID:            teamID,
			Metadata:          config.Metadata,
			MaxInstanceLength: time.Duration(config.MaxInstanceLength) * time.Hour,
		})
	}

	return sandboxesInfo, nil
}
