package managers

import (
	"errors"
	broker_integrations "github.com/r4stl1n/condext/pkg/broker-integrations"
	"github.com/sirupsen/logrus"
	"time"
)

type RebalanceManager struct {
	databaseMgr             *DatabaseManager
	brokerIntegration       *broker_integrations.BrokerIntegrationInterface
	rebalanceProcessRunning bool
	rebalanceFrequency      int64
}

func CreateRebalanceManager(databaseManager *DatabaseManager, selectedBrokerIntegration broker_integrations.BrokerIntegrationInterface) *RebalanceManager {

	return &RebalanceManager{
		databaseMgr:             databaseManager,
		brokerIntegration:       &selectedBrokerIntegration,
		rebalanceProcessRunning: false,
	}
}

func (rebalanceManager *RebalanceManager) rebalanceRoutine() {

	for rebalanceManager.rebalanceProcessRunning == true {

		time.Sleep(time.Duration(rebalanceManager.rebalanceFrequency) * time.Second)

		logrus.Info("Cycled")

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

	go func() {
		rebalanceManager.rebalanceRoutine()
	}()

	rebalanceManager.rebalanceProcessRunning = true

	return nil
}
