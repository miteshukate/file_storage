package api

import (
	"context"
	api "file_storage/pkg/api/controllers"
	"file_storage/pkg/storage"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

// FileController handles HTTP requests for documents using a StorageService.
type FileController struct {
	storageService StorageService
	repository     ContentRepository
	searchService  SearchService
}

func (fc *FileController) BulkDownload(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (fc *FileController) BulkUploadFiles(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (fc *FileController) CopyFile(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (fc *FileController) DeleteFile(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (fc *FileController) DownloadFile(c *gin.Context) {
	fileId := c.Param("fileId")
	id, err := strconv.ParseInt(fileId, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document ID"})
		return
	}
	contents, err := fc.repository.FindByFilter(map[string]interface{}{
		"content_id": id,
		"status":     "Active",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(contents) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}

	reader, meta, err := fc.storageService.StreamDocument(c.Request.Context(), strconv.FormatInt(contents[0].ID, 10))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() { _ = reader.Close() }()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", path.Base(contents[0].Name)))

	c.Header("Content-Type", meta.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", meta.Size))
	c.Status(http.StatusOK)
	if _, err := io.Copy(c.Writer, reader); err != nil {
		_ = c.Error(err)
	}
}

func (fc *FileController) GetAllFiles(c *gin.Context) {
	fc.ListDocuments(c)
}

func (fc *FileController) GetFile(c *gin.Context) {
	// Get file id from path
	contentId := c.Param("fileId")
	contentIdUUID, err := uuid.Parse(contentId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file ID"})
		return
	}
	contentMap := map[string]interface{}{
		"content_id": contentIdUUID,
	}
	contents, err := fc.repository.FindByFilter(contentMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(contents) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	// Convert to FileResponse and return
	response := fc.ContentToFileResponse(&contents[0])
	c.JSON(http.StatusOK, response)
}

func (fc *FileController) ListFileVersions(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (fc *FileController) MoveFile(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (fc *FileController) PreviewFile(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (fc *FileController) RenameFile(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (fc *FileController) RestoreFile(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (fc *FileController) RestoreFileVersion(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (fc *FileController) SetFileTags(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (fc *FileController) UploadFileVersion(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

// NewFileController constructs a FileController with dependencies.
func NewFileController(documentService StorageService, repository ContentRepository, searchService SearchService) *FileController {
	return &FileController{storageService: documentService, repository: repository, searchService: searchService}
}

// UploadFile handles POST /documents for uploading files with metadata.
func (fc *FileController) UploadFile(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file field: " + err.Error()})
		return
	}
	f, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot open uploaded file: " + err.Error()})
		return
	}
	defer func() { _ = f.Close() }()

	var name string
	name = fh.Filename
	if name == "" {
		name = c.PostForm("name")
	}

	ctype := fh.Header.Get("Content-Type")
	if ctype == "" {
		// fallback
		ctype = "application/octet-stream"
	}
	size := fh.Size

	parentId := c.PostForm("parentId")
	var effectiveParent *int64
	if parentId != "" {
		parentMap := map[string]interface{}{
			"contentId": parentId,
		}
		parent, err := fc.repository.FindByFilter(parentMap)
		if len(parent) > 1 || err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parentId: " + err.Error()})
			return
		}
		effectiveParent = &parent[0].ID
	}
	//go uuid v7
	contentId, _ := uuid.NewV7()
	meta := storage.Content{
		ParentID:     effectiveParent,
		Name:         name,
		Type:         c.DefaultPostForm("type", "file"),
		Size:         size,
		ContentType:  ctype,
		ContentId:    contentId,
		Status:       "Uploading",
		LastModified: time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	content, err := fc.repository.Create(&meta)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	etag, lastMod, err := fc.storageService.UploadDocument(c.Request.Context(), strconv.FormatInt(content.ID, 10), ctype, f, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	content.ETag = etag
	content.LastModified = lastMod
	content.Status = "Active"
	_, err = fc.repository.Update(content.ID, content)

	// Index document in OpenSearch for full-text search asynchronously
	go fc.indexFileAsync(content)

	response := fc.ContentToFileResponse(content)
	c.JSON(http.StatusCreated, response)
}

func (fc *FileController) GetDocument(c *gin.Context) {
	name := c.Param("name")
	id, err := strconv.ParseInt(name, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document ID"})
		return
	}
	contents, err := fc.repository.FindByFilter(map[string]interface{}{
		"content_id": id,
		"status":     "Active",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(contents) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}

	reader, meta, err := fc.storageService.StreamDocument(c.Request.Context(), strconv.FormatInt(contents[0].ID, 10))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() { _ = reader.Close() }()
	// Set filename to original name

	// Optional download hint
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", path.Base(contents[0].Name)))

	c.Header("Content-Type", meta.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", meta.Size))
	c.Status(http.StatusOK)
	if _, err := io.Copy(c.Writer, reader); err != nil {
		_ = c.Error(err)
	}
}

// StreamDocument handles GET /documents/:name/stream to stream content
func (fc *FileController) StreamDocument(c *gin.Context) {
	name := c.Param("name")
	reader, meta, err := fc.storageService.StreamDocument(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() { _ = reader.Close() }()
	c.Header("Content-Type", meta.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", meta.Size))
	c.Status(http.StatusOK)
	if _, err := io.Copy(c.Writer, reader); err != nil {
		_ = c.Error(err)
	}
}

// ContentToFileResponse converts a Content object to FileResponse with default values for unknown fields
func (fc *FileController) ContentToFileResponse(content *storage.Content) api.FileResponse {
	// Extract file extension from name
	extension := ""
	if idx := strings.LastIndex(content.Name, "."); idx != -1 {
		extension = content.Name[idx+1:]
	}

	idStr := strconv.FormatInt(content.ID, 10)

	// Convert ParentID *int64 to *string for FolderId
	var folderIdStr *string
	if content.ParentID != nil {
		s := strconv.FormatInt(*content.ParentID, 10)
		folderIdStr = &s
	}

	// Create FileResponse with mapped values
	response := api.FileResponse{
		Id:               idStr,
		Name:             content.Name,
		MimeType:         content.ContentType,
		Extension:        extension,
		Size:             content.Size,
		FolderId:         folderIdStr,
		Path:             idStr,
		OwnerId:          idStr,
		Status:           content.Status,
		CurrentVersion:   1,
		Tags:             []string{},
		Checksum:         content.ETag,
		PreviewAvailable: false,
		CreatedAt:        content.CreatedAt,
		UpdatedAt:        content.UpdatedAt,
		DeletedAt:        nil,
	}

	// Set Owner with default values
	response.Owner = api.UserSummary{
		Id:        idStr,
		Email:     "unknown@example.com",
		FirstName: "Unknown",
		LastName:  "User",
		AvatarUrl: nil,
	}

	return response
}

// ListDocuments handles GET /documents to give content metadata for all the documents
func (fc *FileController) ListDocuments(c *gin.Context) {
	// Check for the principal in context (set by JWT middleware)
	principal, exists := c.Get("principal")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	log.Print("principal: ", principal)
	// Check for search query parameter
	searchQuery := c.Query("search")

	var content []storage.Content
	var err error

	if searchQuery != "" {
		// Perform full-text search
		searchResults, err := fc.searchService.SearchByKeyword(c.Request.Context(), searchQuery, 100)
		if err != nil {
			// Log error but fall back to returning all documents
			fmt.Printf("search error: %v\n", err)
			content, err = fc.repository.FindAll()
		} else {
			// Extract IDs from search results
			ids := make([]string, len(searchResults))
			for i, result := range searchResults {
				ids[i] = result.ID
			}

			// Fetch documents by IDs from PostgreSQL
			if len(ids) > 0 {
				// convert string IDs from OpenSearch to int64
				int64Ids := make([]int64, 0, len(ids))
				for _, sid := range ids {
					if n, parseErr := strconv.ParseInt(sid, 10, 64); parseErr == nil {
						int64Ids = append(int64Ids, n)
					}
				}
				content, err = fc.repository.FindByIds(int64Ids)
				if err != nil {
					fmt.Printf("error fetching documents by ids: %v\n", err)
					content = []storage.Content{}
				}
			}
		}
	} else {
		// Get all documents if no search query
		content, err = fc.repository.FindAll()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert Content objects to FileResponse objects
	fileResponses := make([]api.FileResponse, len(content))
	for i, item := range content {
		fileResponses[i] = fc.ContentToFileResponse(&item)
	}

	c.JSON(http.StatusOK, fileResponses)
}

// GetPresignedURL returns a presigned URL for direct object access
func (fc *FileController) GetPresignedURL(c *gin.Context) {
	name := c.Param("name")
	u, err := fc.storageService.GetPresignedURL(c.Request.Context(), name, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": u})
}

// indexFileAsync extracts text from the file and indexes it in OpenSearch asynchronously
func (fc *FileController) indexFileAsync(content *storage.Content) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		idStr := strconv.FormatInt(content.ID, 10)

		// Retrieve the file from storage
		reader, _, err := fc.storageService.StreamDocument(ctx, idStr)
		if err != nil {
			fmt.Printf("failed to stream document for indexing: %v\n", err)
			return
		}
		defer func() { _ = reader.Close() }()

		// Create OpenSearch index document
		indexDoc := &storage.OpenSearchIndexDocument{
			ID:            idStr,
			Filename:      content.Name,
			ContentType:   content.ContentType,
			ExtractedText: "",
			CreatedAt:     content.CreatedAt,
			UpdatedAt:     content.UpdatedAt,
			FileSize:      content.Size,
		}

		// Index the document with text extraction
		if err := fc.searchService.IndexDocumentWithExtraction(ctx, indexDoc, reader); err != nil {
			fmt.Printf("failed to index document in OpenSearch: %v\n", err)
			return
		}

		// Flush the index to make documents immediately searchable
		if err := fc.searchService.FlushIndex(ctx); err != nil {
			fmt.Printf("failed to flush OpenSearch index: %v\n", err)
		}
	}()
}
