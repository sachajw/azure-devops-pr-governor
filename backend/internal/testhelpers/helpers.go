package testhelpers

import (
	"testing"

	"github.com/sachajw/azure-devops-pr-scheduler/internal/hooks"
	"github.com/sachajw/azure-devops-pr-scheduler/internal/routes"
	_ "github.com/sachajw/azure-devops-pr-scheduler/migrations"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

// NewTestApp creates a PocketBase TestApp with our migrations, hooks, and routes registered.
// The caller must defer app.Cleanup().
func NewTestApp(t testing.TB) *tests.TestApp {
	t.Helper()

	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}

	RegisterHooks(app)
	RegisterRoutes(app)

	return app
}

// RegisterHooks registers all application hooks on the test app.
func RegisterHooks(app core.App) {
	hooks.RegisterPolicyHooks(app)
	hooks.RegisterRunHooks(app)
}

// RegisterRoutes registers all custom routes on the test app.
func RegisterRoutes(app core.App) {
	routes.RegisterSimulateRoute(app)
	routes.RegisterExecuteRoute(app)
	routes.RegisterWebhookRoute(app)
}

// SuperuserToken creates a superuser and returns a valid auth token.
func SuperuserToken(t testing.TB, app *tests.TestApp, email, password string) string {
	t.Helper()

	superusers, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		t.Fatal(err)
	}

	record := core.NewRecord(superusers)
	record.Set("email", email)
	record.Set("password", password)
	record.Set("passwordConfirm", password)
	if err := app.Save(record); err != nil {
		t.Fatal(err)
	}

	token, err := record.NewAuthToken()
	if err != nil {
		t.Fatal(err)
	}

	return token
}

// CreatePolicyRecord creates a policy record and returns its ID.
func CreatePolicyRecord(t testing.TB, app *tests.TestApp, overrides ...func(*core.Record)) string {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("policies")
	if err != nil {
		t.Fatal(err)
	}

	record := core.NewRecord(collection)
	record.Set("name", "test-policy")
	record.Set("version", "1.0.0")
	record.Set("enabled", true)
	record.Set("scope_org", "testorg")
	record.Set("scope_project", "testproject")
	record.Set("scope_repo", "testrepo")
	record.Set("schedule_cron", "0 1 * * *")
	record.Set("schedule_timezone", "UTC")
	record.Set("schedule_enabled", true)
	record.Set("conditions", []any{})
	record.Set("actions", []any{
		map[string]any{
			"type": "create_pr",
			"parameters": map[string]any{
				"sourceRefName": "refs/heads/dev",
				"targetRefName": "refs/heads/qa",
				"title":         "Auto PR: dev -> qa",
			},
		},
	})

	for _, fn := range overrides {
		fn(record)
	}

	if err := app.Save(record); err != nil {
		t.Fatal(err)
	}

	return record.Id
}
