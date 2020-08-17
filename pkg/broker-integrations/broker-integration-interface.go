package broker_integrations

type BrokerIntegrationInterface interface {
	Connect(connectionUrl string) error
	SetCredentials(credentials []string) error
	ValidateCredentials() (bool, error)
	GetAccountValue() (float64, error)
	GetSymbolQuotePrice(symbol string) (float64, error)
	CheckIfSymbolIsValid(symbol string) (bool, error)
}
