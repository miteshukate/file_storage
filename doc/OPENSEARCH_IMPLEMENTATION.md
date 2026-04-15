# OpenSearch Full-Text Search Implementation Guide

## Overview
This document describes the complete implementation of full-text search functionality for the file storage application using OpenSearch.

## Components Implemented

### 1. OpenSearch Client & Service (`pkg/storage/opensearch.go`)
**Purpose:** Manages OpenSearch cluster connection and indexing operations

**Key Functions:**
- `NewOpenSearchClient()` - Creates OpenSearch client from environment configuration
- `NewOpenSearchService()` - Initializes OpenSearchService with environment variables:
  - `OPENSEARCH_HOST` - OpenSearch endpoint (default: `http://localhost:9200`)
  - `OPENSEARCH_INDEX` - Index name (default: `file-search-index`)

**Features:**
- Automatic index creation with proper schema mapping on startup
- Multi-field search (filename + extracted text)
- Document indexing with unique MongoDB ObjectID
- Index flushing for immediate search availability
- Document deletion from index

**Index Schema:**
- `id` - Keyword field for unique document ID (MongoDB ObjectID)
- `filename` - Text field with keyword subfield for filename search
- `contentType` - Keyword field for file MIME type
- `extractedText` - Text field for searchable content
- `createdAt` & `updatedAt` - Date fields for temporal queries
- `fileSize` - Long field for file size information

### 2. Text Extraction Service (`pkg/storage/extractor.go`)
**Purpose:** Extracts searchable text from various file types

**Supported File Types:**
- **Plain Text** (`.txt`) - Direct content
- **PDF** (`.pdf`) - Uses `github.com/ledongthuc/pdf` library for page-by-page extraction
- **Excel** (`.xls`, `.xlsx`) - Uses `github.com/xuri/excelize/v2` to extract cell content from all sheets

**Key Functions:**
- `ExtractTextFromFile()` - Main entry point that routes to appropriate extractor based on MIME type
- `extractPlainText()` - Handles text files
- `extractPDFText()` - Extracts text from each PDF page
- `extractExcelText()` - Extracts text from all Excel sheets with cell delimiters
- `ExtractSummary()` - Creates concise text summaries (default 500 chars)

**Error Handling:**
- Text extraction failures don't block file upload (graceful degradation)
- Unsupported file types return empty text (indexed by filename only)
- Partial extraction (e.g., some PDF pages fail) continues with available content

### 3. Data Models (`pkg/api/model.go`)
**OpenSearchIndexDocument:**
```go
type OpenSearchIndexDocument struct {
    ID            string    // MongoDB ObjectID
    Filename      string    // Original filename
    ContentType   string    // MIME type
    ExtractedText string    // Searchable text content
    CreatedAt     time.Time // Upload timestamp
    UpdatedAt     time.Time // Last update timestamp
    FileSize      int64     // File size in bytes
}
```

### 4. Repository Interfaces & Methods (`pkg/api/repository.go`)

**SearchService Interface:**
```go
type SearchService interface {
    EnsureIndex(ctx context.Context) error
    IndexDocument(ctx context.Context, doc *OpenSearchIndexDocument) error
    IndexDocumentWithExtraction(ctx context.Context, doc *OpenSearchIndexDocument, reader io.ReadCloser) error
    SearchByKeyword(ctx context.Context, keyword string, limit int) ([]SearchResult, error)
    DeleteFromIndex(ctx context.Context, docID string) error
    FlushIndex(ctx context.Context) error
}
```

**ContentRepository Extension:**
- `FindByIds()` - Fetch multiple documents by their MongoDB ObjectIDs

### 5. File Upload Integration (`pkg/api/controller.go`)

**UploadFile Method Enhancement:**
1. File is saved to MinIO storage
2. Document metadata is saved to MongoDB
3. Async goroutine is spawned to:
   - Retrieve file from storage
   - Extract text based on MIME type
   - Create OpenSearch index document
   - Index document in OpenSearch with `IndexDocumentWithExtraction()`
   - Flush index for immediate availability

**Search Integration in ListDocuments:**
- Check for `search` query parameter
- If present:
  - Query OpenSearch using `SearchByKeyword()`
  - Extract document IDs from search results
  - Fetch full documents from MongoDB using `FindByIds()`
  - Return filtered file responses
- If no search parameter:
  - Return all documents (existing behavior)

### 6. Main Application Initialization (`main.go`)

**Startup Sequence:**
```go
// 1. Initialize storage services
svc := storage.NewMinioDocumentService()
repo := storage.NewMongoContentRepository()
searchService := storage.NewOpenSearchService()

// 2. Ensure indexes exist
repo.EnsureIndexes(context.Background())
searchService.EnsureIndex(context.Background())

// 3. Wire dependencies
fileController := api.NewFileController(svc, repo, searchService)
```

## Usage Examples

### Uploading a File
```bash
curl -X POST http://localhost:8082/v1/files \
  -F "file=@document.pdf" \
  -F "parentId=507f1f77bcf86cd799439011"
```

### Searching Files
```bash
# Search for files containing "invoice"
curl "http://localhost:8082/v1/files?search=invoice"

# Results include all files matching filename OR content
```

## Environment Configuration

```bash
# OpenSearch settings
OPENSEARCH_HOST=http://localhost:9200
OPENSEARCH_INDEX=file-search-index

# MinIO settings
MINIO_ENDPOINT=127.0.0.1:9000
MINIO_BUCKET=test
MINIO_ACCESS_KEY=admin
MINIO_SECRET_KEY=ifyouusethispasswordsupportwilllaughatyou

# MongoDB settings
MONGO_URI=mongodb://localhost:3000
```

## Docker Compose Setup

The `docker-compose.yml` provides a complete OpenSearch cluster:
- **opensearch-node1** & **opensearch-node2** - Clustered OpenSearch nodes
- **opensearch-dashboards** - Web UI for cluster management at `http://localhost:5601`

```bash
# Start OpenSearch cluster
docker-compose up -d

# Verify cluster health
curl http://localhost:9200/_cluster/health
```

## Dependencies Added

```go
github.com/opensearch-project/opensearch-go/v2 v2.x.x  // OpenSearch client
github.com/ledongthuc/pdf v0.0.0-20250511090121...     // PDF text extraction
github.com/xuri/excelize/v2 v2.10.1                     // Excel file handling
```

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     API Request                              │
│        (Upload File or Search)                               │
└──────────────────────┬──────────────────────────────────────┘
                       │
        ┌──────────────┼──────────────┐
        │              │              │
        ▼              ▼              ▼
    ┌────────┐   ┌────────┐   ┌─────────────┐
    │ MinIO  │   │MongoDB │   │ OpenSearch  │
    │Storage │   │Database│   │   Index     │
    └────────┘   └────────┘   └─────────────┘
        │              │              ▲
        │              │              │
        └──────────────┼──────────────┘
                       │
              ┌────────▼────────┐
              │   Text          │
              │   Extraction    │
              │   Service       │
              └─────────────────┘
```

## Data Flow - File Upload

```
1. User uploads file
   ↓
2. File saved to MinIO (returns ETag + LastModified)
   ↓
3. Metadata saved to MongoDB (returns ObjectID)
   ↓
4. Async goroutine spawned:
   a. Stream file from MinIO
   b. Extract text based on MIME type
   c. Create OpenSearchIndexDocument
   d. Index document in OpenSearch
   e. Flush index
   ↓
5. Return FileResponse to user
```

## Data Flow - File Search

```
1. User requests GET /v1/files?search=keyword
   ↓
2. Search in OpenSearch:
   - Search filename (boost 2.0x)
   - Search extracted text (normal weight)
   ↓
3. Get document IDs from results
   ↓
4. Fetch full documents from MongoDB using IDs
   ↓
5. Convert to FileResponse and return
```

## Key Design Decisions

1. **Async Text Extraction**: Non-blocking, doesn't delay file upload response
2. **Graceful Degradation**: Extraction failures don't stop indexing
3. **Multi-field Search**: Filename matches weighted higher (2.0x boost)
4. **Immediate Availability**: Index flush ensures documents are searchable immediately
5. **MongoDB-OpenSearch Sync**: Same ObjectID used in both systems for consistency
6. **Lazy Index Creation**: Index created on first startup if not exists

## Future Enhancements

1. **Full DOCX Support**: Add Word document text extraction
2. **Image OCR**: Integrate OCR for image text extraction
3. **Search Analytics**: Track popular searches and results
4. **Re-indexing**: Bulk re-index capability for schema updates
5. **Index Cleanup**: Delete index documents when files are deleted
6. **Fuzzy Search**: Add fuzzy matching for typo tolerance
7. **Faceted Search**: Add filters by file type, date range, size
8. **Highlighting**: Return text snippets showing search matches

## Testing the Implementation

```bash
# 1. Start OpenSearch cluster
docker-compose up -d

# 2. Start application
go run main.go

# 3. Upload a text file
curl -X POST http://localhost:8082/v1/files \
  -F "file=@test.txt" \
  -F "type=file"

# 4. Search for content
curl "http://localhost:8082/v1/files?search=searchterm"

# 5. View OpenSearch index in dashboard
# Navigate to http://localhost:5601
```

## Troubleshooting

**OpenSearch Connection Failed:**
- Verify OpenSearch is running: `curl http://localhost:9200/_cluster/health`
- Check `OPENSEARCH_HOST` environment variable

**No Search Results:**
- Verify text extraction completed (check server logs)
- Check index flush operation in logs
- Verify document in MongoDB: `db.contents.findOne()`

**Text Extraction Issues:**
- PDF: Check PDF is not encrypted or corrupted
- Excel: Verify file format (xlsx vs xls)
- Check file size isn't too large for memory buffer


