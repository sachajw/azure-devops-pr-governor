package routes

import (
	"os"

	"github.com/sachajw/azure-devops-pr-scheduler/internal/azuredevops"
)

// newADOClientFromEnv creates an Azure DevOps client from environment variables.
// Returns nil if the required env vars are not set (e.g. in simulation mode).
func newADOClientFromEnv() *azuredevops.Client {
	orgURL := os.Getenv("AZURE_DEVOPS_ORG_URL")
	pat := os.Getenv("AZURE_DEVOPS_PAT")

	if orgURL == "" || pat == "" {
		return nil
	}

	return azuredevops.NewClient(orgURL, pat)
}
