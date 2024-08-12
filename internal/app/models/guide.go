package models

// Пакет с моделями, представляющими собой способ хранения сущности, используемыми в нашей БД

type Guide struct {
	IATA    string `json:"iata"`
	Title   string `json:"title"`
	Author  string `json:"author"`
	Content string `json:"content"`
}
