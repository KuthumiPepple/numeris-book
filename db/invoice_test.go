package db

import (
	"context"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/kuthumipepple/numeris-book/util"
	"github.com/stretchr/testify/require"
)

func insertRandomInvoice(t *testing.T) Invoice {
	arg := InsertInvoiceParams{
		CustomerName:    util.RandomName(),
		CustomerEmail:   util.RandomEmail(),
		CustomerPhone:   util.RandomPhone(),
		CustomerAddress: util.RandomAddress(),
		SenderName:      util.RandomName(),
		SenderEmail:     util.RandomEmail(),
		SenderPhone:     util.RandomPhone(),
		SenderAddress:   util.RandomAddress(),
		IssueDate:       time.Now(),
		DueDate:         time.Now().AddDate(0, 0, 30),
		Status:          util.RandomStatus(),
		Subtotal:        util.RandomInt(100, 10000),
		DiscountRate:    util.RandomInt(0, 10000),
		Discount:        util.RandomInt(0, 10000),
		TotalAmount:     util.RandomInt(100, 10000),
		PaymentInfo:     util.RandomString(10),
	}

	invoice, err := testStore.InsertInvoice(context.Background(), arg)
	require.NoError(t, err)

	require.NotZero(t, invoice.InvoiceNumber)

	require.Equal(t, arg.CustomerName, invoice.CustomerName)
	require.Equal(t, arg.CustomerEmail, invoice.CustomerEmail)
	require.Equal(t, arg.CustomerPhone, invoice.CustomerPhone)
	require.Equal(t, arg.CustomerAddress, invoice.CustomerAddress)
	require.Equal(t, arg.SenderName, invoice.SenderName)
	require.Equal(t, arg.SenderEmail, invoice.SenderEmail)
	require.Equal(t, arg.SenderPhone, invoice.SenderPhone)
	require.Equal(t, arg.SenderAddress, invoice.SenderAddress)
	require.WithinDuration(t, arg.IssueDate, invoice.IssueDate, time.Second)
	require.WithinDuration(t, arg.DueDate, invoice.DueDate, time.Second)
	require.Equal(t, arg.Status, invoice.Status)
	require.Equal(t, arg.Subtotal, invoice.Subtotal)
	require.Equal(t, arg.DiscountRate, invoice.DiscountRate)
	require.Equal(t, arg.Discount, invoice.Discount)
	require.Equal(t, arg.TotalAmount, invoice.TotalAmount)
	require.Equal(t, money.USD, invoice.BillingCurrency)
	require.Equal(t, arg.PaymentInfo, invoice.PaymentInfo)

	require.NotEmpty(t, invoice.Note)
	require.NotZero(t, invoice.CreatedAt)

	return invoice
}

func TestInsertInvoice(t *testing.T) {
	insertRandomInvoice(t)
}

func TestInsertLineItem(t *testing.T) {
	invoice := insertRandomInvoice(t)

	arg := InsertLineItemParams{
		InvoiceNumber: invoice.InvoiceNumber,
		Description:   util.RandomString(10),
		Quantity:      util.RandomInt(1, 100),
		UnitPrice:     util.RandomInt(100, 1000),
		TotalPrice:    util.RandomInt(100, 1000),
	}

	lineItem, err := testStore.InsertLineItem(context.Background(), arg)
	require.NoError(t, err)

	require.NotZero(t, lineItem.ID)
	require.Equal(t, arg.InvoiceNumber, lineItem.InvoiceNumber)
	require.Equal(t, arg.Description, lineItem.Description)
	require.Equal(t, arg.Quantity, lineItem.Quantity)
	require.Equal(t, arg.UnitPrice, lineItem.UnitPrice)
	require.Equal(t, arg.TotalPrice, lineItem.TotalPrice)
}

func TestCreateInvoiceTx(t *testing.T) {
	n := 5
	testItems := make([]InsertLineItemParams, n)
	for i := 0; i < n; i++ {
		testItems[i] = InsertLineItemParams{
			Description: util.RandomString(10),
			Quantity:    util.RandomInt(1, 100),
			UnitPrice:   util.RandomInt(100, 1000),
			TotalPrice:  util.RandomInt(100, 1000),
		}
	}
	arg := CreateInvoiceTxParams{
		CustomerName:    util.RandomName(),
		CustomerEmail:   util.RandomEmail(),
		CustomerPhone:   util.RandomPhone(),
		CustomerAddress: util.RandomAddress(),
		SenderName:      util.RandomName(),
		SenderEmail:     util.RandomEmail(),
		SenderPhone:     util.RandomPhone(),
		SenderAddress:   util.RandomAddress(),
		IssueDate:       time.Now(),
		DueDate:         time.Now().AddDate(0, 0, 30),
		Status:          util.RandomStatus(),
		Subtotal:        util.RandomInt(100, 10000),
		DiscountRate:    util.RandomInt(0, 10000),
		Discount:        util.RandomInt(0, 10000),
		TotalAmount:     util.RandomInt(100, 10000),
		PaymentInfo:     util.RandomString(10),
		Items:           testItems,
	}

	result, err := testStore.CreateInvoiceTx(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	// Check invoice
	invoice := result.Invoice
	require.NotZero(t, invoice.InvoiceNumber)
	require.Equal(t, arg.CustomerName, invoice.CustomerName)
	require.Equal(t, arg.CustomerEmail, invoice.CustomerEmail)
	require.Equal(t, arg.CustomerPhone, invoice.CustomerPhone)
	require.Equal(t, arg.CustomerAddress, invoice.CustomerAddress)
	require.Equal(t, arg.SenderName, invoice.SenderName)
	require.Equal(t, arg.SenderEmail, invoice.SenderEmail)
	require.Equal(t, arg.SenderPhone, invoice.SenderPhone)
	require.Equal(t, arg.SenderAddress, invoice.SenderAddress)
	require.WithinDuration(t, arg.IssueDate, invoice.IssueDate, time.Second)
	require.WithinDuration(t, arg.DueDate, invoice.DueDate, time.Second)
	require.Equal(t, arg.Status, invoice.Status)
	require.Equal(t, arg.Subtotal, invoice.Subtotal)
	require.Equal(t, arg.DiscountRate, invoice.DiscountRate)
	require.Equal(t, arg.Discount, invoice.Discount)
	require.Equal(t, arg.TotalAmount, invoice.TotalAmount)
	require.Equal(t, money.USD, invoice.BillingCurrency)
	require.Equal(t, arg.PaymentInfo, invoice.PaymentInfo)
	require.NotEmpty(t, invoice.Note)
	require.NotZero(t, invoice.CreatedAt)

	// check line items
	require.Len(t, result.LineItems, n)
	for i, lineItem := range result.LineItems {
		require.NotZero(t, lineItem.ID)
		require.Equal(t, invoice.InvoiceNumber, lineItem.InvoiceNumber)
		require.Equal(t, testItems[i].Description, lineItem.Description)
		require.Equal(t, testItems[i].Quantity, lineItem.Quantity)
		require.Equal(t, testItems[i].UnitPrice, lineItem.UnitPrice)
		require.Equal(t, testItems[i].TotalPrice, lineItem.TotalPrice)
	}

}
