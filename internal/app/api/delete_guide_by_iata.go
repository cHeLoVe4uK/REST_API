package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Удаление гайда по IATA
func (a *API) DeleteGuideByIATA(c *gin.Context) {
	// Каждый обработчик работает с атомарным счетчиком, по которому мы отследим что все они завершились
	a.Add(1)
	defer a.Done()
	a.logger.Info("User do 'DELETE: DeleteGuideByIATA /guide/:iata'")

	// Считываем значение IATA, проверка что это не число и приведение к прописному виду в случае успеха
	iata := c.Params.ByName("iata")
	message, err := a.NotNumb(iata)
	if err == nil {
		c.JSON(http.StatusBadRequest, message)
		return
	}
	iataUP := strings.ToUpper(iata)

	// Ищем гайд в БД
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
		a.logger.Info("User trying to delete non existed guide")
		Message := RequestMessage{
			Message: "Вы пытаетесь удалить несуществующий гайд",
		}
		c.JSON(http.StatusBadRequest, Message)
		return
	}

	// Удаляем гайд
	guideFinal, err := a.storage.Guide().DeleteGuideByIATA(iataUP)
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
		Message: fmt.Sprintf("Гайд с идентификатором {IATA: %v} успешно удален", guideFinal.IATA),
		Guide:   guideFinal,
	}
	c.JSON(http.StatusOK, Message)

	a.logger.Info("Cache invalidation in DeleteGuideByIATA")
	a.cache = a.cache.Delete()

	a.logger.Info("Request 'DELETE: DeleteGuideByIATA /guide/:iata' successfully done")
}
