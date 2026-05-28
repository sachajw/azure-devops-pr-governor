package routes_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/sachajw/azure-devops-pr-scheduler/internal/testhelpers"
	"github.com/pocketbase/pocketbase/tests"
)

func TestSimulateRoute(t *testing.T) {
	t.Run("should reject unauthenticated requests", func(t *testing.T) {
		app := testhelpers.NewTestApp(t)
		defer app.Cleanup()

		policyID := testhelpers.CreatePolicyRecord(t, app)

		scenario := tests.ApiScenario{
			Name:           "no auth",
			Method:         http.MethodPost,
			URL:            "/api/pr-scheduler/simulate",
			Body:           strings.NewReader(`{"policy_id":"` + policyID + `"}`),
			Headers:        map[string]string{"Content-Type": "application/json"},
			ExpectedStatus: 401,
			ExpectedContent: []string{
				`"status":401`,
			},
			TestAppFactory: func(t testing.TB) *tests.TestApp { return app },
		}
		scenario.Test(t)
	})

	t.Run("should simulate an existing policy", func(t *testing.T) {
		app := testhelpers.NewTestApp(t)
		defer app.Cleanup()

		policyID := testhelpers.CreatePolicyRecord(t, app)
		token := testhelpers.SuperuserToken(t, app, "admin@test.com", "testpass1234")

		scenario := tests.ApiScenario{
			Name:   "simulate existing policy",
			Method: http.MethodPost,
			URL:    "/api/pr-scheduler/simulate",
			Body:   strings.NewReader(`{"policy_id":"` + policyID + `"}`),
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": token,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"policy_id"`,
				`"conditions"`,
				`"actions"`,
			},
			TestAppFactory: func(t testing.TB) *tests.TestApp { return app },
		}
		scenario.Test(t)
	})

	t.Run("should reject missing policy_id", func(t *testing.T) {
		app := testhelpers.NewTestApp(t)
		defer app.Cleanup()

		token := testhelpers.SuperuserToken(t, app, "admin@test.com", "testpass1234")

		scenario := tests.ApiScenario{
			Name:   "missing policy_id",
			Method: http.MethodPost,
			URL:    "/api/pr-scheduler/simulate",
			Body:   strings.NewReader(`{}`),
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": token,
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`policy_id`,
			},
			TestAppFactory: func(t testing.TB) *tests.TestApp { return app },
		}
		scenario.Test(t)
	})

	t.Run("should return 404 for nonexistent policy", func(t *testing.T) {
		app := testhelpers.NewTestApp(t)
		defer app.Cleanup()

		token := testhelpers.SuperuserToken(t, app, "admin@test.com", "testpass1234")

		scenario := tests.ApiScenario{
			Name:   "nonexistent policy",
			Method: http.MethodPost,
			URL:    "/api/pr-scheduler/simulate",
			Body:   strings.NewReader(`{"policy_id":"nonexistent123"}`),
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": token,
			},
			ExpectedStatus: 404,
			ExpectedContent: []string{
				`policy not found`,
			},
			TestAppFactory: func(t testing.TB) *tests.TestApp { return app },
		}
		scenario.Test(t)
	})
}

func TestExecuteRoute(t *testing.T) {
	t.Run("should reject unauthenticated requests", func(t *testing.T) {
		app := testhelpers.NewTestApp(t)
		defer app.Cleanup()

		scenario := tests.ApiScenario{
			Name:           "no auth",
			Method:         http.MethodPost,
			URL:            "/api/pr-scheduler/execute",
			Body:           strings.NewReader(`{"policy_id":"test"}`),
			Headers:        map[string]string{"Content-Type": "application/json"},
			ExpectedStatus: 401,
			ExpectedContent: []string{
				`"status":401`,
			},
			TestAppFactory: func(t testing.TB) *tests.TestApp { return app },
		}
		scenario.Test(t)
	})

	t.Run("should execute a policy in dry-run mode", func(t *testing.T) {
		app := testhelpers.NewTestApp(t)
		defer app.Cleanup()

		policyID := testhelpers.CreatePolicyRecord(t, app)
		token := testhelpers.SuperuserToken(t, app, "admin@test.com", "testpass1234")

		scenario := tests.ApiScenario{
			Name:   "dry run execution",
			Method: http.MethodPost,
			URL:    "/api/pr-scheduler/execute",
			Body:   strings.NewReader(`{"policy_id":"` + policyID + `","dry_run":true}`),
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": token,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"run_id"`,
				`"policy_id"`,
				`"dry_run":true`,
			},
			TestAppFactory: func(t testing.TB) *tests.TestApp { return app },
		}
		scenario.Test(t)
	})

	t.Run("should reject missing policy_id", func(t *testing.T) {
		app := testhelpers.NewTestApp(t)
		defer app.Cleanup()

		token := testhelpers.SuperuserToken(t, app, "admin@test.com", "testpass1234")

		scenario := tests.ApiScenario{
			Name:   "missing policy_id",
			Method: http.MethodPost,
			URL:    "/api/pr-scheduler/execute",
			Body:   strings.NewReader(`{"dry_run":true}`),
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": token,
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`policy_id`,
			},
			TestAppFactory: func(t testing.TB) *tests.TestApp { return app },
		}
		scenario.Test(t)
	})
}

func TestWebhookRoute(t *testing.T) {
	t.Run("should accept webhook and return 202", func(t *testing.T) {
		app := testhelpers.NewTestApp(t)
		defer app.Cleanup()

		token := testhelpers.SuperuserToken(t, app, "admin@test.com", "testpass1234")

		scenario := tests.ApiScenario{
			Name:   "webhook accepted",
			Method: http.MethodPost,
			URL:    "/api/pr-scheduler/webhook",
			Body:   strings.NewReader(`{"event":"push","ref":"refs/heads/main"}`),
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": token,
			},
			ExpectedStatus: 202,
			ExpectedContent: []string{
				`"status":"accepted"`,
			},
			TestAppFactory: func(t testing.TB) *tests.TestApp { return app },
		}
		scenario.Test(t)
	})
}
