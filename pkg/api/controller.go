package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"net/http"
	"path"
	"time"
)

// FileController handles HTTP requests for documents using a DocumentService.
type FileController struct {
	documentService DocumentService
	repository      ContentRepository
}

// NewFileController constructs a FileController with dependencies.
func NewFileController(documentService DocumentService, repository ContentRepository) *FileController {
	return &FileController{documentService: documentService, repository: repository}
}

// UploadDocument handles POST /documents for uploading files with metadata.
func (fc *FileController) UploadDocument(c *gin.Context) {
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
	var effectiveParent string
	if parentId != "" {
		if pid, perr := uuid.Parse(parentId); perr == nil {
			// Check if parent exists
			if p, _ := fc.repository.FindById(pid); p != nil {
				effectiveParent = parentId
			}
		}
	}
	meta := Content{
		ParentID:     effectiveParent,
		Name:         name,
		Type:         c.DefaultPostForm("type", "file"),
		Size:         size,
		ContentType:  ctype,
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

	etag, lastMod, err := fc.documentService.UploadDocument(c.Request.Context(), content.ID.Hex(), ctype, f, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	meta.ETag = etag
	meta.LastModified = lastMod
	meta.Status = "Active"
	_, err = fc.repository.Update(meta.ID, struct {
	}{}, &meta)

	c.IndentedJSON(http.StatusCreated, meta)
}

// GetDocument now returns the raw file bytes with appropriate headers.
// GET /documents/:name?download=true
func (fc *FileController) GetDocument(c *gin.Context) {
	name := c.Param("name")
	id, err := primitive.ObjectIDFromHex(name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document ID"})
		return
	}
	contents, err := fc.repository.FindByFilter(struct {
		ID     primitive.ObjectID `bson:"_id"`
		Status string             `bson:"status"`
	}{ID: id, Status: "Active"})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(contents) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}

	reader, meta, err := fc.documentService.StreamDocument(c.Request.Context(), contents[0].ID.Hex())
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
	reader, meta, err := fc.documentService.StreamDocument(c.Request.Context(), name)
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

// ListDocuments handles GET /documents to give content metadata for all the documents
func (fc *FileController) ListDocuments(c *gin.Context) {
	content, err := fc.repository.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, content)
}

// GetPresignedURL returns a presigned URL for direct object access
func (fc *FileController) GetPresignedURL(c *gin.Context) {
	name := c.Param("name")
	u, err := fc.documentService.GetPresignedURL(c.Request.Context(), name, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": u})
}
