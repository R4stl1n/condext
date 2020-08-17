package dto

import "github.com/jinzhu/gorm"

type CondextConfigModel struct {
	gorm.Model

	Active             bool
	ReBalanceThreshold float64
	OrderTimeout       int64
	RebalanceFrequency int64
	StartingBalance    float64
	FloatingPercentage float64
}
