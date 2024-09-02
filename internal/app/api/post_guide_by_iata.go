package api

import (
	"fmt"
	"net/http"
	"restapi/internal/app/models"
	"strings"

	"github.com/gin-gonic/gin"
)

// Добавление гайда по IATA
func (a *API) PostGuideByIATA(c *gin.Context) {
	// Каждый обработчик работает с атомарным счетчиком, по которому мы отследим что все они завершились
	a.Add(1)
	defer a.Done()
	a.logger.Info("User do 'POST: PostGuideByIATA /guide/:iata'")

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
		a.logger.Error("User provided uncorrected JSON")
		Message := RequestMessage{
			Message: "Предоставленные данные имеют неверный формат",
		}
		c.JSON(http.StatusBadRequest, Message)
		return
	}

	// Проверяем что IATA в запросе и IATA в теле запроса не отличаются
	if iataUP != strings.ToUpper(guide.IATA) {
		a.logger.Error("Request IATA and JSON IATA is mismatched")
		Message := RequestMessage{
			Message: "IATA в запросе не соответствует IATA в теле запроса",
		}
		c.JSON(http.StatusBadRequest, Message)
		return
	}

	// Проверяем что страна с таким IATA существует
	suggestions, err := a.GetCountryByIATA(iataUP)
	if err != nil {
		a.logger.Error("An error occurred when creating guide", " ", err)
		Message := RequestMessage{
			Message: "Извините, в данный момент эта функция недоступна. Попробуйте позже",
		}
		c.JSON(http.StatusInternalServerError, Message)
		return
	}

	// Если все прошло хорошо, но никакой информации о стране не было получено
	if len(suggestions) == 0 {
		a.logger.Error("User provided non existed IATA in PostGuideByIATA")
		Message := RequestMessage{
			Message: "Вы указали несуществующий IATA",
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

	// Если гайд был найден
	if ok {
		a.logger.Info("User trying to create existed guide")
		Message := RequestMessage{
			Message: "Вы пытаетесь создать гайд с уже существующим IATA",
		}
		c.JSON(http.StatusBadRequest, Message)
		return
	}

	// Только после всего этого переходим непосредственно к созданию гайда
	guideFinal, err := a.storage.Guide().CreateGuide(&guide)
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
		Message: fmt.Sprintf("Гайд с идентификатором {IATA: %v} успешно добавлен", guideFinal.IATA),
		Guide:   guideFinal,
	}
	c.JSON(http.StatusCreated, Message)

	a.logger.Info("Cache invalidation in PostGuideByIATA")
	a.cache = a.cache.Delete()

	a.logger.Info("Request 'POST: PostGuideByIATA /guide/:iata' successfully done")
}
