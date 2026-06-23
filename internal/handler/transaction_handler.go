package handler

import (
	"net/http"
	"wallet-service/internal/domain"
	"wallet-service/internal/service"

	"github.com/gin-gonic/gin"
)

// TransactionHandler holds the service layer
type TransactionHandler struct {
	service *service.TransactionService
}

func NewTransactionHandler(service *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		service: service,
	}
}

// TopUp is the HTTP endpoint function (e.g., POST /api/v1/transactions/topup)
func (h *TransactionHandler) TopUp(c *gin.Context) {
	var req domain.TopUpRequest

	// Bind and validate the JSON
	// ShouldBindJSON reads the request and checks the 'binding' tags
	if err := c.ShouldBindJSON(&req); err != nil {
		// If validation fails (e.g., missing amount, invalid UUID), return 400 Bad Request
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Pass to the service layer
	// we pass c.Request.Context() so if the user closes their browser,
	// the database transaction can be cleanly canceled!
	err := h.service.ProcessTopUp(c.Request.Context(), req)
	if err != nil {
		if err.Error() == "destination wallet not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process top-up"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Top-up successful",
		"reference_id": req.ReferenceID,
	})
}
