package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/medical-app/backend/internal/entity"
	"github.com/medical-app/backend/internal/external"
	"github.com/medical-app/backend/internal/repository"
)

type DrugService struct {
	repo       repository.DrugRepository
	ncbiClient *external.NCBIClient
}

type DrugDeps struct {
	Repo       repository.DrugRepository
	NCBIClient *external.NCBIClient
}

func NewDrugService(d DrugDeps) *DrugService {
	return &DrugService{
		repo:       d.Repo,
		ncbiClient: d.NCBIClient,
	}
}

func (s *DrugService) List(ctx context.Context, search string, limit int) ([]*entity.Drug, error) {
	return s.repo.List(ctx, search, limit)
}

func (s *DrugService) Get(ctx context.Context, id uuid.UUID) (*entity.Drug, error) {
	return s.repo.GetByID(ctx, id)
}

// SearchPubChem searches for a drug in PubChem using NCBI client.
func (s *DrugService) SearchPubChem(ctx context.Context, query string) (string, error) {
	if s.ncbiClient == nil {
		return "", nil
	}
	return s.ncbiClient.SearchDrug(ctx, query)
}

// VerifyDrug checks if a drug exists in PubChem and returns info.
func (s *DrugService) VerifyDrug(ctx context.Context, name string) (*external.PubChemCompound, error) {
	if s.ncbiClient == nil {
		return nil, nil
	}
	return s.ncbiClient.VerifyDrug(ctx, name)
}

// SearchPubMed searches for articles related to a drug in PubMed.
func (s *DrugService) SearchPubMed(ctx context.Context, drugName string) ([]external.PubMedArticle, error) {
	if s.ncbiClient == nil {
		return nil, nil
	}
	// First search for PMIDs
	pmids, err := s.ncbiClient.SearchPubMed(ctx, drugName, 10)
	if err != nil {
		return nil, err
	}
	if len(pmids) == 0 {
		return nil, nil
	}
	// Then fetch summaries
	return s.ncbiClient.ESummary(ctx, "pubmed", pmids)
}
