package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vibetable/backend/internal/api/handlers"
	authmw "github.com/vibetable/backend/internal/api/middleware"
	"github.com/vibetable/backend/internal/automation"
	"github.com/vibetable/backend/internal/migrate"
	"github.com/vibetable/backend/internal/models"
	"github.com/vibetable/backend/internal/realtime"
	"github.com/vibetable/backend/internal/storage"
	"github.com/vibetable/backend/internal/store"
	"github.com/vibetable/backend/internal/webhook"
)

var db *pgxpool.Pool

func main() {
	// Get configuration from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Connect to database
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err = pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer db.Close()

	// Verify connection
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}
	log.Println("Connected to database")

	// Run migrations
	log.Println("Running database migrations...")
	if err := migrate.Run(context.Background(), db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations complete")

	// Initialize router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Routes
	r.Get("/", handleRoot)
	r.Get("/health", handleHealth)

	// Initialize storage backend
	storagePath := os.Getenv("STORAGE_PATH")
	if storagePath == "" {
		storagePath = "./uploads"
	}
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:" + port
	}
	fileStorage, err := storage.NewLocalStorage(storagePath, baseURL)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	log.Printf("File storage initialized at: %s", storagePath)

	// Initialize realtime hub
	hub := realtime.NewHub()
	go hub.Run()
	log.Println("Real-time hub started")

	// Initialize stores
	authStore := store.NewAuthStore(db)
	baseStore := store.NewBaseStore(db)
	tableStore := store.NewTableStore(db, baseStore)
	fieldStore := store.NewFieldStore(db, baseStore, tableStore)
	recordStore := store.NewRecordStore(db, baseStore, tableStore)
	viewStore := store.NewViewStore(db, baseStore, tableStore)
	formStore := store.NewFormStore(db, baseStore, tableStore, recordStore)
	commentStore := store.NewCommentStore(db, baseStore, tableStore, recordStore)
	activityStore := store.NewActivityStore(db, baseStore)
	attachmentStore := store.NewAttachmentStore(db, baseStore, tableStore, recordStore, fileStorage, baseURL)
	automationStore := store.NewAutomationStore(db, baseStore, tableStore)
	apiKeyStore := store.NewAPIKeyStore(db)
	webhookStore := store.NewWebhookStore(db, baseStore)

	// Set hub on stores that need to broadcast
	recordStore.SetHub(hub)
	fieldStore.SetHub(hub)
	tableStore.SetHub(hub)
	viewStore.SetHub(hub)

	// Initialize automation engine
	automationEngine := automation.NewEngine(automationStore, recordStore, fieldStore)

	// Initialize webhook delivery engine
	webhookEngine := webhook.NewDeliveryEngine(webhookStore, tableStore)
	log.Println("Webhook delivery engine initialized")

	// Set automation and webhook callbacks on record store
	recordStore.SetAutomationCallback(func(tableID uuid.UUID, recordID *uuid.UUID, record *models.Record, oldRecord *models.Record, triggerType string, userID uuid.UUID) {
		ctx := context.Background()

		// Trigger automations
		triggerCtx := &automation.TriggerContext{
			TableID:     tableID,
			RecordID:    recordID,
			Record:      record,
			OldRecord:   oldRecord,
			TriggerType: models.TriggerType(triggerType),
			UserID:      userID,
		}
		automationEngine.ProcessTrigger(ctx, triggerCtx)

		// Trigger webhooks
		table, err := tableStore.GetTable(ctx, tableID, userID)
		if err == nil && table != nil {
			var webhookEvent models.WebhookEvent
			switch triggerType {
			case "record_created":
				webhookEvent = models.WebhookEventRecordCreated
			case "record_updated":
				webhookEvent = models.WebhookEventRecordUpdated
			case "record_deleted":
				webhookEvent = models.WebhookEventRecordDeleted
			}
			if webhookEvent != "" {
				webhookCtx := &webhook.DeliveryContext{
					BaseID:    table.BaseID,
					TableID:   tableID,
					RecordID:  recordID,
					Record:    record,
					OldRecord: oldRecord,
					Event:     webhookEvent,
					UserID:    userID,
				}
				webhookEngine.ProcessEvent(ctx, webhookCtx)
			}
		}
	})
	log.Println("Automation engine initialized")

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authStore)
	baseHandler := handlers.NewBaseHandler(baseStore)
	tableHandler := handlers.NewTableHandler(tableStore)
	fieldHandler := handlers.NewFieldHandler(fieldStore)
	recordHandler := handlers.NewRecordHandler(recordStore, activityStore)
	viewHandler := handlers.NewViewHandler(viewStore)
	csvHandler := handlers.NewCSVHandler(recordStore, fieldStore, tableStore)
	formHandler := handlers.NewFormHandler(formStore)
	commentHandler := handlers.NewCommentHandler(commentStore)
	activityHandler := handlers.NewActivityHandler(activityStore)
	attachmentHandler := handlers.NewAttachmentHandler(attachmentStore)
	automationHandler := handlers.NewAutomationHandler(automationStore)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyStore)
	webhookHandler := handlers.NewWebhookHandler(webhookStore, baseStore)
	wsHandler := handlers.NewWebSocketHandler(hub, authStore, baseStore)

	// Initialize middleware
	authMiddleware := authmw.NewAuthMiddleware(authStore)

	// WebSocket route (outside /api/v1 for simplicity)
	r.Get("/ws", wsHandler.ServeWS)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handleHealth)

		// Auth routes (public)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/forgot-password", authHandler.ForgotPassword)
			r.Post("/reset-password", authHandler.ResetPassword)

			// Protected auth routes
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.Required)
				r.Get("/me", authHandler.GetMe)
				r.Patch("/me", authHandler.UpdateMe)
				r.Post("/logout", authHandler.Logout)
			})
		})

		// Base routes (all protected)
		r.Route("/bases", func(r chi.Router) {
			r.Use(authMiddleware.Required)

			r.Get("/", baseHandler.ListBases)
			r.Post("/", baseHandler.CreateBase)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", baseHandler.GetBase)
				r.Patch("/", baseHandler.UpdateBase)
				r.Delete("/", baseHandler.DeleteBase)
				r.Post("/duplicate", baseHandler.DuplicateBase)

				// Collaborators
				r.Get("/collaborators", baseHandler.ListCollaborators)
				r.Post("/collaborators", baseHandler.AddCollaborator)
				r.Patch("/collaborators/{userId}", baseHandler.UpdateCollaborator)
				r.Delete("/collaborators/{userId}", baseHandler.RemoveCollaborator)

				// Activity log
				r.Get("/activity", activityHandler.ListActivitiesForBase)

				// Webhooks
				r.Route("/webhooks", func(r chi.Router) {
					r.Get("/", webhookHandler.ListWebhooks)
					r.Post("/", webhookHandler.CreateWebhook)
				})
			})

			// Tables within a base
			r.Route("/{baseId}/tables", func(r chi.Router) {
				r.Get("/", tableHandler.ListTables)
				r.Post("/", tableHandler.CreateTable)
				r.Put("/reorder", tableHandler.ReorderTables)
			})
		})

		// Table routes (by table ID)
		r.Route("/tables", func(r chi.Router) {
			r.Use(authMiddleware.Required)
			r.Get("/{id}", tableHandler.GetTable)
			r.Patch("/{id}", tableHandler.UpdateTable)
			r.Delete("/{id}", tableHandler.DeleteTable)
			r.Post("/{id}/duplicate", tableHandler.DuplicateTable)

			// Fields within a table
			r.Route("/{tableId}/fields", func(r chi.Router) {
				r.Get("/", fieldHandler.ListFields)
				r.Post("/", fieldHandler.CreateField)
				r.Put("/reorder", fieldHandler.ReorderFields)
			})

			// Views within a table
			r.Route("/{tableId}/views", func(r chi.Router) {
				r.Get("/", viewHandler.ListViews)
				r.Post("/", viewHandler.CreateView)
			})

			// CSV import/export
			r.Route("/{tableId}/csv", func(r chi.Router) {
				r.Post("/preview", csvHandler.Preview)
				r.Post("/import", csvHandler.Import)
				r.Get("/export", csvHandler.Export)
			})

			// Forms within a table
			r.Route("/{tableId}/forms", func(r chi.Router) {
				r.Get("/", formHandler.ListForms)
				r.Post("/", formHandler.CreateForm)
			})

			// Automations within a table
			r.Route("/{tableId}/automations", func(r chi.Router) {
				r.Get("/", automationHandler.ListAutomations)
				r.Post("/", automationHandler.CreateAutomation)
			})
		})

		// View routes (by view ID)
		r.Route("/views", func(r chi.Router) {
			r.Use(authMiddleware.Required)
			r.Get("/{id}", viewHandler.GetView)
			r.Patch("/{id}", viewHandler.UpdateView)
			r.Delete("/{id}", viewHandler.DeleteView)
			r.Patch("/{id}/public", viewHandler.SetViewPublic)
		})

		// Field routes (by field ID)
		r.Route("/fields", func(r chi.Router) {
			r.Use(authMiddleware.Required)
			r.Get("/{id}", fieldHandler.GetField)
			r.Patch("/{id}", fieldHandler.UpdateField)
			r.Delete("/{id}", fieldHandler.DeleteField)
		})

		// Records within a table
		r.Route("/tables/{tableId}/records", func(r chi.Router) {
			r.Use(authMiddleware.Required)
			r.Get("/", recordHandler.ListRecords)
			r.Post("/", recordHandler.CreateRecord)
			r.Post("/bulk", recordHandler.BulkCreateRecords)
		})

		// Record routes (by record ID)
		r.Route("/records", func(r chi.Router) {
			r.Use(authMiddleware.Required)
			r.Get("/{id}", recordHandler.GetRecord)
			r.Put("/{id}", recordHandler.UpdateRecord)
			r.Patch("/{id}", recordHandler.PatchRecord)
			r.Patch("/{id}/color", recordHandler.UpdateRecordColor)
			r.Delete("/{id}", recordHandler.DeleteRecord)

			// Comments on records
			r.Route("/{recordId}/comments", func(r chi.Router) {
				r.Get("/", commentHandler.ListComments)
				r.Post("/", commentHandler.CreateComment)
			})

			// Attachments on records
			r.Route("/{recordId}/fields/{fieldId}/attachments", func(r chi.Router) {
				r.Get("/", attachmentHandler.ListAttachments)
				r.Post("/", attachmentHandler.UploadAttachment)
			})

			// Activity on records
			r.Get("/{recordId}/activity", activityHandler.ListActivitiesForRecord)
		})

		// Attachment routes (by attachment ID)
		r.Route("/attachments", func(r chi.Router) {
			r.Use(authMiddleware.Required)
			r.Get("/{id}", attachmentHandler.GetAttachment)
			r.Get("/{id}/download", attachmentHandler.DownloadAttachment)
			r.Delete("/{id}", attachmentHandler.DeleteAttachment)
		})

		// Comment routes (by comment ID)
		r.Route("/comments", func(r chi.Router) {
			r.Use(authMiddleware.Required)
			r.Get("/{id}", commentHandler.GetComment)
			r.Patch("/{id}", commentHandler.UpdateComment)
			r.Delete("/{id}", commentHandler.DeleteComment)
			r.Post("/{id}/resolve", commentHandler.ResolveComment)
		})

		// Form routes (by form ID)
		r.Route("/forms", func(r chi.Router) {
			r.Use(authMiddleware.Required)
			r.Get("/{id}", formHandler.GetForm)
			r.Patch("/{id}", formHandler.UpdateForm)
			r.Patch("/{id}/fields", formHandler.UpdateFormFields)
			r.Delete("/{id}", formHandler.DeleteForm)
		})

		// Automation routes (by automation ID)
		r.Route("/automations", func(r chi.Router) {
			r.Use(authMiddleware.Required)
			r.Get("/{id}", automationHandler.GetAutomation)
			r.Patch("/{id}", automationHandler.UpdateAutomation)
			r.Delete("/{id}", automationHandler.DeleteAutomation)
			r.Post("/{id}/toggle", automationHandler.ToggleAutomation)
			r.Get("/{id}/runs", automationHandler.ListRuns)
		})

		// API Key routes
		r.Route("/api-keys", func(r chi.Router) {
			r.Use(authMiddleware.Required)
			r.Get("/", apiKeyHandler.ListAPIKeys)
			r.Post("/", apiKeyHandler.CreateAPIKey)
			r.Get("/{id}", apiKeyHandler.GetAPIKey)
			r.Delete("/{id}", apiKeyHandler.DeleteAPIKey)
		})

		// Webhook routes (by webhook ID)
		r.Route("/webhooks", func(r chi.Router) {
			r.Use(authMiddleware.Required)
			r.Get("/{id}", webhookHandler.GetWebhook)
			r.Patch("/{id}", webhookHandler.UpdateWebhook)
			r.Delete("/{id}", webhookHandler.DeleteWebhook)
			r.Get("/{id}/deliveries", webhookHandler.ListDeliveries)
		})

		// Public form routes (no auth required)
		r.Route("/public/forms", func(r chi.Router) {
			r.Get("/{token}", formHandler.GetPublicForm)
			r.Post("/{token}", formHandler.SubmitPublicForm)
		})

		// Public view routes (no auth required)
		r.Route("/public/views", func(r chi.Router) {
			r.Get("/{token}", viewHandler.GetPublicView)
		})
	})

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"name":    "VibeTable API",
		"version": "0.1.0",
		"status":  "running",
	}
	writeJSON(w, http.StatusOK, response)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	dbStatus := "healthy"
	if err := db.Ping(ctx); err != nil {
		dbStatus = "unhealthy"
	}

	response := map[string]interface{}{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
		"database": map[string]string{
			"status": dbStatus,
		},
	}
	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}
