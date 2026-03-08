package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"file_storage/pkg/api"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

// NewOpenSearchClient creates an OpenSearch client using environment configuration.
func NewOpenSearchClient(ctx context.Context, hosts []string) (*opensearch.Client, error) {
	cfg := opensearch.Config{
		Addresses: hosts,
	}
	return opensearch.NewClient(cfg)
}

// OpenSearchService handles indexing and searching documents
type OpenSearchService struct {
	client    *opensearch.Client
	indexName string
}

// NewOpenSearchService constructs an OpenSearchService from environment configuration.
func NewOpenSearchService() *OpenSearchService {
	hosts := []string{getenv("OPENSEARCH_HOST", "http://localhost:9200")}
	if customHosts := os.Getenv("OPENSEARCH_HOSTS"); customHosts != "" {
		// Parse comma-separated hosts: "http://host1:9200,http://host2:9200"
		hosts = []string{customHosts}
	}

	client, err := NewOpenSearchClient(context.Background(), hosts)
	if err != nil {
		log.Fatalf("failed to create opensearch client: %v", err)
	}

	indexName := getenv("OPENSEARCH_INDEX", "file-search-index")

	return &OpenSearchService{
		client:    client,
		indexName: indexName,
	}
}

// EnsureIndex creates the index with appropriate mappings if it doesn't exist
func (s *OpenSearchService) EnsureIndex(ctx context.Context) error {
	// Check if index exists
	exists, err := s.indexExists(ctx)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	// Create index with mappings
	indexSettings := map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 1,
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "keyword",
				},
				"filename": map[string]interface{}{
					"type": "text",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"contentType": map[string]interface{}{
					"type": "keyword",
				},
				"extractedText": map[string]interface{}{
					"type": "text",
					"analyzer": map[string]interface{}{
						"type": "standard",
					},
				},
				"createdAt": map[string]interface{}{
					"type": "date",
				},
				"updatedAt": map[string]interface{}{
					"type": "date",
				},
				"fileSize": map[string]interface{}{
					"type": "long",
				},
			},
		},
	}

	body, err := json.Marshal(indexSettings)
	if err != nil {
		return fmt.Errorf("failed to marshal index settings: %w", err)
	}

	req := opensearchapi.IndicesCreateRequest{
		Index: s.indexName,
		Body:  bytes.NewReader(body),
	}

	resp, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	log.Printf("OpenSearch index '%s' created successfully", s.indexName)
	return nil
}

// indexExists checks if the index exists
func (s *OpenSearchService) indexExists(ctx context.Context) (bool, error) {
	req := opensearchapi.IndicesExistsRequest{
		Index: []string{s.indexName},
	}

	resp, err := req.Do(ctx, s.client)
	if err != nil {
		return false, fmt.Errorf("failed to check index existence: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200, nil
}

// IndexDocument adds or updates a document in the search index
func (s *OpenSearchService) IndexDocument(ctx context.Context, doc *api.OpenSearchIndexDocument) error {
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	req := opensearchapi.IndexRequest{
		Index:      s.indexName,
		DocumentID: doc.ID,
		Body:       bytes.NewReader(body),
	}

	resp, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// IndexDocumentWithExtraction extracts text from reader and indexes the document
func (s *OpenSearchService) IndexDocumentWithExtraction(ctx context.Context, doc *api.OpenSearchIndexDocument, reader io.ReadCloser) error {
	// Extract text from file
	extractedText, err := ExtractTextFromFile(reader, doc.ContentType)
	if err != nil {
		// Log error but continue - not all files are text-extractable
		fmt.Printf("text extraction failed for %s: %v\n", doc.ID, err)
		extractedText = ""
	}

	// Update document with extracted text
	doc.ExtractedText = extractedText

	// Index the document
	return s.IndexDocument(ctx, doc)
}

// SearchByKeyword searches for documents matching the keyword
func (s *OpenSearchService) SearchByKeyword(ctx context.Context, keyword string, limit int) ([]api.SearchResult, error) {
	if limit <= 0 {
		limit = 100
	}

	// Multi-field search query combining filename and extracted text
	query := map[string]interface{}{
		"bool": map[string]interface{}{
			"should": []map[string]interface{}{
				{
					"match": map[string]interface{}{
						"filename": map[string]interface{}{
							"query": keyword,
							"boost": 2.0, // Boost filename matches
						},
					},
				},
				{
					"match": map[string]interface{}{
						"extractedText": map[string]interface{}{
							"query": keyword,
						},
					},
				},
			},
		},
	}

	searchBody := map[string]interface{}{
		"query": query,
		"size":  limit,
	}

	body, err := json.Marshal(searchBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search query: %w", err)
	}

	req := opensearchapi.SearchRequest{
		Index: []string{s.indexName},
		Body:  bytes.NewReader(body),
	}

	resp, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Hits struct {
			Hits []struct {
				ID     string                      `json:"_id"`
				Score  float64                     `json:"_score"`
				Source api.OpenSearchIndexDocument `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	results := make([]api.SearchResult, len(result.Hits.Hits))
	for i, hit := range result.Hits.Hits {
		results[i] = api.SearchResult{
			ID:       hit.Source.ID,
			Score:    hit.Score,
			Filename: hit.Source.Filename,
		}
	}

	return results, nil
}

// DeleteFromIndex removes a document from the search index
func (s *OpenSearchService) DeleteFromIndex(ctx context.Context, docID string) error {
	req := opensearchapi.DeleteRequest{
		Index:      s.indexName,
		DocumentID: docID,
	}

	resp, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to delete document from index: %w", err)
	}
	defer resp.Body.Close()

	// 404 is acceptable (document doesn't exist)
	if resp.StatusCode >= 400 && resp.StatusCode != 404 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// FlushIndex forces an index flush to make all documents immediately searchable
func (s *OpenSearchService) FlushIndex(ctx context.Context) error {
	req := opensearchapi.IndicesFlushRequest{
		Index: []string{s.indexName},
	}

	resp, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to flush index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
