package managers

import (
	"errors"
	broker_integrations "github.com/r4stl1n/condext/pkg/broker-integrations"
	"github.com/r4stl1n/condext/pkg/util"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"time"
)

type RebalanceManager struct {
	databaseMgr             *DatabaseManager
	brokerIntegration       *broker_integrations.BrokerIntegrationInterface
	rebalanceProcessRunning bool
	rebalanceFrequency      int64
	startingBalance         float64
}

func CreateRebalanceManager(databaseManager *DatabaseManager, selectedBrokerIntegration broker_integrations.BrokerIntegrationInterface) *RebalanceManager {

	return &RebalanceManager{
		databaseMgr:             databaseManager,
		brokerIntegration:       &selectedBrokerIntegration,
		rebalanceProcessRunning: false,
	}
}

func (rebalanceManager *RebalanceManager) calculateCurrentPercentages() error {

	logrus.Info("Processing percentage changes")

	allIndexedSymbols, allIndexedSymbolsError := rebalanceManager.databaseMgr.GetAllIndexedSymbols()

	if allIndexedSymbolsError != nil {
		return allIndexedSymbolsError
	}

	for _, element := range allIndexedSymbols {

		// Grab the quote for the symbol
		currentQuote, currentQuoteError := (*rebalanceManager.brokerIntegration).GetSymbolQuotePrice(element.Symbol)

		if currentQuoteError != nil {
			logrus.Error(currentQuoteError.Error())
			continue
		}

		// Now we calculate the current value of the holdings
		currentHoldingUSDValue, _ := decimal.NewFromFloat(currentQuote).Mul(decimal.NewFromInt(element.Amount)).Round(2).Float64()

		// Calculate the current percentage amount we have above / below the desired for the index
		desiredUSDValue := util.GetPercentage(rebalanceManager.startingBalance, element.DesiredPercentage)

		percentageDifference := util.GetPercentageDifference(desiredUSDValue, currentHoldingUSDValue)

		element.CurrentPercentage, _ = decimal.NewFromFloat(element.DesiredPercentage).Add(decimal.NewFromFloat(percentageDifference)).Round(2).Float64()
		element.CurrentPrice = currentQuote

		logrus.Info("Symbol " + element.Symbol + " new current percentage is - " + decimal.NewFromFloat(percentageDifference).String())

		_, updateSymbolError := rebalanceManager.databaseMgr.UpdateIndexedSymbolModel(element)

		if updateSymbolError != nil {
			logrus.Error(updateSymbolError.Error())
		}
	}

	return nil
}

func (rebalanceManager *RebalanceManager) handleTrades() error {

	configModel, configModelError := rebalanceManager.databaseMgr.GetCondextConfigModel()

	if configModelError != nil {
		return configModelError
	}

	// Now we need to go through all the ones who are over their percentage and rebalance threshold
	allIndexedSymbols, allIndexedSymbolsError := rebalanceManager.databaseMgr.GetAllIndexedSymbols()

	if allIndexedSymbolsError != nil {
		return allIndexedSymbolsError
	}

	// We are going to do this sloppy first we are going to iterate on all the ones we need to sell
	for _, element := range allIndexedSymbols {

		percentageDifference := decimal.NewFromFloat(element.CurrentPercentage).Sub(decimal.NewFromFloat(element.DesiredPercentage)).Round(2)
		percentageDifferenceConv, _ := percentageDifference.Float64()

		if percentageDifference.IsPositive() {
			if percentageDifference.GreaterThan(decimal.NewFromFloat(configModel.ReBalanceThreshold)) == true {

				// If we are above the threshold we now are going to try and sell the above threshold amount
				currentHoldingUSDValue, _ := decimal.NewFromFloat(element.CurrentPrice).Mul(decimal.NewFromInt(element.Amount)).Round(2).Float64()

				// Get the percentage difference in usd
				percentageDifferenceInUsd := util.GetPercentage(currentHoldingUSDValue, percentageDifferenceConv)

				// Now we need to calculate how many we can sell
				amountToSell := decimal.NewFromFloat(percentageDifferenceInUsd).Div(decimal.NewFromFloat(element.CurrentPrice)).IntPart()

				if amountToSell == 0 {
					logrus.Warn("Unable to partial sell " + element.Symbol + " current percentage is to low to fullfill amount")
					continue
				}

				sellError := (*rebalanceManager.brokerIntegration).FulFillMarketOrderSell(element.Symbol, amountToSell)

				if sellError != nil {
					logrus.Error(sellError.Error())
					continue
				}

				element.Amount = element.Amount - amountToSell

				_, symbolUpdateError := rebalanceManager.databaseMgr.UpdateIndexedSymbolModel(element)

				if symbolUpdateError != nil {
					logrus.Error(symbolUpdateError)
				}

				configModel.FloatingPercentage = configModel.FloatingPercentage + percentageDifferenceConv
			}
		}

	}

	// We are now going to look at what we need to buy.
	for _, element := range allIndexedSymbols {

		percentageDifference := decimal.NewFromFloat(element.CurrentPercentage).Sub(decimal.NewFromFloat(element.DesiredPercentage)).Round(2)
		percentageDifferenceConv, _ := percentageDifference.Float64()

		if percentageDifference.IsNegative() {

			if percentageDifference.Abs().GreaterThan(decimal.NewFromFloat(configModel.ReBalanceThreshold)) == true {

				if configModel.FloatingPercentage > percentageDifferenceConv {
					// If we are above the threshold we now are going to try and buy the above threshold amount
					currentHoldingUSDValue, _ := decimal.NewFromFloat(element.CurrentPrice).Mul(decimal.NewFromInt(element.Amount)).Round(2).Float64()

					// Get the percentage difference in usd
					percentageDifferenceInUsd := util.GetPercentage(currentHoldingUSDValue, percentageDifferenceConv)

					// Now we need to calculate how many we can buy
					amountToBuy := decimal.NewFromFloat(percentageDifferenceInUsd).Div(decimal.NewFromFloat(element.CurrentPrice)).Abs().IntPart()

					if amountToBuy == 0 {
						logrus.Warn("Unable to partial buy " + element.Symbol + " current percentage is to low to fullfill amount")
						continue
					}

					buyError := (*rebalanceManager.brokerIntegration).FulFillMarketOrderBuy(element.Symbol, amountToBuy)

					if buyError != nil {
						logrus.Error(buyError.Error())
						continue
					}

					element.Amount = element.Amount + amountToBuy

					_, symbolUpdateError := rebalanceManager.databaseMgr.UpdateIndexedSymbolModel(element)

					if symbolUpdateError != nil {
						logrus.Error(symbolUpdateError)
					}

					configModel.FloatingPercentage = configModel.FloatingPercentage - percentageDifferenceConv
				} else {

					logrus.Warn("Current floating percentage is not large enough to fulfill buy need of " + element.Symbol)

				}
			}
		}

	}

	_, configUpdateError := rebalanceManager.databaseMgr.UpdateCondextConfig(configModel)

	if configUpdateError != nil {
		logrus.Error(configUpdateError.Error())
	}

	return nil
}

func (rebalanceManager *RebalanceManager) rebalanceRoutine() {

	for rebalanceManager.rebalanceProcessRunning == true {

		calculateError := rebalanceManager.calculateCurrentPercentages()

		if calculateError != nil {
			logrus.Error(calculateError.Error())
			continue
		}

		handleTradesError := rebalanceManager.handleTrades()

		if handleTradesError != nil {
			logrus.Error(handleTradesError)
		}

		time.Sleep(time.Duration(rebalanceManager.rebalanceFrequency) * time.Second)


	}
}

func (rebalanceManager *RebalanceManager) StartRebalanceProcess() error {

	configModel, configModelError := rebalanceManager.databaseMgr.GetCondextConfigModel()

	if configModelError != nil {
		return configModelError
	}

	if configModel.Active != true {
		return errors.New("you need to generate the index before calling start")
	}

	if rebalanceManager.rebalanceProcessRunning == true {
		return errors.New("rebalance process already started")
	}

	rebalanceManager.rebalanceFrequency = configModel.RebalanceFrequency
	rebalanceManager.startingBalance = configModel.StartingBalance

	go func() {
		rebalanceManager.rebalanceRoutine()
	}()

	rebalanceManager.rebalanceProcessRunning = true

	return nil
}
