package infrastructure

import (
	"config-manager/api"
	"config-manager/application"
	"config-manager/infrastructure/persistence"
	"database/sql"
	"fmt"
	"log"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

type Container struct {
	Config *viper.Viper
	db     *sql.DB
	apiMux *mux.Router

	// Config Manager Services
	cmService *application.ConfigManagerService

	// Config Manager Controllers
	cmController *api.ConfigManagerController
	apiSpec      *api.ApiSpecServer

	// Repositories
	accountStateRepo *persistence.AccountStateRepository
	runRepo          *persistence.RunRepository
	stateArchiveRepo *persistence.StateArchiveRepository
	clientListRepo   *persistence.ClientListRepository
	dispatcherRepo   *persistence.DispatcherRepository
}

func (c *Container) Database() *sql.DB {
	if c.db == nil {
		connectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
			c.Config.GetString("DBUser"),
			c.Config.GetString("DBPass"),
			c.Config.GetString("DBName"))

		db, err := sql.Open("postgres", connectionString)
		if err != nil {
			log.Fatal(err)
		}

		c.db = db
	}

	return c.db
}

func (c *Container) Mux() *mux.Router {
	if c.apiMux == nil {
		c.apiMux = mux.NewRouter()
	}

	return c.apiMux
}

func (c *Container) CMService() *application.ConfigManagerService {
	if c.cmService == nil {
		c.cmService = &application.ConfigManagerService{
			AccountStateRepo: c.AccountStateRepo(),
			RunRepo:          c.RunRepo(),
			StateArchiveRepo: c.StateArchiveRepo(),
			ClientListRepo:   c.ClientListRepo(),
			DispatcherRepo:   c.DispatcherRepo(),
		}
	}

	return c.cmService
}

func (c *Container) ApiSpec() *api.ApiSpecServer {
	if c.apiSpec == nil {
		c.apiSpec = &api.ApiSpecServer{
			Router:       c.Mux(),
			SpecFileName: c.Config.GetString("ApiSpecFile"),
		}
	}

	return c.apiSpec
}

func (c *Container) CMController() *api.ConfigManagerController {
	if c.cmController == nil {
		c.cmController = &api.ConfigManagerController{
			ConfigManagerService: c.CMService(),
			Router:               c.Mux(),
		}
	}

	return c.cmController
}

func (c *Container) AccountStateRepo() *persistence.AccountStateRepository {
	if c.accountStateRepo == nil {
		c.accountStateRepo = &persistence.AccountStateRepository{
			DB: c.Database(),
		}
	}

	return c.accountStateRepo
}

func (c *Container) RunRepo() *persistence.RunRepository {
	if c.runRepo == nil {
		c.runRepo = &persistence.RunRepository{
			DB: c.Database(),
		}
	}

	return c.runRepo
}

func (c *Container) StateArchiveRepo() *persistence.StateArchiveRepository {
	if c.stateArchiveRepo == nil {
		c.stateArchiveRepo = &persistence.StateArchiveRepository{
			DB: c.Database(),
		}
	}

	return c.stateArchiveRepo
}

func (c *Container) ClientListRepo() *persistence.ClientListRepository {
	if c.clientListRepo == nil {
		c.clientListRepo = &persistence.ClientListRepository{
			InventoryURL: "",
		}
	}

	return c.clientListRepo
}

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
