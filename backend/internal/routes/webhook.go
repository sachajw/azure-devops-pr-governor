package routes

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterWebhookRoute registers the POST /api/pr-scheduler/webhook route.
func RegisterWebhookRoute(app core.App) {
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.POST("/api/pr-scheduler/webhook", func(e *core.RequestEvent) error {
			var payload map[string]interface{}
			if err := e.BindBody(&payload); err != nil {
				return e.JSON(http.StatusBadRequest, map[string]string{
					"error": "invalid request body",
				})
			}

			eventType, _ := payload["eventType"].(string)
			log.Printf("webhook received: event_type=%s", eventType)

			// Phase 2: Process Azure DevOps Service Hook events
			// For now, log receipt and acknowledge
			raw, _ := json.Marshal(payload)
			log.Printf("webhook payload: %s", string(raw))

			return e.JSON(http.StatusAccepted, map[string]interface{}{
				"status":     "accepted",
				"event_type": eventType,
			})
		}).Bind(apis.RequireSuperuserAuth())

		return se.Next()
	})
}
