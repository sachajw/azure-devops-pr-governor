package main

import (
	"context"
	"log"
	"os"

	"github.com/sachajw/azure-devops-pr-scheduler/internal/azuredevops"
	"github.com/sachajw/azure-devops-pr-scheduler/internal/hooks"
	"github.com/sachajw/azure-devops-pr-scheduler/internal/routes"
	"github.com/sachajw/azure-devops-pr-scheduler/internal/services"
	_ "github.com/sachajw/azure-devops-pr-scheduler/migrations"
	"github.com/pocketbase/pocketbase"
)

func main() {
	app := pocketbase.New()

	// Register record lifecycle hooks
	hooks.RegisterPolicyHooks(app)
	hooks.RegisterRunHooks(app)

	// Register custom API routes
	routes.RegisterSimulateRoute(app)
	routes.RegisterExecuteRoute(app)
	routes.RegisterWebhookRoute(app)

	// Register cron schedule runner
	cronExpr := os.Getenv("SCHEDULE_RUNNER_CRON")
	if cronExpr == "" {
		cronExpr = "*/5 * * * *" // every 5 minutes
	}

	app.Cron().MustAdd("schedule-runner", cronExpr, func() {
		log.Println("schedule runner: starting evaluation")

		policyService := services.NewPolicyService(app)
		auditService := services.NewAuditService(app)

		orgURL := os.Getenv("AZURE_DEVOPS_ORG_URL")
		pat := os.Getenv("AZURE_DEVOPS_PAT")
		var adoClient *azuredevops.Client
		if orgURL != "" && pat != "" {
			adoClient = azuredevops.NewClient(orgURL, pat)
		}

		prService := services.NewPullRequestService(adoClient, auditService)
		scheduleService := services.NewScheduleService(app, policyService, prService, auditService)

		ctx := context.Background()
		summary, err := scheduleService.EvaluateAndExecute(ctx, false)
		if err != nil {
			log.Printf("schedule runner: error: %v", err)
			return
		}

		log.Printf("schedule runner: evaluated=%d matched=%d succeeded=%d failed=%d",
			summary.TotalEvaluated, summary.TotalMatched,
			summary.TotalSucceeded, summary.TotalFailed)
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
