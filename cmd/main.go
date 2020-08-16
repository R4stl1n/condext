package main

import (
	"encoding/json"
	broker_integrations "github.com/r4stl1n/condext/pkg/broker-integrations"
	"github.com/r4stl1n/condext/pkg/managers"
	"github.com/r4stl1n/condext/pkg/util"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

func main() {

	configStruct := util.ConfigStruct{}

	util.PrintBanner()

	logrus.SetLevel(logrus.DebugLevel)

	logrus.Info("Loading configuration file")

	configFile, configLoadError := ioutil.ReadFile("config.json")

	if configLoadError != nil {

		logrus.Warn("Could not load the config file")
		logrus.Info("Creating new config file, modify and run again")

		marshaledStruct, marshaledStructError := json.MarshalIndent(configStruct, "", "")

		if marshaledStructError != nil {
			logrus.Fatal("Failed to marshal blank structure")
		}

		writeFileError := ioutil.WriteFile("config.json", marshaledStruct, 0644)

		if writeFileError != nil {
			logrus.Fatal("Could not write empty config file")
		}

		return
	}

	logrus.Info("config file loaded")
	configUnmarshalError := json.Unmarshal(configFile, &configStruct)

	if configUnmarshalError != nil {
		logrus.Fatal("Could not unmarshal config file")
	}

	logrus.Info("Connecting to database")

	databaseManager, databaseManagerError := managers.CreateDatabaseManager("data.db")

	if databaseManagerError != nil {
		logrus.Panic(databaseManagerError.Error())
		return
	}

	// Create the config model if missing
	configModelError := databaseManager.CreateCondextConfigModel()

	if configModelError != nil {
		logrus.Error(configModelError)
		return
	}

	// Create the broker integration
	brokerIntegration := broker_integrations.CreateAlpacaBrokerIntegration()

	logrus.Info("Setting the broker credentials")

	brokerCredsError := brokerIntegration.SetCredentials([]string{configStruct.AlpacaApi, configStruct.AlpacaSecret})

	if brokerCredsError != nil {
		logrus.Error(brokerCredsError.Error())
		return
	}

	logrus.Info("Connecting to the broker")
	brokerConnectionError := brokerIntegration.Connect("https://paper-api.alpaca.markets")

	if brokerConnectionError != nil {
		logrus.Error(brokerConnectionError.Error())
		return
	}

	logrus.Info("Validating credentials to the broker")
	brokerValidCredentials, brokerValidateCredentialsError := brokerIntegration.ValidateCredentials()

	if brokerValidateCredentialsError != nil {
		logrus.Error(brokerValidateCredentialsError.Error())
		return
	}

	if brokerValidCredentials == false {
		logrus.Error("Invalid broker credentials")
		return
	}

	// Create the command manager
	showCommandManager := managers.CreateShowCommandManager(databaseManager, brokerIntegration)
	indexCommandManager := managers.CreateIndexCommandManager(databaseManager, brokerIntegration)

	serviceManager := managers.CreateServiceManager(&configStruct, databaseManager, showCommandManager, indexCommandManager)

	serviceInitError := serviceManager.Initialize()

	if serviceInitError != nil {
		logrus.Fatal(serviceInitError)
	}

	serviceManager.Run()

}
