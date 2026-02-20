package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"file_storage/pkg/api"
	"file_storage/pkg/security"
	"file_storage/pkg/storage"
)

func main() {
	// Init services
	svc := storage.NewMinioDocumentService()
	repo := storage.NewMongoContentRepository()
	if err := repo.EnsureIndexes(context.Background()); err != nil {
		log.Printf("warn: failed to ensure indexes: %v", err)
	}

	// Security wiring
	cfg := security.FromEnv()
	var authn security.Authenticator
	var authz security.Authorizer
	if cfg.Enabled {
		// AuthN
		if cfg.AuthNProvider == "jwt" {
			ja, err := security.NewJWTAuthenticatorHS256(cfg.HS256Secret, cfg.Issuer, cfg.Audience)
			if err != nil {
				log.Fatalf("authn init: %v", err)
			}
			authn = ja
		}
		// AuthZ
		switch cfg.AuthZMode {
		case "noop", "":
			authz = security.NoopAuthorizer{}
		default:
			authz = security.NoopAuthorizer{}
		}
	}

	fileController := api.NewFileController(svc, repo)
	r := gin.Default()

	api.RegisterRoutes(r, fileController, authn, authz)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server exit: %v", err)
	}
}
