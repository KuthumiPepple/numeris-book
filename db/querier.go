package db

import (
	"context"
)

type Querier interface {
	InsertInvoiceRecord(ctx context.Context, arg InsertInvoiceRecordParams) (Invoice, error)
	InsertLineItem(ctx context.Context, arg InsertLineItemParams) (LineItem, error)
}

var _ Querier = (*Queries)(nil)
