package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/varunarora1606/Probo/internal/database"
	"github.com/varunarora1606/Probo/internal/models"
)

type signupReq struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}
type signinReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}
type logoutReq struct {
	Id string `json:"id" binding:"required"`
}

func Signup(c *gin.Context) {
	var req signupReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Something is missing in request", "error": err.Error()})
		return
	}

	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	// TODO: Hash password

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Db error", "error": err.Error()})
		return
	}

	user.Password = ""

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "user": user})
}

func Signin(c *gin.Context) {
	var req signinReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User

	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Db error", "error": err.Error()})
		return
	}

	// TODO: Check hashed password

	req.Password = ""

	c.JSON(http.StatusCreated, gin.H{"message": "User loggedin successfully", "user": req})
}

func Logout(c *gin.Context) {
	// TODO: Add cookie based and jwt
	var user logoutReq

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "user": user})
}
