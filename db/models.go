package db

import "time"

type Invoice struct {
	InvoiceNumber         int64     `json:"invoice_number"`
	CustomerName          string    `json:"customer_name"`
	CustomerEmail         string    `json:"customer_email"`
	CustomerPhone         string    `json:"customer_phone"`
	CustomerAddress       string    `json:"customer_address"`
	SenderName            string    `json:"sender_name"`
	SenderEmail           string    `json:"sender_email"`
	SenderPhone           string    `json:"sender_phone"`
	SenderAddress         string    `json:"sender_address"`
	IssueDate             string    `json:"issue_date"`
	DueDate               string    `json:"due_date"`
	Status                string    `json:"status"`
	Subtotal              int64     `json:"subtotal"`
	DiscountRateInPercent string    `json:"discount_rate_in_percent"`
	Discount              int64     `json:"discount"`
	TotalAmount           int64     `json:"total_amount"`
	PaymentInfo           string    `json:"payment_info"`
	BillingCurrency       string    `json:"billing_currency"`
	Note                  string    `json:"note"`
	CreatedAt             time.Time `json:"created_at"`
}

type LineItem struct {
	ID            int64  `json:"id"`
	InvoiceNumber int64  `json:"invoice_number"`
	Description   string `json:"description"`
	Quantity      int64  `json:"quantity"`
	UnitPrice     int64  `json:"unit_price"`
	TotalPrice    int64  `json:"total_price"`
}
