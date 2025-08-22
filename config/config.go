package config

import (
	"time"

	"github.com/spf13/viper"
)

var LoadedConfig *Config

// Config 应用配置
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Log      LogConfig
	Redis    RedisConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey       string
	ExpirationHours int
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string
	Path   string
	Format string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// Load 加载配置
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// 设置默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// 设置默认配置
func setDefaults() {
	// 服务器默认配置
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.readTimeout", 15)
	viper.SetDefault("server.writeTimeout", 15)
	viper.SetDefault("server.idleTimeout", 60)

	// 数据库默认配置
	viper.SetDefault("database.dsn", "root:password@tcp(localhost:3306)/cms?charset=utf8mb4&parseTime=True&loc=Local")
	viper.SetDefault("database.maxOpenConns", 100)
	viper.SetDefault("database.maxIdleConns", 20)
	viper.SetDefault("database.connMaxLifetime", 300)
	viper.SetDefault("database.connMaxIdleTime", 60)

	// JWT默认配置
	viper.SetDefault("jwt.secretKey", "cms_secret_key")
	viper.SetDefault("jwt.expirationHours", 24)

	// 日志默认配置
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.path", "logs/")
	viper.SetDefault("log.format", "text")

	// Redis默认配置
	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

}
