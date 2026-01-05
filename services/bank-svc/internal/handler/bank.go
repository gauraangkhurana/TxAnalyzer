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
	panic("unimplemented")
}

// SaveBankAccount implements api.ServerInterface.
func (h *BankHandler) SaveBankAccount(c *gin.Context) {
	panic("unimplemented")
}

// SaveNewBank implements api.ServerInterface.
func (h *BankHandler) SaveNewBank(c *gin.Context) {

	var req api.SaveNewBankRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := backend.AddInstitution(h.DB, req.BankName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}
