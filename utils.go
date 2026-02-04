package judgeval

import (
	"fmt"
	"sync"

	"github.com/JudgmentLabs/judgeval-go/internal/api"
	"github.com/JudgmentLabs/judgeval-go/internal/api/models"
	"github.com/JudgmentLabs/judgeval-go/logger"
)

var projectIDCache sync.Map

func resolveProjectID(client *api.Client, projectName string) (string, error) {
	cacheKey := fmt.Sprintf("org:%s:project:%s", client.GetOrganizationID(), projectName)

	if cached, ok := projectIDCache.Load(cacheKey); ok {
		return cached.(string), nil
	}

	logger.Info("Resolving project ID for project: %s", projectName)

	resp, err := client.PostProjectsResolve(&models.ResolveProjectRequest{
		ProjectName: projectName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to resolve project ID: %w", err)
	}

	if resp.ProjectId == "" {
		return "", fmt.Errorf("project ID not found for project: %s", projectName)
	}

	logger.Info("Resolved project ID: %s", resp.ProjectId)
	projectIDCache.Store(cacheKey, resp.ProjectId)
	return resp.ProjectId, nil
}
