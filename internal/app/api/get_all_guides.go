package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Получение всех гайдов (тут особо нет смысла считывать с кэша, т.к. нужны все элементы)
func (a *API) GetAllGuides(c *gin.Context) {
	// Каждый обработчик работает с атомарным счетчиком, по которому мы отследим что все они завершились
	a.Add(1)
	defer a.Done()
	a.logger.Info("User do 'GET: GetAllGuides /guides'")

	// Получаем все гайды
	guides, err := a.storage.Guide().GetAllGuides()
	if err != nil {
		a.logger.Error("Trouble with connecting to DB(table guides):", " ", err)
		Message := RequestMessageGuides{
			Message: "Извините, возникли проблемы с доступом к базе данных",
		}
		c.JSON(http.StatusInternalServerError, Message)
		return
	}

	// Если все прошло успешно
	if len(guides) == 0 {
		Message := RequestMessageGuides{
			Message: "В данный момент в базе данных нет гайдов",
		}
		c.JSON(http.StatusOK, Message)

		a.logger.Info("Request 'GET: GetAllGuides /guides' successfully done")
	} else {
		Message := RequestMessageGuides{
			Message: "Гайды успешно найдены",
			Guides:  guides,
		}
		c.JSON(http.StatusOK, Message)

		a.logger.Info("Request 'GET: GetAllGuides /guides' successfully done")
	}
}
