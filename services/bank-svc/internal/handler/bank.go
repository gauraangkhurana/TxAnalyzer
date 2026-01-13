package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"spendanalyzer.com/bank/api"
	"spendanalyzer.com/bank/backend"
)

type BankHandler struct {
	DB *sql.DB
}

func NewBankHandler(db *sql.DB) *BankHandler {
	return &BankHandler{DB: db}
}

func (h *BankHandler) GetFinancialInstitutions(c *gin.Context) {
	resp, err := backend.GetInstitutions(h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetUserFinancialProfile implements api.ServerInterface.
func (h *BankHandler) GetUserFinancialProfile(c *gin.Context, userID string) {
	if len(userID) < 1 {
		c.JSON(http.StatusBadRequest, g.H{"error": "Invalid request body"})
		return
	}
	var userFinancialProfile api.UserFinancialProfile
	userFinancialProfile, err := backend.GetUserFinancialProfile(h.DB, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
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

	err := backend.SaveBankAcount(h.DB, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	resp, err := backend.AddInstitution(h.DB, req.BankName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}
