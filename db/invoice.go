package db

import "context"

const InsertInvoiceQuery = `
	INSERT INTO invoices (
		customer_name, customer_email, customer_phone, customer_address,
		sender_name, sender_email, sender_phone, sender_address,
		issue_date, due_date, status, subtotal,
		discount_rate_in_percent, discount, total_amount, payment_info
	) VALUES (
	 $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
	) RETURNING *;
`

type InsertInvoiceParams struct {
	CustomerName          string `json:"customer_name"`
	CustomerEmail         string `json:"customer_email"`
	CustomerPhone         string `json:"customer_phone"`
	CustomerAddress       string `json:"customer_address"`
	SenderName            string `json:"sender_name"`
	SenderEmail           string `json:"sender_email"`
	SenderPhone           string `json:"sender_phone"`
	SenderAddress         string `json:"sender_address"`
	IssueDate             string `json:"issue_date"`
	DueDate               string `json:"due_date"`
	Status                string `json:"status"`
	Subtotal              int64  `json:"subtotal"`
	DiscountRateInPercent string `json:"discount_rate_in_percent"`
	Discount              int64  `json:"discount"`
	TotalAmount           int64  `json:"total_amount"`
	PaymentInfo           string `json:"payment_info"`
}

func (q *Queries) InsertInvoice(ctx context.Context, arg InsertInvoiceParams) (Invoice, error) {
	row := q.db.QueryRow(ctx, InsertInvoiceQuery,
		arg.CustomerName, arg.CustomerEmail, arg.CustomerPhone, arg.CustomerAddress,
		arg.SenderName, arg.SenderEmail, arg.SenderPhone, arg.SenderAddress,
		arg.IssueDate, arg.DueDate, arg.Status, arg.Subtotal, 
		arg.DiscountRateInPercent, arg.Discount, arg.TotalAmount, arg.PaymentInfo,
	)
	var i Invoice
	err := row.Scan(
		&i.InvoiceNumber, &i.CustomerName, &i.CustomerEmail, &i.CustomerPhone, &i.CustomerAddress,
		&i.SenderName, &i.SenderEmail, &i.SenderPhone, &i.SenderAddress,
		&i.IssueDate, &i.DueDate, &i.Status,
		&i.Subtotal, &i.DiscountRateInPercent, &i.Discount, &i.TotalAmount,
		&i.BillingCurrency, &i.PaymentInfo, &i.Note, &i.CreatedAt,
	)
	return i, err
}

const InsertLineItemQuery = `
	INSERT INTO line_items (
		invoice_number, description, quantity, unit_price, total_price
	) VALUES (
	 $1, $2, $3, $4, $5
	) RETURNING *;
`

type InsertLineItemParams struct {
	InvoiceNumber int64  `json:"invoice_number"`
	Description   string `json:"description"`
	Quantity      int64  `json:"quantity"`
	UnitPrice     int64  `json:"unit_price"`
	TotalPrice    int64  `json:"total_price"`
}

func (q *Queries) InsertLineItem(ctx context.Context, arg InsertLineItemParams) (LineItem, error) {
	row := q.db.QueryRow(ctx, InsertLineItemQuery,
		arg.InvoiceNumber, arg.Description, arg.Quantity, arg.UnitPrice, arg.TotalPrice,
	)
	var l LineItem
	err := row.Scan(
		&l.ID, &l.InvoiceNumber, &l.Description, &l.Quantity, &l.UnitPrice, &l.TotalPrice,
	)
	return l, err
}
