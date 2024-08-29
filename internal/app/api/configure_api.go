package api

import (
	"log/slog"
	"net/http"
	"os"
	"restapi/internal/app/cache"
	red "restapi/internal/app/redis"
	"restapi/storage"

	"github.com/ekomobile/dadata/v2"
	"github.com/ekomobile/dadata/v2/client"
	"github.com/gin-gonic/gin"
)

// Конфигурируем поле логгер нашего сервера
func (api *API) configureLoggerField() error {
	level := api.config.LoggerLevel

	logLevel := slog.SetLogLoggerLevel(slog.Level(level))
	opt := slog.HandlerOptions{
		Level: logLevel,
	}

	api.logger = slog.New(slog.NewJSONHandler(os.Stdout, &opt))
	return nil
}

// Конфигурируем роутер нашего сервера
func (api *API) configureRouterField() {
	router := gin.Default()
	guideGroup := router.Group("/guide")
	guideGroup.GET("/:iata", api.GetGuideByIATA)
	guideGroup.POST("/:iata", api.PostGuideByIATA)
	guideGroup.PUT("/:iata", api.PutGuideByIATA)
	guideGroup.DELETE("/:iata", api.DeleteGuideByIATA)

	router.GET("/guides", api.GetAllGuides)

	router.GET("/info/:iata", api.GetCountryInfoByIATA)

	api.router = router
}

// Конфигурируем хранилище нашего сервера
func (api *API) configureStorageField() error {
	storage := storage.New(api.config.Storage)

	err := storage.Open()
	if err != nil {
		return err
	}

	api.storage = storage
	return nil
}

// Конфигурируем сторонний API для получения информации о странах
func (api *API) configureCountryAPIField() {
	externalAPI := api.config.ExternalAPI

	creds := client.Credentials{
		ApiKeyValue:    externalAPI.ApiKeyValue,
		SecretKeyValue: externalAPI.SecretKeyValue,
	}

	ap := dadata.NewSuggestApi(client.WithCredentialProvider(&creds))

	api.countryAPI = ap
}

// Конфигурируем Кэш для нашего сервера
func (api *API) configureCacheField() {
	api.cache = cache.NewCache()
}

// Конфигурируем Редис Кэш для нашего сервера
func (api *API) configureRedisClientField() {
	api.redisClient = red.NewRedisClient(api.config.Redis)
}

// Конфигурируем поле server для нашего сервера
func (api *API) configureServerField() {
	api.server = &http.Server{
		Addr:    ":" + api.config.BindAddr,
		Handler: api.router.Handler(),
	}
}
