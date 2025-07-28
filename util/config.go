package util

import "github.com/spf13/viper"

type Config struct {
	DBDriver      string `mapstructure:"DB_DRIVER"`
	DBSource      string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
	APIPort       string `mapstructure:"API_PORT"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	viper.SetDefault("DB_DRIVER", "postgres")
	viper.SetDefault("DB_SOURCE", "postgresql://root:secret@localhost:5432/fiber_cms_test?sslmode=disable")
	viper.SetDefault("SERVER_ADDRESS", ":8080")
	viper.SetDefault("API_PORT", ":8080")

	if err = viper.ReadInConfig(); err != nil {

		if _, ok := err.(viper.ConfigFileNotFoundError); ok {

		} else {
			return
		}
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return
	}

	return
}
