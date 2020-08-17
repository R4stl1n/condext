package managers

import (
	"github.com/r4stl1n/condext/pkg/util"
	"github.com/sirupsen/logrus"
	"gopkg.in/abiosoft/ishell.v2"
)

type ServiceManager struct {
	config              *util.ConfigStruct
	databaseMgr         *DatabaseManager
	showCommandMgr      *ShowCommandManager
	indexCommandManager *IndexCommandManager
}

func CreateServiceManager(config *util.ConfigStruct, databaseClient *DatabaseManager, showCommandManager *ShowCommandManager, indexCommandManager *IndexCommandManager) *ServiceManager {

	return &ServiceManager{
		config:              config,
		databaseMgr:         databaseClient,
		showCommandMgr:      showCommandManager,
		indexCommandManager: indexCommandManager,
	}

}

func (serviceManager *ServiceManager) Initialize() error {
	return nil
}

func (serviceManager *ServiceManager) Run() {
	shell := ishell.New()

	// display welcome info.
	logrus.Println("Condext Ready")

	// register a function for "greet" command.
	shell.AddCmd(&ishell.Cmd{
		Name: "index_add",
		Help: "Add a symbol to be indexed, def: index_add <symbol> <percentage> <locked>, ex. index_add AAPL 5 false",
		Func: serviceManager.indexCommandManager.AddSymbolToIndexCommand,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "index_gen",
		Help: "Initial generation of the index",
		Func: serviceManager.indexCommandManager.AddSymbolToIndexCommand,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "index_start",
		Help: "Starts the rebalance background process",
		Func: serviceManager.indexCommandManager.StartIndexCommand,
	})


	shell.AddCmd(&ishell.Cmd{
		Name: "show_stats",
		Help: "Shows index stats",
		Func: serviceManager.showCommandMgr.ShowStats,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "show_config",
		Help: "Shows index configuration",
		Func: serviceManager.showCommandMgr.ShowConfig,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "show_index",
		Help: "Shows the current index data",
		Func: serviceManager.showCommandMgr.ShowIndex,
	})

	// run shell
	shell.Run()
}
