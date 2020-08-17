package broker_integrations

import (
	"errors"
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"time"
)

type AlpacaBrokerIntegration struct {
	AccessKey    string
	AccessSecret string
}

func CreateAlpacaBrokerIntegration() *AlpacaBrokerIntegration {
	return &AlpacaBrokerIntegration{}
}

func (alpacaBrokerIntegration *AlpacaBrokerIntegration) Connect(connectionUrl string) error {
	alpaca.SetBaseUrl(connectionUrl)
	return nil
}

func (alpacaBrokerIntegration *AlpacaBrokerIntegration) SetCredentials(credentials []string) error {

	if len(credentials) != 2 {
		return errors.New("not enough credentials supplied for alpaca connection")
	}

	alpacaBrokerIntegration.AccessKey = credentials[0]
	alpacaBrokerIntegration.AccessSecret = credentials[1]

	return nil
}

func (alpacaBrokerIntegration *AlpacaBrokerIntegration) ValidateCredentials() (bool, error) {

	alpacaClient := alpaca.NewClient(&common.APIKey{
		ID:           alpacaBrokerIntegration.AccessKey,
		Secret:       alpacaBrokerIntegration.AccessSecret,
		PolygonKeyID: alpacaBrokerIntegration.AccessKey,
	})

	_, accountError := alpacaClient.GetAccount()

	if accountError != nil {
		return false, accountError
	}

	return true, nil

}

func (alpacaBrokerIntegration *AlpacaBrokerIntegration) GetAccountValue() (float64, error) {

	alpacaClient := alpaca.NewClient(&common.APIKey{
		ID:           alpacaBrokerIntegration.AccessKey,
		Secret:       alpacaBrokerIntegration.AccessSecret,
		PolygonKeyID: alpacaBrokerIntegration.AccessKey,
	})

	accountInfo, accountError := alpacaClient.GetAccount()

	if accountError != nil {
		return 0, accountError
	}

	accountValueConv, _ := accountInfo.PortfolioValue.Float64()

	return accountValueConv, nil

}

func (alpacaBrokerIntegration *AlpacaBrokerIntegration) CheckIfSymbolIsValid(symbol string) (bool, error) {
	alpacaClient := alpaca.NewClient(&common.APIKey{
		ID:           alpacaBrokerIntegration.AccessKey,
		Secret:       alpacaBrokerIntegration.AccessSecret,
		PolygonKeyID: alpacaBrokerIntegration.AccessKey,
	})

	_, accountError := alpacaClient.GetAsset(symbol)

	if accountError != nil {
		return false, accountError
	}

	return true, nil
}

func (alpacaBrokerIntegration *AlpacaBrokerIntegration) GetSymbolQuotePrice(symbol string) (float64, error) {
	alpacaClient := alpaca.NewClient(&common.APIKey{
		ID:           alpacaBrokerIntegration.AccessKey,
		Secret:       alpacaBrokerIntegration.AccessSecret,
		PolygonKeyID: alpacaBrokerIntegration.AccessKey,
	})

	quote, quoteError := alpacaClient.GetLastQuote(symbol)

	if quoteError != nil {
		return 0.0, quoteError
	}

	midQuoteValue, _ := decimal.NewFromFloat32(quote.Last.BidPrice).Add(decimal.NewFromFloat32(quote.Last.AskPrice)).Div(decimal.NewFromInt(2)).Round(3).Float64()

	return midQuoteValue, nil
}

func (alpacaBrokerIntegration *AlpacaBrokerIntegration) FulFillMarketOrderBuy(symbol string, amount int64) error {

	alpacaClient := alpaca.NewClient(&common.APIKey{
		ID:           alpacaBrokerIntegration.AccessKey,
		Secret:       alpacaBrokerIntegration.AccessSecret,
		PolygonKeyID: alpacaBrokerIntegration.AccessKey,
	})

	accountInfo, accountError := alpacaClient.GetAccount()

	if accountError != nil {
		return accountError
	}

	placeOrderRequest := alpaca.PlaceOrderRequest{
		AccountID:   accountInfo.ID,
		AssetKey:    &symbol,
		Qty:         decimal.NewFromInt(amount),
		TimeInForce: "gtc",
		Type:        alpaca.Market,
		Side:        alpaca.Buy,
	}

	order, orderError := alpacaClient.PlaceOrder(placeOrderRequest)

	logrus.Info("Placed Market Buy - Symbol: " + symbol + " Amount: " + decimal.NewFromInt(amount).String())

	if orderError != nil {
		return orderError
	}

	filled := false

	for filled != true {

		time.Sleep(time.Duration(1) * time.Second)

		orderInfo, orderInfoError := alpacaClient.GetOrder(order.ID)

		if orderInfoError != nil {
			logrus.Error(orderError.Error())
			continue
		}

		if orderInfo.Status == "filled" {

			logrus.Info("Market Buy For - Symbol: " + symbol + " Amount: " + decimal.NewFromInt(amount).String() + " - Filled")

			filled = true
			break
		}

		if orderInfo.Status == "rejected" {
			return errors.New("Market Buy For - Symbol: " + symbol + " Amount: " + decimal.NewFromInt(amount).String() + " - Rejected")
		}

		logrus.Info("Market Buy For - Symbol: " + symbol + " Amount: " + decimal.NewFromInt(amount).String() + " - Not filled yet")

	}

	return nil
}

func (alpacaBrokerIntegration *AlpacaBrokerIntegration) FulFillMarketOrderSell(symbol string, amount int64) error {

	alpacaClient := alpaca.NewClient(&common.APIKey{
		ID:           alpacaBrokerIntegration.AccessKey,
		Secret:       alpacaBrokerIntegration.AccessSecret,
		PolygonKeyID: alpacaBrokerIntegration.AccessKey,
	})

	accountInfo, accountError := alpacaClient.GetAccount()

	if accountError != nil {
		return accountError
	}

	placeOrderRequest := alpaca.PlaceOrderRequest{
		AccountID:   accountInfo.ID,
		AssetKey:    &symbol,
		Qty:         decimal.NewFromInt(amount),
		TimeInForce: "gtc",
		Type:        alpaca.Market,
		Side:        alpaca.Sell,
	}

	order, orderError := alpacaClient.PlaceOrder(placeOrderRequest)

	logrus.Info("Placed Market Sell - Symbol: " + symbol + " Amount: " + decimal.NewFromInt(amount).String())

	if orderError != nil {
		return orderError
	}

	filled := false

	for filled != true {

		time.Sleep(time.Duration(1) * time.Second)

		orderInfo, orderInfoError := alpacaClient.GetOrder(order.ID)

		if orderInfoError != nil {
			logrus.Error(orderError.Error())
			continue
		}

		if orderInfo.Status == "filled" {
			filled = true

			logrus.Info("Market Sell For - Symbol: " + symbol + " Amount: " + decimal.NewFromInt(amount).String() + " - Filled")

			break
		}

		if orderInfo.Status == "rejected" {
			return errors.New("Market Sell For - Symbol: " + symbol + " Amount: " + decimal.NewFromInt(amount).String() + " - Rejected")
		}

		logrus.Info("Market Sell For - Symbol: " + symbol + " Amount: " + decimal.NewFromInt(amount).String() + " - Not filled yet")

	}

	return nil
}
