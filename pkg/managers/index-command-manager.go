package managers

import (
	"github.com/r4stl1n/condext/pkg/broker-integrations"
	"github.com/r4stl1n/condext/pkg/dto"
	"github.com/r4stl1n/condext/pkg/util"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gopkg.in/abiosoft/ishell.v2"
	"strconv"
	"strings"
)

type IndexCommandManager struct {
	databaseMgr  *DatabaseManager
	rebalanceMgr *RebalanceManager

	brokerIntegration *broker_integrations.BrokerIntegrationInterface
}

func CreateIndexCommandManager(databaseManager *DatabaseManager, rebalanceManager *RebalanceManager, selectedBrokerIntegration broker_integrations.BrokerIntegrationInterface) *IndexCommandManager {

	return &IndexCommandManager{
		databaseMgr:       databaseManager,
		rebalanceMgr:      rebalanceManager,
		brokerIntegration: &selectedBrokerIntegration,
	}
}

func (indexCommandManager *IndexCommandManager) AddSymbolToIndexCommand(c *ishell.Context) {

	// Grab the arguments from the command and validate them

	// Check that we have enough arguments
	if len(c.Args) != 3 {
		logrus.Warn("Not enough parameters submitted")
		return
	}

	symbolToAdd := strings.ToUpper(c.Args[0])
	symbolPercentage, symbolPercentageError := decimal.NewFromString(c.Args[1])
	symbolLocked, symbolLockedError := strconv.ParseBool(c.Args[2])
	symbolPercentageConverted, _ := symbolPercentage.Round(2).Float64()

	if symbolPercentageError != nil {
		logrus.Error(symbolPercentageError.Error())
		return
	}

	if symbolLockedError != nil {
		logrus.Error(symbolLockedError.Error())
		return
	}

	// Round the percentage to 2 points
	symbolPercentage = symbolPercentage.Round(3)

	// Check to make sure we are not already indexing the symbol
	indexSymbolExist := indexCommandManager.databaseMgr.CheckIfSymbolIsIndexed(symbolToAdd)

	if indexSymbolExist == true {
		logrus.Error("Requested symbol is already indexed")
		return
	}

	// Now validate if the symbol exist
	symbolExist, symbolExistError := (*indexCommandManager.brokerIntegration).CheckIfSymbolIsValid(symbolToAdd)

	if symbolExistError != nil {
		logrus.Error(symbolExistError.Error())
		return
	}

	if symbolExist != true {
		logrus.Error("Symbol does not exist or is not tradeable on broker")
		return
	}

	// Now we need to get the quote price
	symbolQuotePrice, symbolQuotePriceError := (*indexCommandManager.brokerIntegration).GetSymbolQuotePrice(symbolToAdd)

	if symbolQuotePriceError != nil {
		logrus.Error(symbolQuotePriceError.Error())
		return
	}

	// First we need to grab all the current indexed symbols
	indexedSymbols, indexedSymbolsError := indexCommandManager.databaseMgr.GetAllIndexedSymbols()

	if indexedSymbolsError != nil {
		logrus.Error(indexedSymbolsError)
		return
	}

	// Next we need to calculate the total locked and unlocked percentages available
	totalPercentageUnlocked := decimal.NewFromFloat(0.0)
	totalPercentageLocked := decimal.NewFromFloat(0.0)
	totalUnlockedSymbolsCount := decimal.NewFromInt(0)

	for _, indexedSymbol := range indexedSymbols {
		if indexedSymbol.Locked == false {
			totalPercentageUnlocked = totalPercentageUnlocked.Add(decimal.NewFromFloat(indexedSymbol.DesiredPercentage))
			totalUnlockedSymbolsCount = totalUnlockedSymbolsCount.Add(decimal.NewFromInt(1))
		} else {
			totalPercentageLocked = totalPercentageLocked.Add(decimal.NewFromFloat(indexedSymbol.DesiredPercentage))
		}
	}

	totalFreePercentage := (decimal.NewFromFloat(100.0).Sub(totalPercentageLocked)).Sub(totalPercentageUnlocked).Round(2)

	// Check if we have enough free percentage
	if totalFreePercentage.GreaterThanOrEqual(symbolPercentage) {
		// We have enough total free we can go ahead and move forward
		_, createIndexSymbolError := indexCommandManager.databaseMgr.CreateIndexSymbolModel(dto.IndexedSymbolModel{
			Symbol:            symbolToAdd,
			Locked:            symbolLocked,
			DesiredPercentage: symbolPercentageConverted,
		})

		if createIndexSymbolError != nil {
			logrus.Error(createIndexSymbolError.Error())
			return
		}

		logrus.Info("Symbol " + symbolToAdd + " added to index")
		return
	}

	// Now validate we have enough unlocked percentage to add
	if totalPercentageUnlocked.LessThan(symbolPercentage) {
		logrus.Error("Requested percentage is more than available unlocked percentage")
		return
	}

	// Remove the percentage from the other coins

	percentageToRemove := symbolPercentage.Div(totalUnlockedSymbolsCount).Round(2)
	_ = symbolLocked

	for _, indexedSymbol := range indexedSymbols {

		if indexedSymbol.Locked == false {

			indexedSymbol.DesiredPercentage, _ = decimal.NewFromFloat(indexedSymbol.DesiredPercentage).Sub(percentageToRemove).Round(2).Float64()

			_, updateError := indexCommandManager.databaseMgr.UpdateIndexedSymbolModel(indexedSymbol)

			if updateError != nil {
				logrus.Error(updateError.Error())
			}
		}
	}

	// We have enough total free we can go ahead and move forward
	_, createIndexSymbolError := indexCommandManager.databaseMgr.CreateIndexSymbolModel(dto.IndexedSymbolModel{
		Symbol:            symbolToAdd,
		Locked:            symbolLocked,
		CurrentPrice:      symbolQuotePrice,
		DesiredPercentage: symbolPercentageConverted,
	})

	if createIndexSymbolError != nil {
		logrus.Error(createIndexSymbolError.Error())
		return
	}

	logrus.Info("Symbol " + symbolToAdd + " added to index")

}

func (indexCommandManager *IndexCommandManager) StartIndexCommand(c *ishell.Context) {


	rebalanceStartError := indexCommandManager.rebalanceMgr.StartRebalanceProcess()

	if rebalanceStartError != nil {
		logrus.Error(rebalanceStartError.Error())
	}

	logrus.Info("Rebalance process initiated")
}

func (indexCommandManager *IndexCommandManager) GenerateIndexCommand(c *ishell.Context) {

	accountBalance, accountBalanceError := (*indexCommandManager.brokerIntegration).GetAccountValue()

	if accountBalanceError != nil {
		logrus.Error(accountBalanceError.Error())
		return
	}

	condextConfigModel, condextConfigModelError := indexCommandManager.databaseMgr.GetCondextConfigModel()

	if condextConfigModelError != nil {
		logrus.Error(condextConfigModelError.Error())
		return
	}

	if accountBalance < condextConfigModel.StartingBalance {
		logrus.Error("The balance in the account is lower than the starting balance")
		return
	}

	// Get all the indexed symbols
	indexedSymbols, indexedSymbolsError := indexCommandManager.databaseMgr.GetAllIndexedSymbols()

	if indexedSymbolsError != nil {
		logrus.Error(indexedSymbolsError.Error())
		return
	}

	for _, element := range indexedSymbols {
		if element.Symbol != "USD" {

			// We get the latest quote before processing
			symbolQuote, symbolQuoteError := (*indexCommandManager.brokerIntegration).GetSymbolQuotePrice(element.Symbol)

			if symbolQuoteError != nil {
				logrus.Error(symbolQuoteError.Error())
				continue
			}

			usdPercentageValue := util.GetPercentage(condextConfigModel.StartingBalance, element.DesiredPercentage)
			amountToBuy := decimal.NewFromFloat(usdPercentageValue).Div(decimal.NewFromFloat(symbolQuote)).IntPart()

			buyError := (*indexCommandManager.brokerIntegration).FulFillMarketOrderBuy(element.Symbol, amountToBuy)

			if buyError != nil {
				logrus.Error(buyError.Error())
				continue
			}

			element.CurrentPrice = symbolQuote
			element.Amount = amountToBuy
			element.CurrentPercentage = element.DesiredPercentage

			_, updateSymbolError := indexCommandManager.databaseMgr.UpdateIndexedSymbolModel(element)

			if updateSymbolError != nil {
				logrus.Error(updateSymbolError.Error())
			}
		}
	}

	condextConfigModel.Active = true

	_, updateError := indexCommandManager.databaseMgr.UpdateCondextConfig(condextConfigModel)

	if updateError != nil {
		logrus.Error(updateError.Error())
	}
}
