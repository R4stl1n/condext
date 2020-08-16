package dto

import "github.com/jinzhu/gorm"

type CondextConfigModel struct {
	gorm.Model

	Active               bool
	BalanceThreshold     float64
	OrderTimeout         int64
	ReBalanceTickSetting float64
}
