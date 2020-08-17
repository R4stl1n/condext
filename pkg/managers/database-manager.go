package managers

import (
	"errors"
	"github.com/satori/go.uuid"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/r4stl1n/condext/pkg/dto"
)

type DatabaseManager struct {
	gormClient *gorm.DB
}

func CreateDatabaseManager(databaseName string) (*DatabaseManager, error) {

	databaseClient, databaseClientError := gorm.Open("sqlite3", databaseName)

	if databaseClientError != nil {
		return &DatabaseManager{}, databaseClientError
	}

	databaseClient.AutoMigrate(&dto.CondextConfigModel{})
	databaseClient.AutoMigrate(&dto.IndexedSymbolModel{})

	return &DatabaseManager{
		gormClient: databaseClient,
	}, nil
}

func (databaseManager *DatabaseManager) CreateIndexSymbolModel(indexedSymbolModel dto.IndexedSymbolModel) (dto.IndexedSymbolModel, error) {

	if databaseManager.CheckIfSymbolIsIndexed(indexedSymbolModel.Symbol) != false {
		return dto.IndexedSymbolModel{}, errors.New("symbol is already indexed")
	}

	newUUID := uuid.NewV4().String()

	indexedSymbolModel.UUID = newUUID

	createError := databaseManager.gormClient.Create(&indexedSymbolModel).Error

	if createError != nil {
		return dto.IndexedSymbolModel{}, createError
	}

	return indexedSymbolModel, nil
}

func (databaseManager *DatabaseManager) CheckIfSymbolIsIndexed(symbol string) bool {

	indexedSymbolModel := dto.IndexedSymbolModel{}

	findError := databaseManager.gormClient.Find(&indexedSymbolModel, "symbol = ?", symbol).Error

	if findError != nil {
		return false
	}

	return true
}

func (databaseManager *DatabaseManager) GetAllIndexedSymbols() ([]dto.IndexedSymbolModel, error) {
	var indexedSymbolModels []dto.IndexedSymbolModel

	findError := databaseManager.gormClient.Find(&indexedSymbolModels).Error

	if findError != nil {
		return indexedSymbolModels, findError
	}

	return indexedSymbolModels, nil
}

func (databaseManager *DatabaseManager) GetIndexedSymbolByUUID(uuid string) (dto.IndexedSymbolModel, error) {

	indexedSymbolModel := dto.IndexedSymbolModel{}

	findError := databaseManager.gormClient.Find(&indexedSymbolModel, "uuid = ?", uuid).Error

	if findError != nil {
		return indexedSymbolModel, findError
	}

	return indexedSymbolModel, nil
}

func (databaseManager *DatabaseManager) GetIndexedSymbolBySymbol(symbol string) (dto.IndexedSymbolModel, error) {

	indexedSymbolModel := dto.IndexedSymbolModel{}

	findError := databaseManager.gormClient.Find(&indexedSymbolModel, "symbol = ?", symbol).Error

	if findError != nil {
		return indexedSymbolModel, findError
	}

	return indexedSymbolModel, nil
}

func (databaseManager *DatabaseManager) UpdateIndexedSymbolModel(updatedIndexedSymbolModel dto.IndexedSymbolModel) (dto.IndexedSymbolModel, error) {

	indexedSymbolModel := dto.IndexedSymbolModel{}

	findError := databaseManager.gormClient.Find(&indexedSymbolModel, "uuid = ?", updatedIndexedSymbolModel.UUID).Error

	if findError != nil {
		return indexedSymbolModel, findError
	}

	indexedSymbolModel.DesiredPercentage = updatedIndexedSymbolModel.DesiredPercentage
	indexedSymbolModel.CurrentPercentage = updatedIndexedSymbolModel.CurrentPercentage
	indexedSymbolModel.Locked = updatedIndexedSymbolModel.Locked
	indexedSymbolModel.CurrentPrice = updatedIndexedSymbolModel.CurrentPrice
	indexedSymbolModel.AmountFromTarget = updatedIndexedSymbolModel.AmountFromTarget
	indexedSymbolModel.Amount = updatedIndexedSymbolModel.Amount
	indexedSymbolModel.LastOrderId = updatedIndexedSymbolModel.LastOrderId

	databaseManager.gormClient.Save(&indexedSymbolModel)

	return indexedSymbolModel, nil
}

func (databaseManager *DatabaseManager) CreateCondextConfigAndFirstSymbolModel() error {

	_, configModelError := databaseManager.GetCondextConfigModel()

	if configModelError != nil {
		condextConfigModel := dto.CondextConfigModel{}

		condextConfigModel.Active = false
		condextConfigModel.BalanceThreshold = 1
		condextConfigModel.OrderTimeout = 10
		condextConfigModel.RebalanceFrequency = 60
		condextConfigModel.StartingBalance = 100000

		createError := databaseManager.gormClient.Create(&condextConfigModel).Error

		if createError != nil {
			return createError
		}

		_, indexSymbolCreateError := databaseManager.CreateIndexSymbolModel(dto.IndexedSymbolModel{
			Symbol:            "USD",
			DesiredPercentage: 100,
		})

		if indexSymbolCreateError != nil {
			return indexSymbolCreateError
		}
	}

	return nil
}

func (databaseManager *DatabaseManager) GetCondextConfigModel() (dto.CondextConfigModel, error) {

	condextConfigModel := dto.CondextConfigModel{}

	findError := databaseManager.gormClient.Last(&condextConfigModel).Error

	if findError != nil {
		return condextConfigModel, findError
	}

	return condextConfigModel, nil
}

func (databaseManager *DatabaseManager) UpdateCondextConfig(updatedConfigModel dto.CondextConfigModel) (dto.CondextConfigModel, error) {

	configModel, configModelError := databaseManager.GetCondextConfigModel()

	if configModelError != nil {
		return dto.CondextConfigModel{}, configModelError
	}

	configModel.Active = updatedConfigModel.Active
	configModel.BalanceThreshold = updatedConfigModel.BalanceThreshold
	configModel.OrderTimeout = updatedConfigModel.OrderTimeout
	configModel.RebalanceFrequency = updatedConfigModel.RebalanceFrequency
	configModel.StartingBalance = updatedConfigModel.StartingBalance

	databaseManager.gormClient.Save(&configModel)

	return configModel, nil
}
