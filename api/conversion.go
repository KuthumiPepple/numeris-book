package api

import (
	"strconv"
)

func convertRateFromPercentToBasisPoints(rate string) int {
	rateFloat, _ := strconv.ParseFloat(rate, 64)
	basisPoints := int((rateFloat * 100))
	return basisPoints
}

func convertStringToFloat64(value string) float64 {
	floatValue, _ := strconv.ParseFloat(value, 64)
	return floatValue
}
