package service

import (
	"log"
	"time"

	"github.com/medical-app/backend/internal/external"
	"github.com/medical-app/backend/internal/repository"
	"github.com/medical-app/backend/pkg/crypto"
)

type Services struct {
	Auth     *AuthService
	Survey   *SurveyService
	Drug     *DrugService
	Therapy  *TherapyService
	AIAdvice *AIAdviceService
}

type Deps struct {
	Repos *repository.Repositories

	JWTSecret        string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration

	EncryptionKey string

	NCBIApiKey      string
	YandexGPTApiKey string
	YandexIAMToken  string
	YandexFolderID  string
	YandexGPTModel  string
}

func NewServices(d Deps) *Services {
	encryptor, _ := crypto.NewEncryptor(d.EncryptionKey)

	// Initialize external clients
	ncbiClient := external.NewNCBIClient(d.NCBIApiKey)
	var gptClient *external.YandexGPTClient
	if d.YandexFolderID != "" && (d.YandexGPTApiKey != "" || d.YandexIAMToken != "") {
		if d.YandexGPTApiKey != "" {
			gptClient = external.NewYandexGPTClient(d.YandexGPTApiKey, d.YandexFolderID)
			log.Printf("[Services] YandexGPT client initialized with API key, folder: %s, model: %s", d.YandexFolderID, d.YandexGPTModel)
		} else {
			gptClient = external.NewYandexGPTClientWithIAMToken(d.YandexIAMToken, d.YandexFolderID)
			log.Printf("[Services] YandexGPT client initialized with IAM token, folder: %s", d.YandexFolderID)
		}
		if d.YandexGPTModel != "" {
			gptClient.SetModel(d.YandexGPTModel)
		}
	} else {
		log.Printf("[Services] YandexGPT client NOT initialized (missing YANDEX_FOLDER_ID=%q or API key=%v/IAM=%v)", d.YandexFolderID, d.YandexGPTApiKey != "", d.YandexIAMToken != "")
	}

	authSvc := NewAuthService(AuthDeps{
		UserRepo:         d.Repos.User,
		RefreshTokenRepo: d.Repos.RefreshToken,
		RoleRepo:         d.Repos.Role,
		JWTSecret:        d.JWTSecret,
		AccessTTL:        d.JWTAccessExpiry,
		RefreshTTL:       d.JWTRefreshExpiry,
	})

	surveySvc := NewSurveyService(SurveyDeps{
		TemplateRepo: d.Repos.SurveyTemplate,
		ResponseRepo: d.Repos.SurveyResponse,
		GPTClient:    gptClient,
	})

	drugSvc := NewDrugService(DrugDeps{
		Repo:       d.Repos.Drug,
		NCBIClient: ncbiClient,
	})
	therapySvc := NewTherapyService(TherapyDeps{Repo: d.Repos.TherapyLog, PatientRepo: d.Repos.Patient})
	aiAdviceSvc := NewAIAdviceService(AIAdviceDeps{TemplateRepo: d.Repos.SurveyTemplate, PatientRepo: d.Repos.Patient, AdviceRepo: d.Repos.AIAdvice, GPTClient: gptClient})

	_ = encryptor // will be used when PatientService is implemented

	return &Services{
		Auth:     authSvc,
		Survey:   surveySvc,
		Drug:     drugSvc,
		Therapy:  therapySvc,
		AIAdvice: aiAdviceSvc,
	}
}
