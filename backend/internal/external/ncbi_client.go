package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

// NCBIClient provides access to NCBI E-utilities for PubChem and PubMed.
type NCBIClient struct {
	httpClient *http.Client
	apiKey     string
	limiter    *rate.Limiter
	baseURL    string
}

// NewNCBIClient creates a new NCBI client with optional API key.
// With API key: 10 req/s, without: 3 req/s (NCBI policy).
func NewNCBIClient(apiKey string) *NCBIClient {
	rps := 3.0
	if apiKey != "" {
		rps = 10.0
	}
	return &NCBIClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiKey:     apiKey,
		limiter:    rate.NewLimiter(rate.Limit(rps), 1),
		baseURL:    "https://eutils.ncbi.nlm.nih.gov/entrez/eutils",
	}
}

// ESearchResult represents ESearch response.
type ESearchResult struct {
	IDList []string `json:"idlist"`
	Count  string   `json:"count"`
}

type eSearchResponse struct {
	ESearchResult ESearchResult `json:"esearchresult"`
}

// ESearch searches a database and returns IDs.
func (c *NCBIClient) ESearch(ctx context.Context, db, term string, maxResults int) (*ESearchResult, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("db", db)
	params.Set("term", term)
	params.Set("retmode", "json")
	params.Set("retmax", fmt.Sprintf("%d", maxResults))
	if c.apiKey != "" {
		params.Set("api_key", c.apiKey)
	}

	reqURL := fmt.Sprintf("%s/esearch.fcgi?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("esearch: status %d", resp.StatusCode)
	}

	var result eSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.ESearchResult, nil
}

// PubChemCompound represents basic compound info from PubChem.
type PubChemCompound struct {
	CID              string `json:"cid"`
	IUPACName        string `json:"iupac_name,omitempty"`
	MolecularFormula string `json:"molecular_formula,omitempty"`
	Title            string `json:"title,omitempty"`
}

// GetPubChemCompound retrieves compound info by CID.
func (c *NCBIClient) GetPubChemCompound(ctx context.Context, cid string) (*PubChemCompound, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("https://pubchem.ncbi.nlm.nih.gov/rest/pug/compound/cid/%s/property/IUPACName,MolecularFormula,Title/JSON", cid)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pubchem: status %d", resp.StatusCode)
	}

	var raw struct {
		PropertyTable struct {
			Properties []struct {
				CID              int    `json:"CID"`
				IUPACName        string `json:"IUPACName"`
				MolecularFormula string `json:"MolecularFormula"`
				Title            string `json:"Title"`
			} `json:"Properties"`
		} `json:"PropertyTable"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	if len(raw.PropertyTable.Properties) == 0 {
		return nil, nil
	}

	p := raw.PropertyTable.Properties[0]
	return &PubChemCompound{
		CID:              fmt.Sprintf("%d", p.CID),
		IUPACName:        p.IUPACName,
		MolecularFormula: p.MolecularFormula,
		Title:            p.Title,
	}, nil
}

// SearchDrug searches PubChem for a drug by name and returns first match CID.
func (c *NCBIClient) SearchDrug(ctx context.Context, name string) (string, error) {
	result, err := c.ESearch(ctx, "pccompound", name, 1)
	if err != nil {
		return "", err
	}
	if len(result.IDList) == 0 {
		return "", nil
	}
	return result.IDList[0], nil
}

// VerifyDrug searches PubChem and returns compound info if found.
func (c *NCBIClient) VerifyDrug(ctx context.Context, name string) (*PubChemCompound, error) {
	cid, err := c.SearchDrug(ctx, name)
	if err != nil {
		return nil, err
	}
	if cid == "" {
		return nil, nil
	}
	return c.GetPubChemCompound(ctx, cid)
}

// SearchPubMed searches PubMed for articles and returns PMIDs.
func (c *NCBIClient) SearchPubMed(ctx context.Context, query string, maxResults int) ([]string, error) {
	result, err := c.ESearch(ctx, "pubmed", query, maxResults)
	if err != nil {
		return nil, err
	}
	return result.IDList, nil
}

// PubMedArticle represents basic article info.
type PubMedArticle struct {
	PMID    string `json:"pmid"`
	Title   string `json:"title"`
	Authors string `json:"authors"`
	Journal string `json:"journal"`
	PubDate string `json:"pubdate"`
}

// ESummary fetches article summaries by PMIDs.
func (c *NCBIClient) ESummary(ctx context.Context, db string, ids []string) ([]PubMedArticle, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("db", db)
	params.Set("id", strings.Join(ids, ","))
	params.Set("retmode", "json")
	if c.apiKey != "" {
		params.Set("api_key", c.apiKey)
	}

	reqURL := fmt.Sprintf("%s/esummary.fcgi?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("esummary: status %d", resp.StatusCode)
	}

	var raw struct {
		Result map[string]json.RawMessage `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	var articles []PubMedArticle
	for _, id := range ids {
		rawArticle, ok := raw.Result[id]
		if !ok {
			continue
		}
		var a struct {
			UID        string `json:"uid"`
			Title      string `json:"title"`
			AuthorList []struct {
				Name string `json:"name"`
			} `json:"authors"`
			Source  string `json:"source"`
			PubDate string `json:"pubdate"`
		}
		if err := json.Unmarshal(rawArticle, &a); err != nil {
			continue
		}
		var authors []string
		for _, auth := range a.AuthorList {
			authors = append(authors, auth.Name)
		}
		articles = append(articles, PubMedArticle{
			PMID:    a.UID,
			Title:   a.Title,
			Authors: strings.Join(authors, ", "),
			Journal: a.Source,
			PubDate: a.PubDate,
		})
	}

	return articles, nil
}
