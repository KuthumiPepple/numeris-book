package api

import (
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
