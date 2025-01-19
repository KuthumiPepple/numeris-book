package db

import (
	"context"
	"fmt"
	"time"
)

// execTx executes the function fn within a database transaction.
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.connPool.Begin(ctx)
	if err != nil {
		return err
	}
	q := New(tx)
	err = fn(q)
	if err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("tx error: %v, rollback error: %v", err, rollbackErr)
		}
		return err
	}
	return tx.Commit(ctx)
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

type CreateInvoiceTxResult struct {
	Invoice
	LineItems []LineItem `json:"line_items"`
}

func (store *SQLStore) CreateInvoiceTx(ctx context.Context, arg CreateInvoiceTxParams) (CreateInvoiceTxResult, error) {
	var result CreateInvoiceTxResult
	err := store.execTx(
		ctx,
		func(q *Queries) error {

			invoice, err := q.InsertInvoice(
				ctx,
				InsertInvoiceParams{
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
