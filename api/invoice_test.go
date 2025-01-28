package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kuthumipepple/numeris-book/db"
	mockdb "github.com/kuthumipepple/numeris-book/db/mock"
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
