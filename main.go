package main

import (
	"context"
	"file_storage/pkg/api"
	sw "file_storage/pkg/api/controllers"
	"file_storage/pkg/security"
	"file_storage/pkg/storage"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"log"
)

func RunMigrations(db *bun.DB) {
	// Run migration to create users table if it doesn't exist
	_, err := db.NewCreateTable().
		Model((*storage.User)(nil)).
		IfNotExists().
		Exec(context.Background())
	log.Print("Error running users migration: ", err)

	// Run migration to create contents table if it doesn't exist
	_, err = db.NewCreateTable().
		Model((*storage.Content)(nil)).
		IfNotExists().
		Exec(context.Background())
	log.Print("Error running contents migration: ", err)
}
func CreateTestUser(db *bun.DB) {
	// find count for user with email john@example.com
	count, _ := db.NewSelect().Model((*storage.User)(nil)).Where("email = ?",
		"john@example.com").Count(context.Background())
	if count > 0 {
		log.Print("Test user already exists, skipping creation")
		return
	}
	password, _ := api.HashPassword("password123")
	user := &storage.User{Email: "john@example.com", PasswordHash: password}
	_, err := db.NewInsert().Model(user).Exec(context.Background())
	log.Print("Error creating test user: ", err)
}

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
	searchService := storage.NewOpenSearchService()

	if err := searchService.EnsureIndex(context.Background()); err != nil {
		log.Printf("warn: failed to ensure opensearch index: %v", err)
	}

	db := storage.NewDB(storage.DBConfig{Host: "localhost", Port: 5432, User: "postgres", Password: "postgres", DBName: "test", SSLMode: "disable"})
	RunMigrations(db)
	CreateTestUser(db)

	repo := storage.NewContentRepositoryImpl(db)
	userStorage := storage.NewUserRepositoryImpl(db)
	authenticator := security.NewJWTAuthenticator(security.FromEnv())
	fileController := api.NewFileController(svc, repo, searchService)
	authController := api.NewAuthController(authenticator, userStorage)
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
