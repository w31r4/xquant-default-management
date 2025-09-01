package config

import "github.com/spf13/viper"

// Config 存储应用的所有配置
type Config struct {
	ServerPort string `mapstructure:"SERVER_PORT"`
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	JWTSecret  string `mapstructure:"JWT_SECRET"`
	TokenTTL   int    `mapstructure:"TOKEN_TTL"` // in hours
}

// LoadConfig 从文件或环境变量中加载配置
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv() // 允许从环境变量中读取

	err = viper.ReadInConfig()
	if err != nil {
		// 如果只是配置文件不存在，可以忽略
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return
		}
	}

	err = viper.Unmarshal(&config)
	return
}
