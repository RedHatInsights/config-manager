package infrastructure

import (
	"config-manager/api/controllers"
	"config-manager/application"
	"config-manager/infrastructure/persistence"
	"config-manager/utils"
	"database/sql"
	"fmt"
	"log"

	"github.com/labstack/echo/v4"

	goMigrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/viper"
)

// Container holds application resources
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
	accountStateRepo *persistence.AccountStateRepository
	stateArchiveRepo *persistence.StateArchiveRepository
	clientListRepo   *persistence.ClientListRepository
	dispatcherRepo   *persistence.DispatcherRepository
}

// Database configures and opens a db connection
func (c *Container) Database() *sql.DB {
	if c.db == nil {
		connectionString := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable",
			c.Config.GetString("DBUser"),
			c.Config.GetString("DBPass"),
			c.Config.GetString("DBName"),
			c.Config.GetString("DBHost"),
			c.Config.GetInt("DBPort"))

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

// Server initializes a new echo server
func (c *Container) Server() *echo.Echo {
	if c.server == nil {
		c.server = echo.New()
	}

	return c.server
}

// CMService provides access to various application resources
func (c *Container) CMService() *application.ConfigManagerService {
	if c.cmService == nil {
		c.cmService = &application.ConfigManagerService{
			Cfg:               c.Config,
			AccountStateRepo:  c.AccountStateRepo(),
			StateArchiveRepo:  c.StateArchiveRepo(),
			ClientListRepo:    c.ClientListRepo(),
			DispatcherRepo:    c.DispatcherRepo(),
			PlaybookGenerator: *c.PlaybookGenerator(),
		}
	}

	return c.cmService
}

func (c *Container) PlaybookGenerator() *application.Generator {
	if c.playbookGenerator == nil {
		templates := utils.FilesIntoMap(c.Config.GetString("PlaybookPath"), "*.yml")
		c.playbookGenerator = &application.Generator{
			Templates: templates,
		}
	}

	return c.playbookGenerator
}

// CMController sets up handlers for api routes
func (c *Container) CMController() *controllers.ConfigManagerController {
	if c.cmController == nil {
		c.cmController = &controllers.ConfigManagerController{
			ConfigManagerService: c.CMService(),
			Server:               c.Server(),
			URLBasePath:          c.Config.GetString("URLBasePath"),
		}
	}

	return c.cmController
}

// AccountStateRepo enables interaction with the account_states db table
func (c *Container) AccountStateRepo() *persistence.AccountStateRepository {
	if c.accountStateRepo == nil {
		c.accountStateRepo = &persistence.AccountStateRepository{
			DB: c.Database(),
		}
	}

	return c.accountStateRepo
}

// StateArchiveRepo enables interaction with the state_archives db table
func (c *Container) StateArchiveRepo() *persistence.StateArchiveRepository {
	if c.stateArchiveRepo == nil {
		c.stateArchiveRepo = &persistence.StateArchiveRepository{
			DB: c.Database(),
		}
	}

	return c.stateArchiveRepo
}

// ClientListRepo enables interaction with inventory (needs a different name)
func (c *Container) ClientListRepo() *persistence.ClientListRepository {
	if c.clientListRepo == nil {
		c.clientListRepo = &persistence.ClientListRepository{
			InventoryURL: "",
		}
	}

	return c.clientListRepo
}

// DispatcherRepo enables interaction with the playbook dispatcher
func (c *Container) DispatcherRepo() *persistence.DispatcherRepository {
	if c.dispatcherRepo == nil {
		c.dispatcherRepo = &persistence.DispatcherRepository{
			DispatcherURL: "",
			PlaybookURL:   "",
			RunStatusURL:  "",
		}
	}

	return c.dispatcherRepo
}
