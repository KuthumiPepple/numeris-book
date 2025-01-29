package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/gin-gonic/gin"
	"github.com/kuthumipepple/numeris-book/db"
)

type createInvoiceRequest struct {
	CustomerName    string                  `json:"customer_name" binding:"required"`
	CustomerEmail   string                  `json:"customer_email" binding:"required,email"`
	CustomerPhone   string                  `json:"customer_phone" binding:"required"`
	CustomerAddress string                  `json:"customer_address" binding:"required"`
	SenderName      string                  `json:"sender_name" binding:"required"`
	SenderEmail     string                  `json:"sender_email" binding:"required,email"`
	SenderPhone     string                  `json:"sender_phone" binding:"required"`
	SenderAddress   string                  `json:"sender_address" binding:"required"`
	IssueDate       string                  `json:"issue_date" binding:"required"`
	DueDate         string                  `json:"due_date" binding:"required"`
	Status          string                  `json:"status" binding:"required"`
	DiscountRate    string                  `json:"discount_rate" binding:"required"`
	PaymentInfo     string                  `json:"payment_info" binding:"required"`
	LineItems       []createLineItemRequest `json:"line_items" binding:"required,dive"`
}

type createLineItemRequest struct {
	Description string `json:"description" binding:"required"`
	Quantity    int64  `json:"quantity" binding:"required,gt=0"`
	UnitPrice   string `json:"unit_price" binding:"required"`
}

type createInvoiceResponse struct {
	InvoiceNumber int64     `json:"invoice_number"`
	CreatedAt     time.Time `json:"created_at"`
}

func (server *Server) createInvoice(c *gin.Context) {
	var req createInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	issueDate, _ := time.Parse(time.DateOnly, req.IssueDate)

	dueDate, _ := time.Parse(time.DateOnly, req.DueDate)

	discountRate := convertRateFromPercentToBasisPoints(req.DiscountRate)

	subtotal := money.New(0, money.USD)

	items := make([]db.InsertLineItemParams, len(req.LineItems))

	for i, v := range req.LineItems {
		unitPrice := money.NewFromFloat(convertStringToFloat64(v.UnitPrice), money.USD)
		totalPrice := unitPrice.Multiply(v.Quantity)
		items[i] = db.InsertLineItemParams{
			Description: v.Description,
			Quantity:    v.Quantity,
			UnitPrice:   unitPrice.Amount(),
			TotalPrice:  totalPrice.Amount(),
		}
		subtotal, _ = subtotal.Add(totalPrice)
	}

	parts, _ := subtotal.Allocate(discountRate, 10000-discountRate)
	discount, totalAmount := parts[0], parts[1]

	arg := db.CreateInvoiceTxParams{
		CustomerName:    req.CustomerName,
		CustomerEmail:   req.CustomerEmail,
		CustomerPhone:   req.CustomerPhone,
		CustomerAddress: req.CustomerAddress,
		SenderName:      req.SenderName,
		SenderEmail:     req.SenderEmail,
		SenderPhone:     req.SenderPhone,
		SenderAddress:   req.SenderAddress,
		IssueDate:       issueDate,
		DueDate:         dueDate,
		Status:          req.Status,
		Subtotal:        subtotal.Amount(),
		DiscountRate:    int64(discountRate),
		Discount:        discount.Amount(),
		TotalAmount:     totalAmount.Amount(),
		PaymentInfo:     req.PaymentInfo,
		Items:           items,
	}

	result, err := server.store.CreateInvoiceTx(c, arg)
	if err != nil {
		if errorCode := ErrorCode(err); errorCode == ForeignKeyViolation {
			c.JSON(http.StatusForbidden, errorResponse(err))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(
		http.StatusCreated,
		createInvoiceResponse{
			InvoiceNumber: result.InvoiceNumber,
			CreatedAt:     result.CreatedAt,
		},
	)
}

type getInvoiceRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type getInvoiceResponse struct {
	InvoiceNumber   int64                    `json:"invoice_number"`
	CustomerName    string                   `json:"customer_name"`
	CustomerEmail   string                   `json:"customer_email"`
	CustomerPhone   string                   `json:"customer_phone"`
	CustomerAddress string                   `json:"customer_address"`
	SenderName      string                   `json:"sender_name"`
	SenderEmail     string                   `json:"sender_email"`
	SenderPhone     string                   `json:"sender_phone"`
	SenderAddress   string                   `json:"sender_address"`
	IssueDate       string                   `json:"issue_date"`
	DueDate         string                   `json:"due_date"`
	Status          string                   `json:"status"`
	Subtotal        string                   `json:"subtotal"`
	DiscountRate    string                   `json:"discount_rate"`
	Discount        string                   `json:"discount"`
	TotalAmount     string                   `json:"total_amount"`
	PaymentInfo     string                   `json:"payment_info"`
	BillingCurrency string                   `json:"billing_currency"`
	Note            string                   `json:"note"`
	CreatedAt       string                   `json:"created_at"`
	Items           []getInvoiceResponseItem `json:"items"`
}

type getInvoiceResponseItem struct {
	ID            int64  `json:"id"`
	InvoiceNumber int64  `json:"invoice_number"`
	Description   string `json:"description"`
	Quantity      int64  `json:"quantity"`
	UnitPrice     string `json:"unit_price"`
	TotalPrice    string `json:"total_price"`
}

func (s *Server) getInvoice(c *gin.Context) {
	var req getInvoiceRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	result, err := s.store.GetInvoice(c, req.ID)
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := generateGetInvoiceResponse(result)
	c.JSON(http.StatusOK, response)

}

func generateGetInvoiceResponse(result db.InvoiceResult) getInvoiceResponse {
	items := make([]getInvoiceResponseItem, len(result.LineItems))
	for i, v := range result.LineItems {
		items[i] = getInvoiceResponseItem{
			ID:            v.ID,
			InvoiceNumber: v.InvoiceNumber,
			Description:   v.Description,
			Quantity:      v.Quantity,
			UnitPrice:     money.New(v.UnitPrice, result.BillingCurrency).Display(),
			TotalPrice:    money.New(v.TotalPrice, result.BillingCurrency).Display(),
		}
	}

	return getInvoiceResponse{
		InvoiceNumber:   result.InvoiceNumber,
		CustomerName:    result.CustomerName,
		CustomerEmail:   result.CustomerEmail,
		CustomerPhone:   result.CustomerPhone,
		CustomerAddress: result.CustomerAddress,
		SenderName:      result.SenderName,
		SenderEmail:     result.SenderEmail,
		SenderPhone:     result.SenderPhone,
		SenderAddress:   result.SenderAddress,
		IssueDate:       result.IssueDate.Format(time.DateOnly),
		DueDate:         result.DueDate.Format(time.DateOnly),
		Status:          result.Status,
		Subtotal:        money.New(result.Subtotal, result.BillingCurrency).Display(),
		DiscountRate:    fmt.Sprintf("%s%%", basisPointsToPercent(result.DiscountRate)),
		Discount:        money.New(result.Discount, result.BillingCurrency).Display(),
		TotalAmount:     money.New(result.TotalAmount, result.BillingCurrency).Display(),
		PaymentInfo:     result.PaymentInfo,
		BillingCurrency: result.BillingCurrency,
		Note:            result.Note,
		CreatedAt:       result.CreatedAt.Format(time.RFC3339),
		Items:           items,
	}
}
