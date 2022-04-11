package infrastructure

import (
	"config-manager/api/controllers"
	"config-manager/application"
	"config-manager/domain"
	"config-manager/infrastructure/persistence"
	"config-manager/infrastructure/persistence/dispatcher"
	"config-manager/utils"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	goMigrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/viper"
)

// Container is the primary application structure. It holds references to the
// various API controllers, service managers and external data repositories. It
// provides a collection of accessor methods to retrieve handles to each
// application component.
type Container struct {
	Config *viper.Viper
	db     *sql.DB
	server *echo.Echo

	// Config Manager Services
	cmService         *application.ConfigManagerService
	playbookGenerator *application.Generator

	// API Controllers
	cmController *controllers.ConfigManagerController

	// Repositories
	accountStateRepo   *persistence.AccountStateRepository
	stateArchiveRepo   *persistence.StateArchiveRepository
	dispatcherRepo     dispatcher.DispatcherClient
	cloudConnectorRepo *persistence.CloudConnectorClient
	inventoryRepo      *persistence.InventoryClient
}

// Database lazily initializes an sql.DB, performs any necessary migrations, and
// returns it.
func (c *Container) Database() *sql.DB {
	if c.db == nil {
		connectionString := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable",
			c.Config.GetString("DB_User"),
			c.Config.GetString("DB_Pass"),
			c.Config.GetString("DB_Name"),
			c.Config.GetString("DB_Host"),
			c.Config.GetInt("DB_Port"))

		db, err := sql.Open("postgres", connectionString)
		if err != nil {
			log.Fatal(err)
		}

		err = db.Ping()
		if err != nil {
			log.Fatal(err)
		}

		driver, err := postgres.WithInstance(db, &postgres.Config{})
		if err != nil {
			log.Fatal(err)
		}
		m, err := goMigrate.NewWithDatabaseInstance(
			"file://./db/migrations",
			"postgres", driver)
		if err != nil {
			log.Fatal(err)
		}
		err = m.Up()
		if err != nil {
			if err != goMigrate.ErrNoChange {
				log.Fatal(err)
			} else {
				log.Println("no change")
			}
		}

		c.db = db
	}

	return c.db
}

// Server lazily initializes a new Echo HTTP server and returns it.
func (c *Container) Server() *echo.Echo {
	if c.server == nil {
		c.server = echo.New()
	}

	return c.server
}

// CMService lazily initializes a new application.ConfigManagerService and
// returns it.
func (c *Container) CMService() *application.ConfigManagerService {
	if c.cmService == nil {
		c.cmService = &application.ConfigManagerService{
			Cfg:                c.Config,
			AccountStateRepo:   c.AccountStateRepo(),
			StateArchiveRepo:   c.StateArchiveRepo(),
			CloudConnectorRepo: c.CloudConnectorRepo(),
			DispatcherRepo:     c.DispatcherRepo(),
			PlaybookGenerator:  *c.PlaybookGenerator(),
			InventoryRepo:      c.InventoryRepo(),
		}
	}

	return c.cmService
}

// PlaybookGenerator lazily initializes a new application.Generator and returns
// it.
func (c *Container) PlaybookGenerator() *application.Generator {
	if c.playbookGenerator == nil {
		templates := utils.FilesIntoMap(c.Config.GetString("Playbook_Files"), "*.yml")
		c.playbookGenerator = &application.Generator{
			Templates: templates,
		}
	}

	return c.playbookGenerator
}

// CMController lazily initializes a new controllers.ConfigManagerController and
// returns it.
func (c *Container) CMController() *controllers.ConfigManagerController {
	if c.cmController == nil {
		c.cmController = &controllers.ConfigManagerController{
			ConfigManagerService: c.CMService(),
			Server:               c.Server(),
			URLBasePath:          c.Config.GetString("URL_Base_Path"),
		}
	}

	return c.cmController
}

// AccountStateRepo lazily initializes a new persistence.AccountStateRepository
// and returns it.
func (c *Container) AccountStateRepo() *persistence.AccountStateRepository {
	if c.accountStateRepo == nil {
		c.accountStateRepo = &persistence.AccountStateRepository{
			DB: c.Database(),
		}
	}

	return c.accountStateRepo
}

// StateArchiveRepo lazily initializes a new persistence.StateArchiveRepository
// and returns it.
func (c *Container) StateArchiveRepo() *persistence.StateArchiveRepository {
	if c.stateArchiveRepo == nil {
		c.stateArchiveRepo = &persistence.StateArchiveRepository{
			DB: c.Database(),
		}
	}

	return c.stateArchiveRepo
}

// DispatcherRepo lazily initializes a new dispatcher.DispatcherClient and
// returns it.
func (c *Container) DispatcherRepo() dispatcher.DispatcherClient {
	if c.dispatcherRepo == nil {
		if c.Config.GetString("Dispatcher_Impl") == "mock" {
			c.dispatcherRepo = dispatcher.NewDispatcherClientMock()
		} else {
			c.dispatcherRepo = dispatcher.NewDispatcherClient(c.Config)
		}
	}

	return c.dispatcherRepo
}

// CloudConnectorRepo lazily initializes a new persistence.CloundConnectorClient
// and returns it.
func (c *Container) CloudConnectorRepo() domain.CloudConnectorClient {
	if c.cloudConnectorRepo == nil {
		client := &http.Client{
			Timeout: time.Duration(int(time.Second) * c.Config.GetInt("Cloud_Connector_Timeout")),
		}

		c.cloudConnectorRepo = &persistence.CloudConnectorClient{
			CloudConnectorHost:     c.Config.GetString("Cloud_Connector_Host"),
			CloudConnectorClientID: c.Config.GetString("Cloud_Connector_Client_ID"),
			CloudConnectorPSK:      c.Config.GetString("Cloud_Connector_PSK"),
			CloudConnectorImpl:     c.Config.GetString("Cloud_Connector_Impl"),
			Client:                 client,
		}
	}

	return c.cloudConnectorRepo
}

// InventoryRepo lazily initializes a new persistence.InventoryClient and
// returns it.
func (c *Container) InventoryRepo() domain.InventoryClient {
	if c.inventoryRepo == nil {
		client := &http.Client{
			Timeout: time.Duration(int(time.Second) * c.Config.GetInt("Inventory_Timeout")),
		}

		c.inventoryRepo = &persistence.InventoryClient{
			InventoryHost: c.Config.GetString("Inventory_Host"),
			InventoryImpl: c.Config.GetString("Inventory_Impl"),
			Client:        client,
		}
	}

	return c.inventoryRepo
}
