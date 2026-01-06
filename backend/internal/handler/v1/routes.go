package v1

import (
	"github.com/gofiber/fiber/v2"

	"github.com/medical-app/backend/internal/handler/middleware"
	"github.com/medical-app/backend/internal/handler/v1/handlers"
	"github.com/medical-app/backend/internal/service"
)

type RouterDeps struct {
	Services *service.Services

	AuthMiddleware  *middleware.AuthMiddleware
	AuditMiddleware *middleware.AuditMiddleware
}

func SetupRoutes(app *fiber.App, deps RouterDeps) {
	api := app.Group("/api")
	v1 := api.Group("/v1")

	authHandler := handlers.NewAuthHandler(deps.Services.Auth, deps.AuditMiddleware)
	surveyHandler := handlers.NewSurveyHandler(deps.Services.Survey, deps.Services.AIAdvice, deps.AuthMiddleware, deps.AuditMiddleware)
	drugHandler := handlers.NewDrugHandler(deps.Services.Drug, deps.AuthMiddleware)
	therapyHandler := handlers.NewTherapyHandler(deps.Services.Therapy, deps.AuthMiddleware)

	// Auth
	v1.Post("/auth/register", authHandler.Register)
	v1.Post("/auth/login", authHandler.Login)
	v1.Post("/auth/refresh", authHandler.Refresh)
	v1.Post("/auth/logout", deps.AuthMiddleware.RequireAuth(), authHandler.Logout)
	v1.Get("/auth/me", deps.AuthMiddleware.RequireAuth(), authHandler.Me)

	// Surveys
	v1.Get("/surveys/templates", deps.AuthMiddleware.OptionalAuth(), surveyHandler.ListTemplates)
	v1.Get("/surveys/templates/:code", deps.AuthMiddleware.OptionalAuth(), surveyHandler.GetTemplateByCode)
	v1.Post("/surveys/:code/calculate", deps.AuthMiddleware.OptionalAuth(), surveyHandler.Calculate)
	v1.Post("/surveys/responses", deps.AuthMiddleware.RequireAuth(), surveyHandler.SubmitResponse)
	v1.Post("/surveys/:code/advice", deps.AuthMiddleware.RequireAuth(), surveyHandler.CreateAdvice)
	v1.Get("/ai/advice", deps.AuthMiddleware.RequireAuth(), surveyHandler.ListAdvice)

	// Drugs
	v1.Get("/drugs", deps.AuthMiddleware.OptionalAuth(), drugHandler.List)
	v1.Get("/drugs/:id", deps.AuthMiddleware.OptionalAuth(), drugHandler.Get)
	v1.Get("/drugs/pubchem/search", deps.AuthMiddleware.RequireAuth(), drugHandler.SearchPubChem)
	v1.Get("/drugs/pubchem/verify", deps.AuthMiddleware.RequireAuth(), drugHandler.VerifyPubChem)
	v1.Get("/drugs/pubmed/search", deps.AuthMiddleware.RequireAuth(), drugHandler.SearchPubMed)

	// Therapy
	v1.Post("/therapy/logs", deps.AuthMiddleware.RequireAuth(), therapyHandler.CreateLog)
	v1.Delete("/therapy/logs/:logId", deps.AuthMiddleware.RequireAuth(), therapyHandler.DeleteLog)
	v1.Get("/patients/:patientId/therapy", deps.AuthMiddleware.RequireAuth(), therapyHandler.ListByPatient)
}
