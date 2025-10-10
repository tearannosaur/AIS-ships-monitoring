package main

import (
	"ais/database"
	"ais/handler"
	s "ais/server_html"
	"fmt"
	"log"
)

func main() {
	//инициализация бд и завершение ее работы по завершении работы программы
	err := database.DbInit()
	if err != nil {
		log.Fatalln("Ошибка подключения к базе данных", err)
		database.AllErrors = append(database.AllErrors, database.MsgErrors{
			Err:    err.Error(),
			ErrMsg: "Ошибка подключения к базе данных",
		})
	}
	defer database.DdClose()
	fmt.Println("Успешное подключение к базе данных")
	//выгрузка всех данных из бд
	err = database.DbDownload()
	if err != nil {
		fmt.Println("Ошибка выгрузки базы данных")
		database.AllErrors = append(database.AllErrors, database.MsgErrors{
			Err:    err.Error(),
			ErrMsg: "Ошибка выгрузки базы данных",
		})

	}
	//запуск сервера параллельно дальнейшей работе программы
	go s.ServerRequest()
	fmt.Println("Успешное подключение к серверу")
	//цикл работы программы
	//по завершении сканирования папки с файлами снова выбрать другую папку и  продолжить сканирование
	for {
		//ожидание выбора пути и папки сканирования в канале
		//продолжит работу после указания пути на главной странице
		handler.NewFolderPause = true
		handler.ResumeScanPause = true
		handler.AutoScanPause = true
		path := <-s.PathChan
		fmt.Println("Получен путь")
		//вызываем обработчик файлов
		err := handler.FileWalk(path)
		if err != nil {
			fmt.Println("Ошибка сканирования", err)
			database.AllErrors = append(database.AllErrors, database.MsgErrors{
				Err:    err.Error(),
				ErrMsg: "Ошибка сканирования",
			})
		}
	}
}
