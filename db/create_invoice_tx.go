package db

import (
	"context"
	"fmt"

	"github.com/Rhymond/go-money"
	"github.com/kuthumipepple/numeris-book/util"
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
	CustomerName          string      `json:"customer_name"`
	CustomerEmail         string      `json:"customer_email"`
	CustomerPhone         string      `json:"customer_phone"`
	CustomerAddress       string      `json:"customer_address"`
	SenderName            string      `json:"sender_name"`
	SenderEmail           string      `json:"sender_email"`
	SenderPhone           string      `json:"sender_phone"`
	SenderAddress         string      `json:"sender_address"`
	IssueDate             string      `json:"issue_date"`
	DueDate               string      `json:"due_date"`
	Status                string      `json:"status"`
	DiscountRateInPercent string      `json:"discount_rate_in_percent"`
	PaymentInfo           string      `json:"payment_info"`
	Items                 []ItemsData `json:"items"`
}

type ItemsData struct {
	Description string  `json:"description"`
	Quantity    int64   `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
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
			lineItemsArgs := make([]InsertLineItemParams, len(arg.Items))
			subtotal := money.New(0, money.USD)

			for i, item := range arg.Items {
				unitPrice := money.NewFromFloat(item.UnitPrice, money.USD)
				totalPrice := unitPrice.Multiply(item.Quantity)
				lineItemsArgs[i] = InsertLineItemParams{
					Description: item.Description,
					Quantity:    item.Quantity,
					UnitPrice:   unitPrice.Amount(),
					TotalPrice:  totalPrice.Amount(),
				}
				subtotal, _ = subtotal.Add(totalPrice)
			}

			rateInBasisPoints := util.ConvertRateFromPercentToBasisPoints(arg.DiscountRateInPercent)
			parts, _ := subtotal.Allocate(rateInBasisPoints, 10000-rateInBasisPoints)
			discount, totalAmount := parts[0], parts[1]

			invoice, err := q.InsertInvoice(
				ctx,
				InsertInvoiceParams{
					CustomerName:          arg.CustomerName,
					CustomerEmail:         arg.CustomerEmail,
					CustomerPhone:         arg.CustomerPhone,
					CustomerAddress:       arg.CustomerAddress,
					SenderName:            arg.SenderName,
					SenderEmail:           arg.SenderEmail,
					SenderPhone:           arg.SenderPhone,
					SenderAddress:         arg.SenderAddress,
					IssueDate:             arg.IssueDate,
					DueDate:               arg.DueDate,
					Status:                arg.Status,
					Subtotal:              subtotal.Amount(),
					DiscountRateInPercent: arg.DiscountRateInPercent,
					Discount:              discount.Amount(),
					TotalAmount:           totalAmount.Amount(),
					PaymentInfo:           arg.PaymentInfo,
				},
			)
			if err != nil {
				return err
			}

			result.Invoice = invoice

			for _, item := range lineItemsArgs {
				lineItem, err := q.InsertLineItem(
					ctx,
					InsertLineItemParams{
						InvoiceNumber: result.InvoiceNumber,
						Description:   item.Description,
						Quantity:      item.Quantity,
						UnitPrice:     item.UnitPrice,
						TotalPrice:    item.TotalPrice,
					},
				)
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
