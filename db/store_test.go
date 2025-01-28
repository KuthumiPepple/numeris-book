package db

import (
	"context"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/kuthumipepple/numeris-book/util"
	"github.com/stretchr/testify/require"
)

func TestCreateInvoiceTx(t *testing.T) {
	createRandomInvoiceTx(t)
}

func createRandomInvoiceTx(t *testing.T) InvoiceResult {
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

	return result
}

func TestGetInvoice(t *testing.T) {
	result1 := createRandomInvoiceTx(t)
	result2, err := testStore.GetInvoice(context.Background(), result1.InvoiceNumber)
	require.NoError(t, err)
	require.NotEmpty(t, result2)

	// check that both invoices are the same
	require.Equal(t, result1.InvoiceNumber, result2.InvoiceNumber)
	require.Equal(t, result1.CustomerName, result2.CustomerName)
	require.Equal(t, result1.CustomerEmail, result2.CustomerEmail)
	require.Equal(t, result1.CustomerPhone, result2.CustomerPhone)
	require.Equal(t, result1.CustomerAddress, result2.CustomerAddress)
	require.Equal(t, result1.SenderName, result2.SenderName)
	require.Equal(t, result1.SenderEmail, result2.SenderEmail)
	require.Equal(t, result1.SenderPhone, result2.SenderPhone)
	require.Equal(t, result1.SenderAddress, result2.SenderAddress)
	require.WithinDuration(t, result1.IssueDate, result2.IssueDate, time.Second)
	require.WithinDuration(t, result1.DueDate, result2.DueDate, time.Second)
	require.Equal(t, result1.Status, result2.Status)
	require.Equal(t, result1.Subtotal, result2.Subtotal)
	require.Equal(t, result1.DiscountRate, result2.DiscountRate)
	require.Equal(t, result1.Discount, result2.Discount)
	require.Equal(t, result1.TotalAmount, result2.TotalAmount)
	require.Equal(t, result1.BillingCurrency, result2.BillingCurrency)
	require.Equal(t, result1.PaymentInfo, result2.PaymentInfo)
	require.Equal(t, result1.Note, result2.Note)
	require.WithinDuration(t, result1.CreatedAt, result2.CreatedAt, time.Second)

	require.Equal(t, result1.LineItems, result2.LineItems)
}
