package api

import (
	"file_storage/pkg/security"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires the document endpoints.
func RegisterRoutes(r *gin.Engine, fc *FileController, authn security.Authenticator, authz security.Authorizer) {
	grp := r.Group("/documents")
	// Global AuthN for this group
	grp.Use(AuthnMiddleware(authn))
	// List documents: GET /documents
	grp.GET("", Require(authz, "document", "list"), fc.ListDocuments)
	// Upload document: POST /documents
	grp.POST("", Require(authz, "document", "create"), fc.UploadDocument)
	// Get document bytes: GET /documents/:name
	grp.GET(":name", Require(authz, "document", "read"), fc.GetDocument)
	// Stream document: GET /documents/:name/stream
	grp.GET(":name/stream", Require(authz, "document", "read"), fc.StreamDocument)
	// Presigned URL: GET /documents/:name/url?expiry=
	grp.GET(":name/url", Require(authz, "document", "read"), fc.GetPresignedURL)
}
