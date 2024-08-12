package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"restapi/internal/app/api"
	"restapi/internal/app/config"
	"syscall"
)

// Переменная, которая будет указывать путь до файла с конфигурацией
var configPathAPI string

// Создание флага для пользователя, сигнализирующего о том, что нужен файл с конфигурацией
func init() {
	flag.StringVar(&configPathAPI, "path", "configs/api.toml", "path to config file in .toml format")
}

func main() {
	// Парсим флаг
	flag.Parse()

	// Создаем конфиги и парсим в них значения из конфигурационного файла
	apiConfig, err := config.AllApiConfig(&configPathAPI)
	if err != nil {
		fmt.Println(err)
	}

	// Создаем сервер
	server := api.New(apiConfig)

	// Стартуем сервер
	// Делаем это в отдельной горутине для возможности отловить сигналы завершения приложения далее
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка: %s\n", err)
		}
	}()

	// Создаем канал для отлавливания сигнала ОС или пользователя
	quit := make(chan os.Signal, 1)

	// Начинаем прослушку сигнала
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 1 Версия завершения сервера основанная на таймауте (на мой взгляд не самая эффективная, потому что не факт что обработчики http запросов успеют отработать за отведенное время)
	// Как только сигнал получен начинаем плавное завершение сервера
	// log.Println("Завершение работы сервера...")
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	//
	// Если сервер не успел завершиться за отведенное в контексте время
	// if err := server.GetSrv().Shutdown(ctx); err != nil {
	//	log.Fatal("Сервер завершил работу неккоректно:", err)
	// }
	// Считываем канал и завершаем работу сервера
	// <-ctx.Done()

	// 2 версия завершения сервера основанная на атомарном счетчике (куда практичнее, потому что ведется учет выполнения обработчиков http запросов)
	// Как только сигнал получен начинаем плавное завершение сервера
	log.Println("Завершение работы сервера...")
	ctx := context.Background()
	// Если при завершении работы сервера произошла ошибка
	if err := server.GetSrv().Shutdown(ctx); err != nil {
		log.Fatal("Сервер завершил работу неккоректно:", err)
	}
	server.Wait()
	log.Println("Время вышло, сервер завершил работу успешно")
}
