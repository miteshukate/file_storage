package api

import (
	api "file_storage/pkg/api/controllers"
	"file_storage/pkg/security"
	"file_storage/pkg/storage"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"log"
)

type AuthController struct {
	Authenticator  security.Authenticator
	UserRepository storage.UserRepository
}

func NewAuthController(authn security.Authenticator, userRepo storage.UserRepository) *AuthController {
	return &AuthController{Authenticator: authn, UserRepository: userRepo}
}

func (a AuthController) ChangePassword(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a AuthController) GetMe(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a AuthController) Login(c *gin.Context) {
	email := "john@example.com"
	password := "password123"
	// Call authenticator to generate token with user info (this is a simplified example, you should validate credentials properly)
	log.Print("Login attempt for email: ", email)
	log.Print("Password: ", password)
	// Get user from repository
	user, err := a.UserRepository.GetUserByEmail(email)
	if err != nil {
		log.Print("Error fetching user: ", err)
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}
	// Validate encrypted password
	if !VerifyPassword(password, user.PasswordHash) {
		log.Print("Invalid password for email: ", email)
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := a.Authenticator.GenerateToken(c.Request.Context(), map[string]any{
		"id":    user.UserId,
		"name":  user.Email,
		"email": user.Email,
		"roles": []string{"user"}, // Assign roles as needed
		"attrs": map[string]any{}, // Add any additional attributes if needed
	})
	if err != nil {
		log.Print("Error generating token: ", err)
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}
	refreshToken, err := a.Authenticator.GenerateRefreshToken(c.Request.Context(), map[string]any{
		"id":    user.UserId,
		"name":  user.Email,
		"email": user.Email,
		"roles": []string{"user"}, // Assign roles as needed
		"attrs": map[string]any{}, // Add any additional attributes if needed
	})
	loginResponse := api.AuthResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // Set token expiration as needed
		User: api.UserSummary{
			Id:    email,
			Email: email,
		},
	}
	c.JSON(200, loginResponse)
}

func HashPassword(password string) (string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(passwordHash), nil
}

func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		log.Println(err)
	}
	return err == nil
}

func (a AuthController) Logout(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a AuthController) RefreshToken(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a AuthController) RegisterUser(c *gin.Context) {
	// Create LoginRequest from formData
	var loginRequest api.LoginRequest
	// create LoginRequest by automatically filling from form gin context
	err := c.ShouldBind(&loginRequest)
	if err != nil {
		// return error
		log.Println(err)
		c.JSON(401, gin.H{"error": "Invalid credentials"})
	}
	// Insert into User table
	hash, err := HashPassword(loginRequest.Password)
	if err != nil {
		log.Println(err)
		c.JSON(401, gin.H{"error": "Invalid credentials"})
	}
	user := storage.User{
		Email:        loginRequest.Email,
		PasswordHash: hash,
	}
	createUser, err := a.UserRepository.CreateUser(&user)
	if err != nil {
		log.Println(err)
		c.JSON(401, gin.H{"error": "Error while creating user"})
	}
	println(createUser != nil)
	c.JSON(200, gin.H{"status": "ok"})
}
