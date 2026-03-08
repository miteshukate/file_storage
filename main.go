package main

import (
	"context"
	"file_storage/pkg/api"
	sw "file_storage/pkg/api/controllers"
	"file_storage/pkg/security"
	"file_storage/pkg/storage"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	//// Init services
	//svc := storage.NewMinioDocumentService()
	//repo := storage.NewMongoContentRepository()
	//if err := repo.EnsureIndexes(context.Background()); err != nil {
	//	log.Printf("warn: failed to ensure indexes: %v", err)
	//}
	//
	//// Security wiring
	//cfg := security.FromEnv()
	//var authn security.Authenticator
	//var authz security.Authorizer
	//if cfg.Enabled {
	//	// AuthN
	//	if cfg.AuthNProvider == "jwt" {
	//		ja, err := security.NewJWTAuthenticatorHS256(cfg.HS256Secret, cfg.Issuer, cfg.Audience)
	//		if err != nil {
	//			log.Fatalf("authn init: %v", err)
	//		}
	//		authn = ja
	//	}
	//	// AuthZ
	//	switch cfg.AuthZMode {
	//	case "noop", "":
	//		authz = security.NoopAuthorizer{}
	//	default:
	//		authz = security.NoopAuthorizer{}
	//	}
	//}
	//
	//fileController := api.NewFileController(svc, repo)
	//r := gin.Default()
	//
	//api.RegisterRoutes(r, fileController, authn, authz)
	//
	//if err := r.Run(":8080"); err != nil {
	//	log.Fatalf("server exit: %v", err)
	//}

	svc := storage.NewMinioDocumentService()
	repo := storage.NewMongoContentRepository()
	searchService := storage.NewOpenSearchService()

	if err := repo.EnsureIndexes(context.Background()); err != nil {
		log.Printf("warn: failed to ensure indexes: %v", err)
	}

	if err := searchService.EnsureIndex(context.Background()); err != nil {
		log.Printf("warn: failed to ensure opensearch index: %v", err)
	}

	// Security wiring
	/*cfg := security.FromEnv()
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
	}*/

	fileController := api.NewFileController(svc, repo, searchService)
	authenticator := security.NewJWTAuthenticator("mysecretkey", "myapp", "myaudience")
	authController := api.NewAuthController(authenticator)
	routes := sw.ApiHandleFunctions{FilesAPI: fileController, AuthAPI: authController}

	log.Printf("Server started")

	engine := gin.Default()
	// Apply JWT middleware with login endpoint excluded at engine level
	excludedPaths := []string{"/v1/auth/login"}
	engine.Use(security.JWTAuthMiddlewareWithExclusions(authenticator, excludedPaths))

	// Now register routes with middleware applied
	engine = sw.NewRouterWithGinEngine(engine, routes)

	if err := engine.Run(":8082"); err != nil {
		log.Fatalf("server exit: %v", err)
	}
}
