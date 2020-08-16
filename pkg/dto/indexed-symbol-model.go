package dto

import "github.com/jinzhu/gorm"

type IndexedSymbolModel struct {
	gorm.Model

	UUID              string
	Symbol            string
	Locked            bool
	DesiredPercentage float64
	CurrentPercentage float64
	CurrentPrice      float64
	Amount            int64
	AmountFromTarget  int64
	LastOrderId       string
}
