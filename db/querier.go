package db

import (
	"context"
)

type Querier interface {
	InsertInvoice(ctx context.Context, arg InsertInvoiceParams) (Invoice, error)
	InsertLineItem(ctx context.Context, arg InsertLineItemParams) (LineItem, error)
}

var _ Querier = (*Queries)(nil)
