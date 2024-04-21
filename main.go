package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Kaese72/finding-registry/event"
	"github.com/Kaese72/finding-registry/internal/application"
	"github.com/Kaese72/finding-registry/internal/database"
	"github.com/Kaese72/finding-registry/rest"
	"github.com/Kaese72/riskie-lib/logging"
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
	viper.SetDefault("database.name", "riskRegistry")

	// JWT configuration
	viper.BindEnv("jwt.secret")

	// HTTP listen config
	viper.BindEnv("listen.host")
	viper.SetDefault("listen.host", "0.0.0.0")
	viper.BindEnv("listen.port")
	viper.SetDefault("listen.port", "8080")

	// Event configuration
	viper.BindEnv("event.findingUpdates")
	viper.SetDefault("event.findingUpdates", "findingUpdates")
	viper.BindEnv("event.connectionString")

	err := viper.Unmarshal(&Loaded)
	if err != nil {
		logging.Fatal(context.Background(), err.Error())
	}
	if Loaded.Database.ConnectionString == "" {
		logging.Fatal(context.Background(), "Database connection string not set")
	}
	if Loaded.JWT.Secret == "" {
		logging.Fatal(context.Background(), "JWT secret key not set")
	}
	if Loaded.Event.ConnectionString == "" {
		logging.Fatal(context.Background(), "Event connection string not set")
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
	router := rest.InitMux(application.NewApplicationLogic(db, updateChannel), Loaded.JWT.Secret)
	http.ListenAndServe(fmt.Sprintf("%s:%d", Loaded.Listen.Host, Loaded.Listen.Port), router)
}
