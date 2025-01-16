package util

import (
	"strconv"
)

func ConvertRateFromPercentToBasisPoints(rate string) int {
	rateFloat, _ := strconv.ParseFloat(rate, 64)
	basisPoints := int((rateFloat * 100))
	return basisPoints
}
