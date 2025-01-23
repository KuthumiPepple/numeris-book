package util

// all valid invoice statuses
const (
	DRAFT           = "draft"
	PENDING_PAYMENT = "pending_payment"
	OVERDUE         = "overdue"
	PAID            = "paid"
)

// Contains checks if a slice of strings contains a specific string element.
func Contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
