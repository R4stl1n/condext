package managers

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/r4stl1n/condext/pkg/broker-integrations"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gopkg.in/abiosoft/ishell.v2"
	"os"
	"strconv"
)

type ShowCommandManager struct {
	databaseMgr       *DatabaseManager
	brokerIntegration *broker_integrations.BrokerIntegrationInterface
}

func CreateShowCommandManager(databaseManager *DatabaseManager, selectedBrokerIntegration broker_integrations.BrokerIntegrationInterface) *ShowCommandManager {

	return &ShowCommandManager{
		databaseMgr:       databaseManager,
		brokerIntegration: &selectedBrokerIntegration,
	}
}

func (showCommandManager *ShowCommandManager) ShowIndex(c *ishell.Context) {

	data := [][]string{}

	allIndexedSymbols, allIndexedSymbolsError := showCommandManager.databaseMgr.GetAllIndexedSymbols()

	if allIndexedSymbolsError != nil {
		logrus.Error(allIndexedSymbolsError.Error())
		return
	}

	for _, element := range allIndexedSymbols {
		currentValue := decimal.NewFromInt(element.Amount).Mul(decimal.NewFromFloat(element.CurrentPrice))
		data = append(data, []string{element.Symbol, decimal.NewFromInt(element.Amount).String(),
			currentValue.String(), decimal.NewFromFloat(element.CurrentPrice).String(), strconv.FormatBool(element.Locked),
			decimal.NewFromFloat(element.DesiredPercentage).String(), decimal.NewFromFloat(element.CurrentPercentage).String()})
	}

	fmt.Println()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Symbol", "Amount", "Current USD Value", "Current Price", "Locked", "Desired %", "Current %"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data) // Add Bulk Data
	table.Render()
	fmt.Println()
}

func (showCommandManager *ShowCommandManager) ShowStats(c *ishell.Context) {

	allIndexedSymbols, allIndexedSymbolsError := showCommandManager.databaseMgr.GetAllIndexedSymbols()

	if allIndexedSymbolsError != nil {
		logrus.Error(allIndexedSymbolsError.Error())
		return
	}

	accountBalance, accountBalanceError := (*showCommandManager.brokerIntegration).GetAccountValue()

	if accountBalanceError != nil {
		logrus.Error(accountBalanceError.Error())
		return
	}

	data := [][]string{
		{decimal.NewFromFloat(accountBalance).String(), decimal.NewFromInt(int64(len(allIndexedSymbols))).String()},
	}

	fmt.Println()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Account Value", "# Indexed"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data) // Add Bulk Data
	table.Render()
	fmt.Println()
}

func (showCommandManager *ShowCommandManager) ShowConfig(c *ishell.Context) {

	configModel, configModelError := showCommandManager.databaseMgr.GetCondextConfigModel()

	if configModelError != nil {
		logrus.Error(configModelError.Error())
		return
	}

	data := [][]string{
		{
			strconv.FormatBool(configModel.Active),
			decimal.NewFromFloat(configModel.ReBalanceThreshold).String(),
			decimal.NewFromInt(configModel.OrderTimeout).String(),
			decimal.NewFromInt(configModel.RebalanceFrequency).String(),
		},
	}

	fmt.Println()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Active", "Balance Threshold %", "Order Timeout", "ReBalance Tick Setting"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data) // Add Bulk Data
	table.Render()
	fmt.Println()
}
