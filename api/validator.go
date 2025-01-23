package api

import (
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kuthumipepple/numeris-book/util"
)

var ratePattern = regexp.MustCompile(`^(?:[0-9]|[1-9][0-9])(?:\.[0-9]{1,})?$`)
var pricePattern = regexp.MustCompile(`^\d+(?:\.\d{1,2})?$`)

var createInvoiceRequestValidation validator.StructLevelFunc = func(sl validator.StructLevel) {
	req := sl.Current().Interface().(createInvoiceRequest)

	// Validate Status
	validStatuses := []string{util.DRAFT, util.PENDING_PAYMENT, util.OVERDUE}
	if !util.Contains(validStatuses, req.Status) {
		sl.ReportError(req.Status, "Status", "status", "valid_status", "")
	}

	// Validate DiscountRate
	if !ratePattern.MatchString(req.DiscountRate) {
		sl.ReportError(req.DiscountRate, "DiscountRate", "discount_rate", "rate_>=_0_AND_rate_<_100", "")
	}

	// Validate DueDate comes after IssueDate
	issueDate, err := time.Parse("2006-01-02", req.IssueDate)
	if err != nil {
		sl.ReportError(req.IssueDate, "IssueDate", "issue_date", "date_format_is_YYYY-MM-DD", "2006-01-02")
		return
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		sl.ReportError(req.DueDate, "DueDate", "due_date", "date_format_is_YYYY-MM-DD", "2006-01-02")
		return
	}

	if !dueDate.After(issueDate) {
		sl.ReportError(req.DueDate, "DueDate", "due_date", "duedate_is_later_than_issuedate", "IssueDate")
	}

	// Validate UnitPrice is not a negative number
	for _, item := range req.LineItems {
		if !pricePattern.MatchString(item.UnitPrice) {
			sl.ReportError(item.UnitPrice, "UnitPrice", "unit_price", "unitprice_is_positive_AND_unitprice_has_not_more_than_two_decimal_places", "")
		}
	}
}
