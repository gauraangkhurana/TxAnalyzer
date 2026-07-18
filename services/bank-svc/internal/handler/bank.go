package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"spendanalyzer.com/bank/api"
	"spendanalyzer.com/bank/backend"
)

type BankHandler struct {
	Service *backend.BankService
}

func NewBankHandler(service *backend.BankService) *BankHandler {
	return &BankHandler{Service: service}
}

// writeError maps a backend error to an HTTP status code. Plaid validation
// errors (bad/expired token, unknown account) are the client's fault (400);
// everything else is treated as a server error.
func writeError(c *gin.Context, err error) {
	var plaidErr *backend.PlaidValidationError
	if errors.As(err, &plaidErr) {
		c.JSON(http.StatusBadRequest, gin.H{"error": plaidErr.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func (h *BankHandler) GetFinancialInstitutions(c *gin.Context) {
	resp, err := h.Service.GetInstitutions()
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetUserFinancialProfile implements api.ServerInterface.
func (h *BankHandler) GetUserFinancialProfile(c *gin.Context, userID string) {
	if len(userID) < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID"})
		return
	}
	userFinancialProfile, err := h.Service.GetUserFinancialProfile(userIDInt)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, userFinancialProfile)
}

// SaveBankAccount implements api.ServerInterface.
func (h *BankHandler) SaveBankAccount(c *gin.Context) {

	var req api.SaveNewAccountRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := h.Service.SaveBankAccount(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, "201 OK")
}

// SaveNewBank implements api.ServerInterface.
func (h *BankHandler) SaveNewBank(c *gin.Context) {

	var req api.SaveNewBankRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := h.Service.AddInstitution(req.BankName)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// CreateLinkToken implements api.ServerInterface.
func (h *BankHandler) CreateLinkToken(c *gin.Context) {
	var req api.CreateLinkTokenRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := h.Service.CreateLinkToken(c.Request.Context(), req.UserId)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// ExchangePublicToken implements api.ServerInterface.
func (h *BankHandler) ExchangePublicToken(c *gin.Context) {
	var req api.ExchangePublicTokenRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := h.Service.ExchangePublicToken(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}
