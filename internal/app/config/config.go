package config

import (
	"github.com/BurntSushi/toml"
)

// Пакет с конфигами для нашего приложения

// Конфиг для сервера
type Config struct {
	BindAddr    string `toml:"bind_addr"`
	LoggerLevel int    `toml:"logger_level"`
	Storage     *StorageConfig
	ExternalAPI *ExternalConfigAPI
	Redis       *RedisConfig
}

// Конструктор конфига для сервера
func NewConfigAPI(storage *StorageConfig, external *ExternalConfigAPI, redis *RedisConfig) *Config {
	return &Config{
		Storage:     storage,
		ExternalAPI: external,
		Redis:       redis,
	}
}

// Конфиг для БД
type StorageConfig struct {
	DataBaseURI string `toml:"database_uri"`
	DriverName  string `toml:"driver_name"`
	TableName   string `toml:"table_name"`
}

// Конструктор конфига для БД
func NewStorageConfig() *StorageConfig {
	return &StorageConfig{}
}

// Конфиг для сторонней API
type ExternalConfigAPI struct {
	ApiKeyValue    string `toml:"api_key"`
	SecretKeyValue string `toml:"secret_key"`
}

// Конструктор конфига для стороннего API
func NewExternalConfigAPI() *ExternalConfigAPI {
	return &ExternalConfigAPI{}
}

type RedisConfig struct {
	Addr     string `json:"redis_addr"`
	Password string `json:"redis_password"`
}

func NewRedisConfig() *RedisConfig {
	return &RedisConfig{}
}

// Функция для настройки конфигов приложения
func AllApiConfig(configPathAPI *string) (*Config, error) {
	// конфиг для БД
	storConfig := NewStorageConfig()
	_, err := toml.DecodeFile(*configPathAPI, storConfig)
	if err != nil {
		return nil, err
	}

	// конфиг для стороннего API
	externalConfigAPI := NewExternalConfigAPI()
	_, err = toml.DecodeFile(*configPathAPI, externalConfigAPI)
	if err != nil {
		return nil, err
	}

	// Конфиг для Редиса
	redisConfig := NewRedisConfig()
	_, err = toml.DecodeFile(*configPathAPI, redisConfig)
	if err != nil {
		return nil, err
	}

	// конфиг для нашего API
	apiConfig := NewConfigAPI(storConfig, externalConfigAPI, redisConfig)
	_, err = toml.DecodeFile(*configPathAPI, apiConfig)
	if err != nil {
		return nil, err
	}

	return apiConfig, nil
}
