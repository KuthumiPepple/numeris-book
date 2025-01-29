package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kuthumipepple/numeris-book/db"
	mockdb "github.com/kuthumipepple/numeris-book/db/mock"
	"github.com/kuthumipepple/numeris-book/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateInvoiceAPI(t *testing.T) {

	fixedTime := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"customer_name":    "john doe",
				"customer_email":   "jdoe@fakemail.com",
				"customer_phone":   "+1234567890",
				"customer_address": "123 A Street",
				"sender_name":      "acme inc",
				"sender_email":     "xyz@acme.com",
				"sender_phone":     "+9876543210",
				"sender_address":   "456 X Street",
				"issue_date":       fixedTime.Format(time.DateOnly),
				"due_date":         fixedTime.AddDate(0, 0, 1).Format(time.DateOnly),
				"status":           "pending_payment",
				"discount_rate":    "5.80",
				"payment_info":     "Bank transfer",
				"line_items": []gin.H{
					{
						"description": "item 1",
						"quantity":    1,
						"unit_price":  "100.00",
					},
					{
						"description": "item 2",
						"quantity":    2,
						"unit_price":  "58.99",
					},
				},
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateInvoiceTxParams{
					CustomerName:    "john doe",
					CustomerEmail:   "jdoe@fakemail.com",
					CustomerPhone:   "+1234567890",
					CustomerAddress: "123 A Street",
					SenderName:      "acme inc",
					SenderEmail:     "xyz@acme.com",
					SenderPhone:     "+9876543210",
					SenderAddress:   "456 X Street",
					IssueDate:       fixedTime,
					DueDate:         fixedTime.AddDate(0, 0, 1),
					Status:          "pending_payment",
					Subtotal:        int64(21798),
					DiscountRate:    int64(580),
					Discount:        int64(1265),
					TotalAmount:     int64(20533),
					PaymentInfo:     "Bank transfer",
					Items: []db.InsertLineItemParams{
						{
							Description: "item 1",
							Quantity:    int64(1),
							UnitPrice:   int64(10000),
							TotalPrice:  int64(10000),
						},
						{
							Description: "item 2",
							Quantity:    int64(2),
							UnitPrice:   int64(5899),
							TotalPrice:  int64(11798),
						},
					},
				}
				result := db.InvoiceResult{
					Invoice: db.Invoice{
						InvoiceNumber: int64(1),
						Subtotal:      int64(21798),
						CreatedAt:     fixedTime,
					},
				}

				store.EXPECT().
					CreateInvoiceTx(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(result, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				requireBodyMatchResponse(
					t,
					recorder.Body,
					createInvoiceResponse{1, fixedTime})
			},
		},

		{
			name: "DueDateNotLaterThanIssueDate",
			body: gin.H{
				"customer_name":            "john doe",
				"customer_email":           "jdoe@fakemail.com",
				"customer_phone":           "+1234567890",
				"customer_address":         "123 A Street",
				"sender_name":              "acme inc",
				"sender_email":             "xyz@acme.com",
				"sender_phone":             "+9876543210",
				"sender_address":           "456 X Street",
				"issue_date":               fixedTime.Format(time.DateOnly),
				"due_date":                 fixedTime.AddDate(0, 0, -1).Format(time.DateOnly),
				"status":                   "pending_payment",
				"discount_rate_in_percent": "5.8",
				"payment_info":             "Bank transfer",
				"line_items": []gin.H{
					{
						"description": "item 1",
						"quantity":    1,
						"unit_price":  "100.00",
					},
					{
						"description": "item 2",
						"quantity":    2,
						"unit_price":  "58.99",
					},
				},
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					CreateInvoiceTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},

		{
			name: "IncompleteRequestData",
			body: gin.H{
				"customer_name": "john doe",
				"payment_info":  "Bank transfer",
				"line_items": []gin.H{
					{
						"description": "item 1",
						"quantity":    1,
					},
					{
						"description": "item 2",
						"quantity":    2,
						"unit_price":  "58.99",
					},
				},
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					CreateInvoiceTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},

		{
			name: "InvalidFormatForDueDate",
			body: gin.H{
				"customer_name":    "john doe",
				"customer_email":   "jdoe@fakemail.com",
				"customer_phone":   "+1234567890",
				"customer_address": "123 A Street",
				"sender_name":      "acme inc",
				"sender_email":     "xyz@acme.com",
				"sender_phone":     "+9876543210",
				"sender_address":   "456 X Street",
				"issue_date":       fixedTime.Format(time.DateOnly),
				"due_date":         "22/01/2025",
				"status":           "pending_payment",
				"discount_rate":    "5.80",
				"payment_info":     "Bank transfer",
				"line_items": []gin.H{
					{
						"description": "item 1",
						"quantity":    1,
						"unit_price":  "100.00",
					},
					{
						"description": "item 2",
						"quantity":    2,
						"unit_price":  "58.99",
					},
				},
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					CreateInvoiceTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},

		{
			name: "DiscountRateLessThanZero",
			body: gin.H{
				"customer_name":    "john doe",
				"customer_email":   "jdoe@fakemail.com",
				"customer_phone":   "+1234567890",
				"customer_address": "123 A Street",
				"sender_name":      "acme inc",
				"sender_email":     "xyz@acme.com",
				"sender_phone":     "+9876543210",
				"sender_address":   "456 X Street",
				"issue_date":       fixedTime.Format(time.DateOnly),
				"due_date":         fixedTime.AddDate(0, 0, 1).Format(time.DateOnly),
				"status":           "pending_payment",
				"discount_rate":    "-1",
				"payment_info":     "Bank transfer",
				"line_items": []gin.H{
					{
						"description": "item 1",
						"quantity":    1,
						"unit_price":  "100.00",
					},
					{
						"description": "item 2",
						"quantity":    2,
						"unit_price":  "58.99",
					},
				},
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					CreateInvoiceTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},

		{
			name: "DiscountRateTooHigh",
			body: gin.H{
				"customer_name":    "john doe",
				"customer_email":   "jdoe@fakemail.com",
				"customer_phone":   "+1234567890",
				"customer_address": "123 A Street",
				"sender_name":      "acme inc",
				"sender_email":     "xyz@acme.com",
				"sender_phone":     "+9876543210",
				"sender_address":   "456 X Street",
				"issue_date":       fixedTime.Format(time.DateOnly),
				"due_date":         fixedTime.AddDate(0, 0, 1).Format(time.DateOnly),
				"status":           "pending_payment",
				"discount_rate":    "100",
				"payment_info":     "Bank transfer",
				"line_items": []gin.H{
					{
						"description": "item 1",
						"quantity":    1,
						"unit_price":  "100.00",
					},
					{
						"description": "item 2",
						"quantity":    2,
						"unit_price":  "58.99",
					},
				},
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					CreateInvoiceTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},

		{
			name: "UnitPriceIsNegative",
			body: gin.H{
				"customer_name":    "john doe",
				"customer_email":   "jdoe@fakemail.com",
				"customer_phone":   "+1234567890",
				"customer_address": "123 A Street",
				"sender_name":      "acme inc",
				"sender_email":     "xyz@acme.com",
				"sender_phone":     "+9876543210",
				"sender_address":   "456 X Street",
				"issue_date":       fixedTime.Format(time.DateOnly),
				"due_date":         fixedTime.AddDate(0, 0, 1).Format(time.DateOnly),
				"status":           "pending_payment",
				"discount_rate":    "5.80",
				"payment_info":     "Bank transfer",
				"line_items": []gin.H{
					{
						"description": "item 1",
						"quantity":    1,
						"unit_price":  "-100.00",
					},
					{
						"description": "item 2",
						"quantity":    2,
						"unit_price":  "58.99",
					},
				},
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					CreateInvoiceTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},

		{
			name: "UnitPriceHasMoreThanTwoDecimalPlaces",
			body: gin.H{
				"customer_name":    "john doe",
				"customer_email":   "jdoe@fakemail.com",
				"customer_phone":   "+1234567890",
				"customer_address": "123 A Street",
				"sender_name":      "acme inc",
				"sender_email":     "xyz@acme.com",
				"sender_phone":     "+9876543210",
				"sender_address":   "456 X Street",
				"issue_date":       fixedTime.Format(time.DateOnly),
				"due_date":         fixedTime.AddDate(0, 0, 1).Format(time.DateOnly),
				"status":           "pending_payment",
				"discount_rate":    "5.80",
				"payment_info":     "Bank transfer",
				"line_items": []gin.H{
					{
						"description": "item 1",
						"quantity":    1,
						"unit_price":  "100.00",
					},
					{
						"description": "item 2",
						"quantity":    2,
						"unit_price":  "58.999",
					},
				},
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					CreateInvoiceTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},

		{
			name: "QuantityIsNegative",
			body: gin.H{
				"customer_name":    "john doe",
				"customer_email":   "jdoe@fakemail.com",
				"customer_phone":   "+1234567890",
				"customer_address": "123 A Street",
				"sender_name":      "acme inc",
				"sender_email":     "xyz@acme.com",
				"sender_phone":     "+9876543210",
				"sender_address":   "456 X Street",
				"issue_date":       fixedTime.Format(time.DateOnly),
				"due_date":         fixedTime.AddDate(0, 0, 1).Format(time.DateOnly),
				"status":           "pending_payment",
				"discount_rate":    "5.80",
				"payment_info":     "Bank transfer",
				"line_items": []gin.H{
					{
						"description": "item 1",
						"quantity":    -1,
						"unit_price":  "100.00",
					},
					{
						"description": "item 2",
						"quantity":    2,
						"unit_price":  "58.99",
					},
				},
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					CreateInvoiceTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},

		{
			name: "InternalError",
			body: gin.H{
				"customer_name":    "john doe",
				"customer_email":   "jdoe@fakemail.com",
				"customer_phone":   "+1234567890",
				"customer_address": "123 A Street",
				"sender_name":      "acme inc",
				"sender_email":     "xyz@acme.com",
				"sender_phone":     "+9876543210",
				"sender_address":   "456 X Street",
				"issue_date":       fixedTime.Format(time.DateOnly),
				"due_date":         fixedTime.AddDate(0, 0, 1).Format(time.DateOnly),
				"status":           "pending_payment",
				"discount_rate":    "5.80",
				"payment_info":     "Bank transfer",
				"line_items": []gin.H{
					{
						"description": "item 1",
						"quantity":    1,
						"unit_price":  "100.00",
					},
					{
						"description": "item 2",
						"quantity":    2,
						"unit_price":  "58.99",
					},
				},
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					CreateInvoiceTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.InvoiceResult{}, &pgconn.PgError{})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			store := mockdb.NewMockStore(ctrl)

			tc.buildStubs(store)

			url := "/invoices"
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			server := NewServer(store)

			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(recorder)
		})
	}
}

func requireBodyMatchResponse(t *testing.T, body *bytes.Buffer, response createInvoiceResponse) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotResponse createInvoiceResponse
	err = json.Unmarshal(data, &gotResponse)

	require.NoError(t, err)
	require.Equal(t, response, gotResponse)
}

func TestGetInvoiceAPI(t *testing.T) {
	fakeID := util.RandomInt(1, 1000)
	fixedTime := time.Now().UTC()

	testCases := []struct {
		name          string
		invoiceNumber int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:          "OK",
			invoiceNumber: fakeID,
			buildStubs: func(store *mockdb.MockStore) {
				result := db.InvoiceResult{
					Invoice: db.Invoice{
						InvoiceNumber:   fakeID,
						IssueDate:       fixedTime,
						DueDate:         fixedTime.AddDate(0, 0, 1),
						Subtotal:        int64(1234567890),
						DiscountRate:    int64(1234),
						Discount:        int64(123456),
						TotalAmount:     int64(123456789),
						BillingCurrency: "USD",
						Note:            "Thank you for your patronage",
						CreatedAt:       fixedTime.Add(2 * time.Hour),
					},
					LineItems: []db.LineItem{
						{
							ID:            fakeID + 1,
							InvoiceNumber: fakeID,
							Description:   "item 1",
							Quantity:      int64(1),
							UnitPrice:     int64(12345),
							TotalPrice:    int64(1234567),
						},
						{
							ID:            fakeID + 2,
							InvoiceNumber: fakeID,
							Description:   "item 2",
							Quantity:      int64(12),
							UnitPrice:     int64(123),
							TotalPrice:    int64(12345),
						},
					},
				}
				store.EXPECT().
					GetInvoice(gomock.Any(), gomock.Eq(fakeID)).
					Times(1).
					Return(result, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchGetResponse(
					t,
					recorder.Body,
					getInvoiceResponse{
						InvoiceNumber:   fakeID,
						IssueDate:       fixedTime.Format(time.DateOnly),
						DueDate:         fixedTime.AddDate(0, 0, 1).Format(time.DateOnly),
						Subtotal:        "$12,345,678.90",
						DiscountRate:    "12.34%",
						Discount:        "$1,234.56",
						TotalAmount:     "$1,234,567.89",
						BillingCurrency: "USD",
						Note:            "Thank you for your patronage",
						CreatedAt:       fixedTime.Add(2 * time.Hour).Format(time.RFC3339),
						Items: []getInvoiceResponseItem{
							{fakeID + 1, fakeID, "item 1", 1, "$123.45", "$12,345.67"},
							{fakeID + 2, fakeID, "item 2", 12, "$1.23", "$123.45"},
						},
					},
				)
			},
		},

		{
			name:          "InvalidID",
			invoiceNumber: -1,
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetInvoice(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			store := mockdb.NewMockStore(ctrl)

			tc.buildStubs(store)

			url := fmt.Sprintf("/invoices/%d", tc.invoiceNumber)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			server := NewServer(store)

			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(recorder)
		})
	}
}

func requireBodyMatchGetResponse(t *testing.T, body *bytes.Buffer, response getInvoiceResponse) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotResponse getInvoiceResponse
	err = json.Unmarshal(data, &gotResponse)

	require.NoError(t, err)
	require.Equal(t, response, gotResponse)
}
