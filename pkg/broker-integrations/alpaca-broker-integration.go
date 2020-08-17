package broker_integrations

import (
	"errors"
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/shopspring/decimal"
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
