package api

import (
	"fmt"
	"net/http"
	"restapi/internal/app/models"
	"strings"

	"github.com/gin-gonic/gin"
)

// Удаление гайда по IATA
func (a *API) DeleteGuideByIATA(c *gin.Context) {
	// Каждый обработчик работает с атомарным счетчиком, по которому мы отследим что все они завершились
	a.Add(1)
	defer a.Done()
	a.logger.Info("Пользователь выполняет 'DELETE: DeleteGuideByIATA /guide/:iata'")

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
		a.logger.Info("Проблемы c подключением к БД (таблица guides)")
		Message := Message{
			Message: "Извините, возникли проблемы с доступом к базе данных",
		}
		c.JSON(http.StatusInternalServerError, Message)
		return
	}
	// Если гайд не найден
	if !ok {
		a.logger.Info("Пользователь пытается удалить несуществующий гайд")
		Message := Message{
			Message: "Вы пытаетесь удалить несуществующий гайд",
		}
		c.JSON(http.StatusBadRequest, Message)
		return
	}

	// Удаляем гайд
	guideFinal, err := a.storage.Guide().DeleteGuideByIATA(iataUP)
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
		Message: fmt.Sprintf("Гайд с идентификатором {IATA: %v} успешно удален", guideFinal.IATA),
		Guide:   append(guideSlice, guideFinal),
	}
	c.JSON(http.StatusOK, Message)

	a.logger.Info("Инвалидация кэша при удалении гайда")
	a.cache = a.cache.Delete()

	a.logger.Info("Запрос 'DELETE: DeleteGuideByIATA /guide/:iata' успешно выполнен")
}
