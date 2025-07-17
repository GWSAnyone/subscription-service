package configs

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Config хранит все настройки приложения
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Logger   LoggerConfig
}

// ServerConfig хранит настройки HTTP-сервера
type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig хранит настройки базы данных
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	MigrationsPath  string
	MigrationsTable string
}

// LoggerConfig хранит настройки логгера
type LoggerConfig struct {
	Level  string
	Format string
}

// LoadConfig загружает конфигурацию из файла и переменных окружения
func LoadConfig(configPath string) (*Config, error) {
	// Загружаем .env файл, если он существует
	if err := godotenv.Load(); err != nil {
		log.Warn().Err(err).Msg("Error loading .env file")
	}

	// Инициализируем Viper для работы с config.yaml
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	if configPath != "" {
		viper.AddConfigPath(configPath)
	}
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Автоматическая подстановка переменных окружения для полей конфигурации
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Установим стандартные значения
	setDefaults()

	// Читаем конфигурационный файл
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		log.Warn().Msg("Config file not found. Using environment variables and defaults")
	} else {
		log.Info().Str("file", viper.ConfigFileUsed()).Msg("Using config file")
	}

	// Парсим конфигурацию
	config := &Config{
		Server: ServerConfig{
			Port:         viper.GetInt("server.port"),
			ReadTimeout:  viper.GetDuration("server.read_timeout"),
			WriteTimeout: viper.GetDuration("server.write_timeout"),
			IdleTimeout:  viper.GetDuration("server.idle_timeout"),
		},
		Database: DatabaseConfig{
			Host:            viper.GetString("database.host"),
			Port:            viper.GetInt("database.port"),
			User:            viper.GetString("database.user"),
			Password:        viper.GetString("database.password"),
			DBName:          viper.GetString("database.dbname"),
			SSLMode:         viper.GetString("database.sslmode"),
			MaxOpenConns:    viper.GetInt("database.max_open_conns"),
			MaxIdleConns:    viper.GetInt("database.max_idle_conns"),
			ConnMaxLifetime: viper.GetDuration("database.conn_max_lifetime"),
			MigrationsPath:  viper.GetString("database.migrations_path"),
			MigrationsTable: viper.GetString("database.migrations_table"),
		},
		Logger: LoggerConfig{
			Level:  viper.GetString("logger.level"),
			Format: viper.GetString("logger.format"),
		},
	}

	// Логируем загруженную конфигурацию
	log.Debug().Interface("config", config).Msg("Loaded configuration")

	return config, nil
}

// DSN возвращает строку подключения к PostgreSQL
func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// ConnectionString возвращает строку подключения для библиотеки миграций
func (c DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode,
	)
}

// setDefaults устанавливает значения по умолчанию
func setDefaults() {
	// Настройки сервера
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "15s")
	viper.SetDefault("server.write_timeout", "15s")
	viper.SetDefault("server.idle_timeout", "60s")

	// Настройки базы данных
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "subscription_service")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.max_open_conns", 20)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "5m")
	viper.SetDefault("database.migrations_path", "./migrations")
	viper.SetDefault("database.migrations_table", "schema_migrations")

	// Настройки логгера
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "json")
}
