package api

import (
	"context"
	"restapi/internal/app/models"
	"strconv"

	"github.com/ekomobile/dadata/v2/api/suggest"
)

// Структура для предоставления информации пользователю, где нужен один гайд
type RequestMessage struct {
	Message     string        `json:"message"`
	Guide       *models.Guide `json:"guide,omitempty"`
	CountryInfo string        `json:"country_info,omitempty"`
}

// Структура для предоставления информации пользователю, где нужен срез гайдов
type RequestMessageGuides struct {
	Message string          `json:"message"`
	Guides  []*models.Guide `json:"guides,omitempty"`
}

// Метод (вспомогательный), проверяющий, что пользователь не вводил численные значение IATA (т.к. наше API такого не позволяет)
func (a *API) NotNumb(numb string) (RequestMessage, error) {
	_, err := strconv.Atoi(numb)
	if err == nil {
		a.logger.Error("User try to use number value of IATA")
		return RequestMessage{Message: "Убедитесь в правильности введенного запроса. IATA не может быть числом"}, nil
	}
	return RequestMessage{}, err
}

// Функция для получения информации о стране через наш сторонний API (проверка что IATA введена верно)
func (a *API) GetCountryByIATA(iata string) ([]*suggest.CountrySuggestion, error) {
	suggestions, err := a.countryAPI.CountryByID(context.Background(), iata)
	if err != nil {
		a.logger.Error("Error while trying to receive info about country")
		return nil, err
	}

	return suggestions, nil
}
