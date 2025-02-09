package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/e2b-dev/infra/packages/api/internal/api"
	"github.com/e2b-dev/infra/packages/api/internal/auth"
	authcache "github.com/e2b-dev/infra/packages/api/internal/cache/auth"
	"github.com/e2b-dev/infra/packages/shared/pkg/models"
	"github.com/e2b-dev/infra/packages/shared/pkg/models/envbuild"
	"github.com/e2b-dev/infra/packages/shared/pkg/telemetry"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

func (a *APIStore) getSandboxes(ctx context.Context, teamID uuid.UUID, query *string) ([]api.RunningSandbox, error) {

	instanceInfo := a.orchestrator.GetSandboxes(ctx, &teamID)

	if query != nil {
		// Unescape query
		query, err := url.QueryUnescape(*query)
		if err != nil {
			return nil, fmt.Errorf("error when unescaping query: %w", err)
		}

		// Parse filters, both key and value are also unescaped
		filters := make(map[string]string)

		for _, filter := range strings.Split(query, "&") {
			parts := strings.Split(filter, "=")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid key value pair in query")
			}

			key, err := url.QueryUnescape(parts[0])
			if err != nil {
				return nil, fmt.Errorf("error when unescaping key: %w", err)
			}

			value, err := url.QueryUnescape(parts[1])
			if err != nil {
				return nil, fmt.Errorf("error when unescaping value: %w", err)
			}

			filters[key] = value
		}

		// Filter instances to match all filters
		n := 0
		for _, instance := range instanceInfo {
			if instance.Metadata == nil {
				continue
			}

			matchesAll := true
			for key, value := range filters {
				if metadataValue, ok := instance.Metadata[key]; !ok || metadataValue != value {
					matchesAll = false
					break
				}
			}

			if matchesAll {
				instanceInfo[n] = instance
				n++
			}
		}

		// Trim slice
		instanceInfo = instanceInfo[:n]
	}

	buildIDs := make([]uuid.UUID, 0)
	for _, info := range instanceInfo {
		if info.TeamID == nil {
			continue
		}

		if *info.TeamID != teamID {
			continue
		}

		buildIDs = append(buildIDs, *info.BuildID)
	}

	builds, err := a.db.Client.EnvBuild.Query().Where(envbuild.IDIn(buildIDs...)).All(ctx)
	if err != nil {
		telemetry.ReportCriticalError(ctx, err)

		return nil, fmt.Errorf("error when getting builds: %w", err)
	}

	buildsMap := make(map[uuid.UUID]*models.EnvBuild, len(builds))
	for _, build := range builds {
		buildsMap[build.ID] = build
	}

	sandboxes := make([]api.RunningSandbox, 0)

	for _, info := range instanceInfo {
		if info.TeamID == nil {
			continue
		}

		if *info.TeamID != teamID {
			continue
		}

		if info.BuildID == nil {
			continue
		}

		instance := api.RunningSandbox{
			ClientID:   info.Instance.ClientID,
			TemplateID: info.Instance.TemplateID,
			Alias:      info.Instance.Alias,
			SandboxID:  info.Instance.SandboxID,
			StartedAt:  info.StartTime,
			CpuCount:   int32(buildsMap[*info.BuildID].Vcpu),
			MemoryMB:   int32(buildsMap[*info.BuildID].RAMMB),
			EndAt:      info.EndTime,
		}

		if info.Metadata != nil {
			meta := api.SandboxMetadata(info.Metadata)
			instance.Metadata = &meta
		}

		sandboxes = append(sandboxes, instance)
	}

	// Sort sandboxes by start time descending
	slices.SortFunc(sandboxes, func(a, b api.RunningSandbox) int {
		return a.StartedAt.Compare(b.StartedAt)
	})

	return sandboxes, nil
}

func (a *APIStore) GetSandboxes(c *gin.Context, params api.GetSandboxesParams) {
	ctx := c.Request.Context()
	telemetry.ReportEvent(ctx, "list running instances")

	teamInfo := c.Value(auth.TeamContextKey).(authcache.AuthTeamInfo)
	team := teamInfo.Team

	a.posthog.IdentifyAnalyticsTeam(team.ID.String(), team.Name)
	properties := a.posthog.GetPackageToPosthogProperties(&c.Request.Header)
	a.posthog.CreateAnalyticsTeamEvent(team.ID.String(), "listed running instances", properties)

	sandboxes, err := a.getSandboxes(ctx, team.ID, params.Query)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, sandboxes)
}

const (
	getSandboxesMetricsTimeout = 30 * time.Second
	maxConcurrentMetricFetches = 30
)

func (a *APIStore) getSandboxesMetrics(
	ctx context.Context,
	teamID uuid.UUID,
	sandboxes []api.RunningSandbox,
) ([]api.RunningSandboxWithMetrics, error) {
	// Add operation telemetry
	telemetry.ReportEvent(ctx, "fetch metrics for sandboxes")
	telemetry.SetAttributes(ctx,
		attribute.String("team.id", teamID.String()),
		attribute.Int("sandboxes.count", len(sandboxes)),
	)

	startTime := time.Now()
	defer func() {
		// Report operation duration
		duration := time.Since(startTime)
		telemetry.SetAttributes(ctx,
			attribute.Float64("operation.duration_ms", float64(duration.Milliseconds())),
		)
	}()

	type metricsResult struct {
		sandbox api.RunningSandbox
		metrics []api.SandboxMetric
		err     error
	}

	// TODO: Potential scaling consideration:
	//
	// Current implementation creates a buffered channel sized to the number of sandboxes.
	// This could consume significant memory with very large numbers of sandboxes (thousands+).
	// Consider implementing a fixed-size channel with backpressure for better resource management:
	//
	// Example approach:
	// const maxChannelBuffer = 100
	// results := make(chan metricsResult, maxChannelBuffer)
	//
	// Then in the goroutine:
	// select {
	// case results <- result:
	// case <-ctx.Done():
	//     return
	// }
	//
	// And process results concurrently:
	// go func() {
	//     for result := range results {
	//         processAndAppendResult(result)
	//     }
	// }()

	results := make(chan metricsResult, len(sandboxes))
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrentMetricFetches)

	// Error tracking with atomic counters
	var errorCount atomic.Int32
	var timeoutCount atomic.Int32
	var successCount atomic.Int32

	// Fetch metrics for each sandbox concurrently with rate limiting
	for _, sandbox := range sandboxes {
		wg.Add(1)
		go func(s api.RunningSandbox) {
			defer wg.Done()

			// Try to acquire semaphore with context
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }() // Release semaphore when done
			case <-ctx.Done():
				timeoutCount.Add(1)
				err := fmt.Errorf("context cancelled while waiting for rate limiter: %w", ctx.Err())
				telemetry.ReportError(ctx, err)
				results <- metricsResult{
					sandbox: s,
					err:     err,
				}
				return
			}

			// Get metrics for this sandbox
			metrics, err := a.getSandboxesSandboxIDMetrics(
				ctx,
				s.SandboxID,
				teamID.String(),
				1,
				oldestLogsLimit,
			)

			if err != nil {
				errorCount.Add(1)
				telemetry.ReportError(ctx, fmt.Errorf("failed to fetch metrics for sandbox %s: %w", s.SandboxID, err))
			} else {
				successCount.Add(1)
			}

			results <- metricsResult{
				sandbox: s,
				metrics: metrics,
				err:     err,
			}
		}(sandbox)
	}

	// Close results channel once all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results and build final response
	sandboxesWithMetrics := make([]api.RunningSandboxWithMetrics, 0, len(sandboxes))

	// Process each result as it arrives
	for result := range results {
		sandbox := api.RunningSandboxWithMetrics{
			ClientID:   result.sandbox.ClientID,
			TemplateID: result.sandbox.TemplateID,
			Alias:      result.sandbox.Alias,
			SandboxID:  result.sandbox.SandboxID,
			StartedAt:  result.sandbox.StartedAt,
			CpuCount:   result.sandbox.CpuCount,
			MemoryMB:   result.sandbox.MemoryMB,
			EndAt:      result.sandbox.EndAt,
			Metadata:   result.sandbox.Metadata,
		}

		if result.err == nil {
			sandbox.Metrics = &result.metrics
		}

		sandboxesWithMetrics = append(sandboxesWithMetrics, sandbox)
	}

	// Report final metrics
	telemetry.SetAttributes(ctx,
		attribute.Int("metrics.success_count", int(successCount.Load())),
		attribute.Int("metrics.error_count", int(errorCount.Load())),
		attribute.Int("metrics.timeout_count", int(timeoutCount.Load())),
	)

	// Log operation summary
	a.logger.Info("Completed fetching sandbox metrics",
		zap.String("team_id", teamID.String()),
		zap.Int("total_sandboxes", len(sandboxes)),
		zap.Int32("successful_fetches", successCount.Load()),
		zap.Int32("failed_fetches", errorCount.Load()),
		zap.Int32("timeouts", timeoutCount.Load()),
		zap.Duration("duration", time.Since(startTime)),
	)

	return sandboxesWithMetrics, nil
}

func (a *APIStore) GetSandboxesMetrics(c *gin.Context, params api.GetSandboxesMetricsParams) {
	ctx := c.Request.Context()
	telemetry.ReportEvent(ctx, "list running instances with metrics")

	// Cancel context after timeout to ensure no goroutines are left hanging for too long
	ctx, cancel := context.WithTimeout(ctx, getSandboxesMetricsTimeout)
	defer cancel()

	team := c.Value(auth.TeamContextKey).(authcache.AuthTeamInfo).Team

	a.posthog.IdentifyAnalyticsTeam(team.ID.String(), team.Name)
	properties := a.posthog.GetPackageToPosthogProperties(&c.Request.Header)
	a.posthog.CreateAnalyticsTeamEvent(team.ID.String(), "listed running instances with metrics", properties)

	sandboxes, err := a.getSandboxes(ctx, team.ID, params.Query)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	sandboxesWithMetrics, err := a.getSandboxesMetrics(ctx, team.ID, sandboxes)

	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, sandboxesWithMetrics)
}
