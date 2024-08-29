package api

import (
	"fmt"
	"net/http"
	"restapi/internal/app/models"
	"strings"

	"github.com/gin-gonic/gin"
)

// Обновление гайда по IATA
func (a *API) PutGuideByIATA(c *gin.Context) {
	// Каждый обработчик работает с атомарным счетчиком, по которому мы отследим что все они завершились
	a.Add(1)
	defer a.Done()
	a.logger.Info("Пользователь выполняет 'PUT: PutGuideByIATA /guide/:iata'")

	// Считываем значение IATA, проверка что это не число и приведение к прописному виду в случае успеха
	iata := c.Params.ByName("iata")
	message, err := a.NotNumb(iata)
	if err == nil {
		c.JSON(http.StatusBadRequest, message)
		return
	}
	iataUP := strings.ToUpper(iata)

	// Подготавливаем переменную для гайда
	var guide models.Guide

	// Парсим тело запроса
	err = c.ShouldBindJSON(&guide)
	if err != nil {
		a.logger.Info("Пользователь предоставил неккоректный json")
		Message := Message{
			Message: "Предоставленные данные имеют неверный формат",
		}
		c.JSON(http.StatusBadRequest, Message)
		return
	}

	// Проверяем что IATA в запросе и IATA в теле запроса не отличаются
	if iataUP != strings.ToUpper(guide.IATA) {
		a.logger.Info("Данные в запросе не совпадают с данными в json")
		Message := Message{
			Message: "IATA в запросе не соответствует IATA в теле запроса",
		}
		c.JSON(http.StatusBadRequest, Message)
		return
	}

	// Если все прошло хорошо переходим к поиску гайда в БД
	_, ok, err := a.storage.Guide().GetGuideByIATA(iataUP)
	if err != nil {
		a.logger.Info("Проблемы c подключением к БД (таблица guides)")
		Message := Message{
			Message: "Извините, возникли проблемы с доступом к базе данных",
		}
		c.JSON(http.StatusInternalServerError, Message)
		return
	}
	// Если гайд не найден
	if !ok {
		a.logger.Info("Пользователь пытается обновить несуществующий гайд")
		Message := Message{
			Message: "Вы пытаетесь обновить несуществующий гайд",
		}
		c.JSON(http.StatusBadRequest, Message)
		return
	}

	// Только после всего этого переходим непосредственно к обновлению гайда
	guideFinal, err := a.storage.Guide().UpdateGuideByIATA(&guide)
	if err != nil {
		a.logger.Info("Проблемы c подключением к БД (таблица guides)")
		Message := Message{
			Message: "Извините, возникли проблемы с доступом к базе данных",
		}
		c.JSON(http.StatusInternalServerError, Message)
		return
	}

	// Если все прошло хорошо
	guideSlice := make([]*models.Guide, 0, 1)
	Message := Message{
		Message: fmt.Sprintf("Гайд с идентификатором {IATA: %v} успешно обновлен", guideFinal.IATA),
		Guide:   append(guideSlice, guideFinal),
	}
	c.JSON(http.StatusOK, Message)

	a.logger.Info("Инвалидация кэша при обновлении гайда")
	a.cache = a.cache.Delete()

	a.logger.Info("Запрос 'PUT: PutGuideByIATA /guide/:iata' успешно выполнен")
}
