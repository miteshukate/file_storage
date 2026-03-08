package api

import (
	api "file_storage/pkg/api/controllers"
	"file_storage/pkg/security"
	"github.com/gin-gonic/gin"
	"log"
)

type AuthController struct {
	Authenticator security.Authenticator
}

func NewAuthController(authn security.Authenticator) *AuthController {
	return &AuthController{Authenticator: authn}
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
	email := c.PostForm("email")
	password := c.PostForm("password")
	// Call authenticator to generate token with user info (this is a simplified example, you should validate credentials properly)
	log.Print("Login attempt for email: ", email)
	log.Print("Password: ", password)
	token, err := a.Authenticator.GenerateToken(c.Request.Context(), map[string]any{
		"id":    email,
		"name":  email,
		"roles": []string{"user"}, // Assign roles as needed
		"attrs": map[string]any{}, // Add any additional attributes if needed
	})
	if err != nil {
		log.Print("Error generating token: ", err)
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}
	loginResponse := api.AuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   3600, // Set token expiration as needed
		User: api.UserSummary{
			Id:    email,
			Email: email,
		},
	}
	c.JSON(200, loginResponse)
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
	//TODO implement me
	panic("implement me")
}
