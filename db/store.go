package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	Querier
	CreateInvoiceTx(ctx context.Context, arg CreateInvoiceTxParams) (InvoiceResult, error)
	GetInvoice(ctx context.Context, id int64) (InvoiceResult, error)
}

// SQLStore provides all functions to execute SQL queries and transactions.
type SQLStore struct {
	connPool *pgxpool.Pool
	*Queries
}

func NewStore(connPool *pgxpool.Pool) Store {
	return &SQLStore{
		connPool: connPool,
		Queries:  New(connPool),
	}
}

type CreateInvoiceTxParams struct {
	CustomerName    string                 `json:"customer_name"`
	CustomerEmail   string                 `json:"customer_email"`
	CustomerPhone   string                 `json:"customer_phone"`
	CustomerAddress string                 `json:"customer_address"`
	SenderName      string                 `json:"sender_name"`
	SenderEmail     string                 `json:"sender_email"`
	SenderPhone     string                 `json:"sender_phone"`
	SenderAddress   string                 `json:"sender_address"`
	IssueDate       time.Time              `json:"issue_date"`
	DueDate         time.Time              `json:"due_date"`
	Status          string                 `json:"status"`
	Subtotal        int64                  `json:"subtotal"`
	DiscountRate    int64                  `json:"discount_rate"`
	Discount        int64                  `json:"discount"`
	TotalAmount     int64                  `json:"total_amount"`
	PaymentInfo     string                 `json:"payment_info"`
	Items           []InsertLineItemParams `json:"line_items"`
}

type InvoiceResult struct {
	Invoice
	LineItems []LineItem `json:"line_items"`
}

func (store *SQLStore) CreateInvoiceTx(ctx context.Context, arg CreateInvoiceTxParams) (InvoiceResult, error) {
	var result InvoiceResult
	err := store.execTx(
		ctx,
		func(q *Queries) error {

			invoice, err := q.InsertInvoiceRecord(
				ctx,
				InsertInvoiceRecordParams{
					CustomerName:    arg.CustomerName,
					CustomerEmail:   arg.CustomerEmail,
					CustomerPhone:   arg.CustomerPhone,
					CustomerAddress: arg.CustomerAddress,
					SenderName:      arg.SenderName,
					SenderEmail:     arg.SenderEmail,
					SenderPhone:     arg.SenderPhone,
					SenderAddress:   arg.SenderAddress,
					IssueDate:       arg.IssueDate,
					DueDate:         arg.DueDate,
					Status:          arg.Status,
					Subtotal:        arg.Subtotal,
					DiscountRate:    arg.DiscountRate,
					Discount:        arg.Discount,
					TotalAmount:     arg.TotalAmount,
					PaymentInfo:     arg.PaymentInfo,
				},
			)
			if err != nil {
				return err
			}

			result.Invoice = invoice

			for _, item := range arg.Items {
				item.InvoiceNumber = invoice.InvoiceNumber
				lineItem, err := q.InsertLineItem(ctx, item)
				if err != nil {
					return err
				}
				result.LineItems = append(result.LineItems, lineItem)
			}
			return nil
		},
	)
	return result, err
}

const getInvoiceQuery = `
SELECT
	i.invoice_number, i.customer_name, i.customer_email, i.customer_phone,
    i.customer_address, i.sender_name, i.sender_email, i.sender_phone,
    i.sender_address, i.issue_date, i.due_date, i.status,
    i.subtotal, i.discount_rate, i.discount, i.total_amount, i.payment_info,
    i.billing_currency, i.note, i.created_at,
    li.id, li.invoice_number, li.description, li.quantity,
    li.unit_price, li.total_price
FROM
	invoices i
JOIN
	line_items li ON i.invoice_number = li.invoice_number
WHERE
	i.invoice_number = $1
`

func (store *SQLStore) GetInvoice(ctx context.Context, id int64) (InvoiceResult, error) {
	rows, err := store.db.Query(ctx, getInvoiceQuery, id)
	if err != nil {
		return InvoiceResult{}, err
	}
	defer rows.Close()

	var result InvoiceResult
	invoiceInitialized := false

	for rows.Next() {
		var lineItem LineItem
		if !invoiceInitialized {
			err := rows.Scan(
				&result.Invoice.InvoiceNumber,
				&result.Invoice.CustomerName,
				&result.Invoice.CustomerEmail,
				&result.Invoice.CustomerPhone,
				&result.Invoice.CustomerAddress,
				&result.Invoice.SenderName,
				&result.Invoice.SenderEmail,
				&result.Invoice.SenderPhone,
				&result.Invoice.SenderAddress,
				&result.Invoice.IssueDate,
				&result.Invoice.DueDate,
				&result.Invoice.Status,
				&result.Invoice.Subtotal,
				&result.Invoice.DiscountRate,
				&result.Invoice.Discount,
				&result.Invoice.TotalAmount,
				&result.Invoice.PaymentInfo,
				&result.Invoice.BillingCurrency,
				&result.Invoice.Note,
				&result.Invoice.CreatedAt,
				&lineItem.ID,
				&lineItem.InvoiceNumber,
				&lineItem.Description,
				&lineItem.Quantity,
				&lineItem.UnitPrice,
				&lineItem.TotalPrice,
			)
			if err != nil {
				return InvoiceResult{}, err
			}
			invoiceInitialized = true
		} else {
			err := rows.Scan(
				nil, nil, nil, nil, nil,
				nil, nil, nil, nil, nil,
				nil, nil, nil, nil, nil,
				nil, nil, nil, nil, nil,
				&lineItem.ID,
				&lineItem.InvoiceNumber,
				&lineItem.Description,
				&lineItem.Quantity,
				&lineItem.UnitPrice,
				&lineItem.TotalPrice,
			)
			if err != nil {
				return InvoiceResult{}, err
			}
		}
		result.LineItems = append(result.LineItems, lineItem)
	}
	if err = rows.Err(); err != nil {
		return InvoiceResult{}, err
	}
	return result, nil
}
