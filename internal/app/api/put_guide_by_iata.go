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
	a.logger.Info("User do 'PUT: PutGuideByIATA /guide/:iata'")

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
		a.logger.Info("User provided uncorrected JSON")
		Message := RequestMessage{
			Message: "Предоставленные данные имеют неверный формат",
		}
		c.JSON(http.StatusBadRequest, Message)
		return
	}

	// Проверяем что IATA в запросе и IATA в теле запроса не отличаются
	if iataUP != strings.ToUpper(guide.IATA) {
		a.logger.Info("Request IATA and JSON IATA is mismatched")
		Message := RequestMessage{
			Message: "IATA в запросе не соответствует IATA в теле запроса",
		}
		c.JSON(http.StatusBadRequest, Message)
		return
	}

	// Если все прошло хорошо переходим к поиску гайда в БД
	_, ok, err := a.storage.Guide().GetGuideByIATA(iataUP)
	if err != nil {
		a.logger.Error("Trouble with connecting to DB (table guides):", " ", err)
		Message := RequestMessage{
			Message: "Извините, возникли проблемы с доступом к базе данных",
		}
		c.JSON(http.StatusInternalServerError, Message)
		return
	}
	// Если гайд не найден
	if !ok {
		a.logger.Info("User trying to update non existed guide")
		Message := RequestMessage{
			Message: "Вы пытаетесь обновить несуществующий гайд",
		}
		c.JSON(http.StatusBadRequest, Message)
		return
	}

	// Только после всего этого переходим непосредственно к обновлению гайда
	guideFinal, err := a.storage.Guide().UpdateGuideByIATA(&guide)
	if err != nil {
		a.logger.Error("Trouble with connecting to DB (table guides):", " ", err)
		Message := RequestMessage{
			Message: "Извините, возникли проблемы с доступом к базе данных",
		}
		c.JSON(http.StatusInternalServerError, Message)
		return
	}

	// Если все прошло хорошо
	Message := RequestMessage{
		Message: fmt.Sprintf("Гайд с идентификатором {IATA: %v} успешно обновлен", guideFinal.IATA),
		Guide:   guideFinal,
	}
	c.JSON(http.StatusOK, Message)

	a.logger.Info("Cache invalidation in PutGuideByIATA")
	a.cache = a.cache.Delete()

	a.logger.Info("Request 'PUT: PutGuideByIATA /guide/:iata' successfully done")
}
