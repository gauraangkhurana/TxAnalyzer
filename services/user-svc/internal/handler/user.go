package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"spendanalyzer.com/user/api"
	"spendanalyzer.com/user/backend"
)

type UserHandler struct {
	Service *backend.UserService
}

func NewUserHandler(service *backend.UserService) *UserHandler {
	return &UserHandler{Service: service}
}

func writeError(c *gin.Context, err error) {
	if errors.Is(err, backend.ErrUserNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

// CreateUser implements api.ServerInterface.
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req api.CreateUserRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := h.Service.CreateUser(req.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// GetUser implements api.ServerInterface.
func (h *UserHandler) GetUser(c *gin.Context, userID string) {
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID"})
		return
	}

	resp, err := h.Service.GetUser(userIDInt)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
