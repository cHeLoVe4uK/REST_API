package api

import (
	"log"
	"log/slog"
	"net/http"
	"restapi/internal/app/cache"
	"restapi/internal/app/config"
	"restapi/internal/app/redis"
	"restapi/storage"
	"sync"

	"github.com/ekomobile/dadata/v2/api/suggest"
	"github.com/gin-gonic/gin"
)

// Инстанс нашего сервера
type API struct {
	// Поля неэкспортируемые (конфендициальная информация)
	server         *http.Server       // поле - проводник между нашим апи и встроенным сервером в пакете http, сделано это, чтобы реализовать gracefull shutdown
	config         *config.Config     // конфиг, который будет использоваться для настройки сервера
	logger         *slog.Logger       // логер который будет использоваться в процессе работы сервера
	router         *gin.Engine        // роутер который будет использоваться в процессе работы сервера (в нашем случае используем фреймворк gin)
	storage        *storage.Storage   // БД, которая будет использоваться в процессе работы сервера
	countryAPI     *suggest.Api       // внешний API, который будет использоваться для получения данных о странах
	cache          *cache.Cache       // внутренний кэш нашего приложения
	redisClient    *redis.RedisClient // redis кэш для нашего приложения
	sync.WaitGroup                    // поле для контроля обработчиков http запросов, чтобы точно знать об их завершении (поможет для плавного завершения работы сервера)
}

// Конструктор, возвращающий инстанс нашего сервера
func New(config *config.Config) *API {
	return &API{
		config: config,
	}
}

// Метод для получения нашего проводника
func (api *API) GetServer() *http.Server {
	return api.server
}

// Метод, стартующий наш сервер
func (api *API) Start() error {
	// Настройка поля логгер
	err := api.configureLoggerField()
	if err != nil {
		return err
	}
	api.logger.Info("Logger succsessfully configured")

	// Настройка поля с хранилищем
	err = api.configureStorageField()
	if err != nil {
		return err
	}
	api.logger.Info("DB connection succsessfully installed")

	// Настройка поля роутер
	api.configureRouterField()
	api.logger.Info("Маршрутизатор готов к работе")

	// Настройка поля стороннего API
	api.configureCountryAPIField()
	api.logger.Info("External API succsessfully configured")

	// Настройка поля server
	api.configureServerField()
	api.logger.Info("Server for gracefull shutdown succsessfully configured")

	// Настройка Кэша для приложения
	api.configureCacheField()
	api.logger.Info("InMemoryCash succsessfully created")

	// Настройка Редис кэша для приложения
	api.configureRedisClientField()
	api.logger.Info("Redis connection succsessfully installed")

	// Запуск очистки кэша по интервалам
	go api.cache.StartGC()
	api.logger.Info("Gorutine for removing cache's outdated items succsessfully started")

	// Сигнал о том, что настройка прошла успешно и запуск нашего сервера
	api.logger.Info("Starting on port:", " ", api.config.BindAddr)

	// Запускаем сервер в отдельной горутине для возможности отловить сигналы завершения приложения
	go func() {
		if err := api.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("While server is running: %s\n", err)
		}
	}()

	return nil
}
