package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Kaese72/findings-registry/event"
	"github.com/Kaese72/findings-registry/internal/application"
	"github.com/Kaese72/findings-registry/internal/database"
	"github.com/Kaese72/findings-registry/rest"
	"github.com/spf13/viper"
)

type Config struct {
	Database struct {
		ConnectionString string `mapstructure:"connectionString"`
		Name             string `mapstructure:"name"`
	} `mapstructure:"database"`
	JWT struct {
		Secret string `mapstructure:"secret"`
	} `mapstructure:"jwt"`
	Listen struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	} `mapstructure:"listen"`
	Event struct {
		FindingUpdates   string `mapstructure:"findingUpdates"`
		ConnectionString string `mapstructure:"connectionString"`
	} `mapstructure:"event"`
}

var Loaded Config

func init() {
	// Load configuration from environment
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.BindEnv("database.connectionString")
	viper.BindEnv("database.name")
	viper.SetDefault("database.name", "assetsregistry")

	// JWT configuration
	viper.BindEnv("jwt.secret")

	// HTTP listen config
	viper.BindEnv("listen.host")
	viper.SetDefault("listen.host", "0.0.0.0")
	viper.BindEnv("listen.port")
	viper.SetDefault("listen.port", "8080")
	viper.BindEnv("event.findingUpdates")
	viper.SetDefault("event.findingUpdates", "findingUpdates")
	viper.BindEnv("event.connectionString")

	err := viper.Unmarshal(&Loaded)
	if err != nil {
		log.Fatal(err.Error())
	}
	if Loaded.Database.ConnectionString == "" {
		log.Fatal("Database connection string not set")
	}
	if Loaded.JWT.Secret == "" {
		log.Fatal("JWT secret key not set")
	}
	if Loaded.Event.ConnectionString == "" {
		log.Fatal("Event connection string not set")
	}
}

func main() {
	db, err := database.NewMongoFindingsPersistence(database.MongoDBConfig{
		ConnectionString: Loaded.Database.ConnectionString,
		DbName:           Loaded.Database.Name,
	})
	if err != nil {
		panic(err)
	}
	// if err := db.Purge(); err != nil {
	// 	panic(err)
	// }
	updateChannel, err := event.Setup(Loaded.Event.ConnectionString, Loaded.Event.FindingUpdates)
	if err != nil {
		panic(err)
	}
	router := rest.InitMux(application.NewApplicationLogic(db, updateChannel))
	http.ListenAndServe(fmt.Sprintf("%s:%d", Loaded.Listen.Host, Loaded.Listen.Port), router)
}
