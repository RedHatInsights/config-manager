package infrastructure

import (
	"config-manager/api/controllers"
	"config-manager/application"
	"config-manager/domain"
	"config-manager/infrastructure/persistence"
	"config-manager/infrastructure/persistence/cloudconnector"
	"config-manager/infrastructure/persistence/dispatcher"
	"config-manager/internal/config"
	"config-manager/internal/db"
	"config-manager/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/labstack/echo/v4"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Container is the primary application structure. It holds references to the
// various API controllers, service managers and external data repositories. It
// provides a collection of accessor methods to retrieve handles to each
// application component.
type Container struct {
	db     *db.DB
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
	cloudConnectorRepo cloudconnector.CloudConnectorClient
	inventoryRepo      *persistence.InventoryClient
}

// Database lazily initializes a db.DB, performs any necessary migrations, and
// returns it.
func (c *Container) Database() *db.DB {
	if c.db == nil {
		connectionString := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable",
			config.DefaultConfig.DBUser,
			config.DefaultConfig.DBPass,
			config.DefaultConfig.DBName,
			config.DefaultConfig.DBHost,
			config.DefaultConfig.DBPort)

		db, err := db.Open("pgx", connectionString)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot open database")
		}

		if err := db.Migrate("file://./db/migrations", false); err != nil {
			log.Fatal().Err(err).Msg("cannot migrate database")
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
		templates := utils.FilesIntoMap(config.DefaultConfig.PlaybookFiles, "*.yml")
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
			URLBasePath:          config.DefaultConfig.URLBasePath(),
			DB:                   c.Database(),
		}
	}

	return c.cmController
}

// AccountStateRepo lazily initializes a new persistence.AccountStateRepository
// and returns it.
func (c *Container) AccountStateRepo() *persistence.AccountStateRepository {
	if c.accountStateRepo == nil {
		c.accountStateRepo = &persistence.AccountStateRepository{
			DB: c.Database().Handle(),
		}
	}

	return c.accountStateRepo
}

// StateArchiveRepo lazily initializes a new persistence.StateArchiveRepository
// and returns it.
func (c *Container) StateArchiveRepo() *persistence.StateArchiveRepository {
	if c.stateArchiveRepo == nil {
		c.stateArchiveRepo = &persistence.StateArchiveRepository{
			DB: c.Database().Handle(),
		}
	}

	return c.stateArchiveRepo
}

// DispatcherRepo lazily initializes a new dispatcher.DispatcherClient and
// returns it.
func (c *Container) DispatcherRepo() dispatcher.DispatcherClient {
	if c.dispatcherRepo == nil {
		if config.DefaultConfig.DispatcherImpl.Value == "mock" {
			c.dispatcherRepo = dispatcher.NewDispatcherClientMock()
		} else {
			c.dispatcherRepo = dispatcher.NewDispatcherClient()
		}
	}

	return c.dispatcherRepo
}

// CloudConnectorRepo lazily initializes a new persistence.CloundConnectorClient
// and returns it.
func (c *Container) CloudConnectorRepo() cloudconnector.CloudConnectorClient {
	if c.cloudConnectorRepo == nil {
		if config.DefaultConfig.CloudConnectorImpl.Value == "mock" {
			c.cloudConnectorRepo = cloudconnector.NewCloudConnectorClientMock()
		} else {
			client, err := cloudconnector.NewCloudConnectorClient()
			if err != nil {
				log.Fatal().Err(err).Msg("cannot create cloud connector client")
			}
			c.cloudConnectorRepo = client
		}
	}

	return c.cloudConnectorRepo
}

// InventoryRepo lazily initializes a new persistence.InventoryClient and
// returns it.
func (c *Container) InventoryRepo() domain.InventoryClient {
	if c.inventoryRepo == nil {
		client := &http.Client{
			Timeout: time.Duration(int(time.Second) * config.DefaultConfig.InventoryTimeout),
		}

		c.inventoryRepo = &persistence.InventoryClient{
			InventoryHost: config.DefaultConfig.InventoryHost.String(),
			InventoryImpl: config.DefaultConfig.InventoryImpl.Value,
			Client:        client,
		}
	}

	return c.inventoryRepo
}
