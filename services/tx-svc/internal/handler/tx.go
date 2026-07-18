package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"spendanalyzer.com/tx/api"
	"spendanalyzer.com/tx/backend"
)

type TxHandler struct {
	Service *backend.TxService
}

func NewTxHandler(service *backend.TxService) *TxHandler {
	return &TxHandler{Service: service}
}

// writeError maps a backend error to an HTTP status code. Plaid validation
// errors (bad/expired token) are the client's fault (400), an unrecognized
// account is a 404; everything else is treated as a server error.
func writeError(c *gin.Context, err error) {
	var plaidErr *backend.PlaidValidationError
	if errors.As(err, &plaidErr) {
		c.JSON(http.StatusBadRequest, gin.H{"error": plaidErr.Error()})
		return
	}
	if errors.Is(err, backend.ErrAccountNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

// PullTransactions implements api.ServerInterface.
func (h *TxHandler) PullTransactions(c *gin.Context) {
	var req api.PullTransactionsRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	pulled, err := h.Service.PullTransactions(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, api.PullTransactionsResponse{Pulled: pulled})
}

// GetTransactions implements api.ServerInterface.
func (h *TxHandler) GetTransactions(c *gin.Context, params api.GetTransactionsParams) {
	transactions, err := h.Service.GetTransactions(params.UserId)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, api.GetTransactionsResponse{Transactions: transactions})
}

// GetSpendByCategory implements api.ServerInterface.
func (h *TxHandler) GetSpendByCategory(c *gin.Context, params api.GetSpendByCategoryParams) {
	resp, err := h.Service.GetSpendByCategory(params.UserId)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
