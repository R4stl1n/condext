package util

import (
	"fmt"
	"github.com/shopspring/decimal"
)

func PrintBanner() {
	fmt.Println(`   ______                __          __ 
  / ____/___  ____  ____/ /__  _  __/ /_
 / /   / __ \/ __ \/ __  / _ \| |/_/ __/
/ /___/ /_/ / / / / /_/ /  __/>  </ /_  
\____/\____/_/ /_/\__,_/\___/_/|_|\__/  
       -- By R4stl1n`)
	fmt.Println()
}

func GetPercentageDifference(oldNumber float64, newNumber float64) float64 {
	oldNumDec := decimal.NewFromFloat(oldNumber)
	newNumDec := decimal.NewFromFloat(newNumber)

	diff := newNumDec.Sub(oldNumDec)

	deltaPercentage := 0.0

	if oldNumber != 0 {
		deltaPercentage, _ = diff.Div(oldNumDec).Mul(decimal.NewFromFloat(100)).Float64()
	}

	return deltaPercentage
}

func GetPercentageDifferenceDecimal(oldNumber decimal.Decimal, newNumber decimal.Decimal) decimal.Decimal {
	diff := newNumber.Sub(oldNumber)
	if diff.IsZero() == false && oldNumber.IsZero() == false {
		return diff.Div(oldNumber).Mul(decimal.NewFromFloat(100))
	} else {
		return decimal.NewFromFloat(0.0)
	}
}

func GetPercentage(number float64, percent float64) float64 {
	numDec := decimal.NewFromFloat(number)
	perDec := decimal.NewFromFloat(percent)
	retVal, _ := numDec.Mul(perDec).Div(decimal.NewFromFloat(100)).Float64()
	return retVal
}

func GetDecimalPercentage(number decimal.Decimal, percent decimal.Decimal) decimal.Decimal {
	return number.Mul(percent).Div(decimal.NewFromFloat(100))
}