package cache

import (
	"restapi/internal/app/models"
	"testing"
	"time"
)

// Тест для конструктора Кэша
func TestNewCache(t *testing.T) {
	cache := NewCache()
	if cache == nil {
		t.Error("Тест провален, cache не может иметь значение nil")
	}
}

///////////////////////////////////////////////////////////////////////////////

// Тест для добавления элемента в Кэш по ключу
type TestCaseCacheSet struct {
	key   string
	guide *models.Guide
}

var TestCaseSet = TestCaseCacheSet{
	key: "RO",
	guide: &models.Guide{
		IATA:    "RO",
		Title:   "AAA",
		Author:  "OOO",
		Content: "SSS",
	},
}

func TestCacheSet(t *testing.T) {
	// Создаем Кэш
	cache := NewCache()
	// Добавляем элемент в Кэш и сразу проверяем есть ли он там
	cache.Set(TestCaseSet.key, TestCaseSet.guide)
	var ok bool
	// Если добавленный нами элемент не найден, тест провален
	_, ok = cache.item[TestCaseSet.key]
	if !ok {
		t.Errorf("Тест на добавление элемента в кэш провален, значение %v по ключу %v не найдено", TestCaseSet.guide, TestCaseSet.key)
	}
	if ok {
		if *cache.item[TestCaseSet.key].Value != *TestCaseSet.guide {
			t.Errorf("Тест на добавление элемента в кэш провален, значение %v по ключу %v не соответствует ожидаемому значению %v", TestCaseSet.guide, TestCaseSet.key, TestCaseSet.guide)
		}

	}
}

/////////////////////////////////////////////////////////////////////////////////////

// Тест для получения гайда из Кеша по ключу
type TestCaseCacheGet struct {
	nameTest string
	key      string
	guide    *models.Guide
}

var TestCaseGet = TestCaseCacheGet{
	nameTest: "First test",
	key:      "RO",
	guide: &models.Guide{
		IATA:    "RO",
		Title:   "AAA",
		Author:  "OOO",
		Content: "EEE",
	},
}

func TestCacheGet(t *testing.T) {
	// Создаем Кэш
	cache := NewCache()
	// Добавляем элемент в Кэш
	cache.Set(TestCaseGet.key, TestCaseGet.guide)
	// И тут же проверка на его наличие
	_, ok := cache.Get(TestCaseGet.key)
	// Если был найден элемент, чье время жизни истекло, тест провален
	if ok {
		if time.Now().UnixNano() > cache.item[TestCaseGet.key].Expiration {
			t.Error("Тест на получение элемента из кэша провален, элемент не мог быть найден, т.к. его время жизни истекло")
		}
	}
	// Если элемент не был найден, тест провален
	if !ok {
		if time.Now().UnixNano() < cache.item[TestCaseGet.key].Expiration {
			t.Errorf("Тест на получение элемента из кэша провален, т.к. элемент %v не мог быть не найден по ключу %s, ибо его время еще не вышло", TestCaseGet.guide, TestCaseGet.key)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////

// Тест для инвалидации кэша
type TestCaseCacheDelete struct {
	key   string
	guide *models.Guide
}

var TestCaseDelete = TestCaseCacheDelete{
	key: "RO",
	guide: &models.Guide{
		IATA:    "RO",
		Title:   "AAA",
		Author:  "OOO",
		Content: "EEE",
	},
}

func TestCacheDelete(t *testing.T) {
	// Создаем Кэш
	cache := NewCache()
	// Добавляем элемент
	cache.Set(TestCaseDelete.key, TestCaseDelete.guide)
	// Инвалидируем Кэш
	cache = cache.Delete()
	if len(cache.item) > 0 {
		t.Error("Тест по инвалидации кэша провален, т.к. после этого он не должен содержать элементов")
	}
	if cache == nil {
		t.Error("Тест по инвалидации кэша провален, т.к. после этого он не может иметь значение nil")
	}
}

/////////////////////////////////////////////////////////////////////////////////

// Тест для удаления просроченных элементов по ключам
type TestCaseCacheExpiredItemsClear struct {
	nameTest string
	key      string
	guide    *models.Guide
}

var sliceTestCasesExpiredItemsClear = []TestCaseCacheExpiredItemsClear{
	{
		nameTest: "First test",
		key:      "RO",
		guide: &models.Guide{
			IATA:    "RO",
			Title:   "AAA",
			Author:  "OOO",
			Content: "SSS",
		},
	},
	{
		nameTest: "Second test",
		key:      "FR",
		guide: &models.Guide{
			IATA:    "FR",
			Title:   "OOO",
			Author:  "AAA",
			Content: "SSS",
		},
	},
}

func TestCacheExpiredItemsClear(t *testing.T) {
	// Создаем кэш
	cache := NewCache()
	// Создаем слайс для наших ключей
	var slice = make([]string, 0, 2)
	// Проходимся по слайсу тестовых вариантов добавляя их в кэш и в срез
	for _, test := range sliceTestCasesExpiredItemsClear {
		cache.Set(test.key, test.guide)
		slice = append(slice, test.key)
	}
	// Вызываем метод очистки Кэша по устаревшим ключам (в данном случае предполагаем что они устаревшие)
	cache.clearExpiredItems(slice)
	// Если длина кэша больше 0
	if len(cache.item) > 0 {
		t.Error("Тест по удалению устаревших элементов из кэша провален, после этого он не должен содержать элементов")
	}
	if cache == nil {
		t.Errorf("Тест по удалению устаревших элементов из кэша провален, после этого он не может иметь значение nil")
	}
}

/////////////////////////////////////////////////////////////////////////////////////

// Тест для поиска просроченных ключей
func TestCacheExpiredKeys(t *testing.T) {
	// Создаем Кэш
	cache := NewCache()
	// Создаем три элемента для добавления в Кэш (2 из них будут просроченны (1 и 3))
	firstElem := ItemCache{
		Value: &models.Guide{
			IATA:    "RO",
			Title:   "AAA",
			Author:  "OOO",
			Content: "SSS",
		},
		Created:    time.Now().Add(-2 * time.Hour),
		Expiration: time.Now().Add(-1 * time.Hour).UnixNano(),
	}
	secondElem := ItemCache{
		Value: &models.Guide{
			IATA:    "FR",
			Title:   "AAA",
			Author:  "OOO",
			Content: "SSS",
		},
		Created:    time.Now(),
		Expiration: time.Now().Add(1 * time.Hour).UnixNano(),
	}
	thirdElem := ItemCache{
		Value: &models.Guide{
			IATA:    "GER",
			Title:   "AAA",
			Author:  "OOO",
			Content: "SSS",
		},
		Created:    time.Now().Add(-2 * time.Hour),
		Expiration: time.Now().Add(-1 * time.Hour).UnixNano(),
	}

	// Добавляем элементы в кэш
	cache.item[firstElem.Value.IATA] = firstElem
	cache.item[secondElem.Value.IATA] = secondElem
	cache.item[thirdElem.Value.IATA] = thirdElem

	// Вызываем метод для поиска просроченных ключей
	slice := cache.expiredKeys()
	// Проходимся по слайсу ключей и проверяем что были добавлены только просроченные
	for _, val := range slice {
		if val == secondElem.Value.IATA {
			t.Error("Тест по поиску устаревших ключей провален, т.к. второй элемент не мог быть найден (просроченное время мы установили только 1 и 3)")
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////

// Тест для удаления просроченных элементов Кэша (метод GC)
// Методы внутри GC смысла тестировать уже нет, т.к. для них тесты написаны выше (и мы уже убеждены что они работают правильно)
func TestGC(t *testing.T) {
	// Создаем Кэш
	cache := NewCache()
	// Устанавливаем длительнось очистки Кэша 0, чтобы не ждать исполнения функции
	cache.cleanupInterval = time.Duration(0)
	// Устанавливаем nil значение для мапы в Кэше, чтобы функция завершилась
	cache.item = nil
	// Ну а если функция завершилась во время тестирования сразу, значит она работает правильно
	cache.GC()
	// А если по итогу ее выполнения длительность очистки Кэша не стала 30 минут, значит она работает неправильно
	if cache.cleanupInterval != 30*time.Minute {
		t.Error("Тест по удалению устаревших элементов кэша провален, т.к. cleanupInterval кэша после этого должен был стать больше 0")
	}
}
