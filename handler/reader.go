package handler

import (
	"ais/database"
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

/* считываем файлы начиная с пути root и для каждого файла вызываем анонимную функцию
функция содержит путь до файла информацию о найденом элементе и ошибку если она возникнет
внутри функции мы проверяем является ли файл папкой и есть ли у него расширение .txt
после чего проверяем имя папки получая полный путь до корня и с помощью filepath.Base получаем имя последнего файла
потом открываем файл если имя папки dir_buf_whith_time
открытие файла осуществляется функцией OpenFile,если все считало возвращаем что нет ошибки */

var IsScannig bool = true
var ResumeScan = make(chan struct{})
var StopScan = make(chan struct{})
var AutoScan = make(chan struct{})
var Folder string
var Pause = true
var NewFolderPause bool
var ResumeScanPause bool
var AutoScanPause bool

type CountFiles struct {
	AllFiles  int `json:"AllFiles"`
	FileNow   int `json:"FileNow"`
	FindCount int `json:"FindCount"`
}

var CountFile CountFiles
var messagecount int

func FileWalk(root string) error {
	NewFolderPause = false
	ResumeScanPause = false
	AutoScanPause = false
	messagecount = 0

	count, err := FileCount(root)
	if err != nil {
		return err
	}
	CountFile.AllFiles = count
	fmt.Println("Всего файлов", count)
	CountFile.FileNow = 0
	CountFile.FindCount = 0
	inittime := time.Now()
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".txt") {
			parrentdir := filepath.Base(filepath.Dir(path))
			if parrentdir == Folder { //dir_buf_whith_time
				fmt.Println("Найденно в файле:", len(database.FindObjects))
				if Pause {
					ResumeScanPause = false
					fmt.Println("Ожидание ответа на продолжение сканирования")
				waitLoop:
					for {
						select {
						case <-ResumeScan:
							ResumeScanPause = true
							NewFolderPause = false
							break waitLoop
						case <-StopScan:
							NewFolderPause = true
							AutoScanPause = false
							Pause = true
							return filepath.SkipAll
						case <-AutoScan:
							AutoScanPause = true
							Pause = false
							break waitLoop
						}
					}
				}
				select {
				case <-ResumeScan:
					AutoScanPause = false
					ResumeScanPause = true
					NewFolderPause = false
					Pause = true
				case <-StopScan:
					NewFolderPause = true
					AutoScanPause = false
					return filepath.SkipAll
				default:

					AllMessageinFile = AllMessageinFile[:0]
					database.FindObjects = database.FindObjects[:0]
					database.AllErrors = database.AllErrors[:0]
					fmt.Println("Открываем файл", path, d.Name())

					//time.Sleep(5 * time.Second)
					err := OpenFile(path)
					if err != nil {
						log.Println("Ошибка чтения файла", err)
						database.AllErrors = append(database.AllErrors, database.MsgErrors{
							Err:    err.Error(),
							ErrMsg: "Ошибка чтения файла",
						})

					}
					CountFile.FileNow++
					CountFile.FindCount += len(database.FindObjects)
					fmt.Println("Отсканировано файлов ", CountFile.FileNow)
					fmt.Println("Найдено всего", CountFile.FindCount)
					fmt.Println("Файл отсканирован")
					fmt.Println("Сообщений отсканировано", messagecount)
					fmt.Println("Прошло времени:", time.Since(inittime))
				}

			}
		}
		return nil
	})

	if err != nil {
		return err
	}
	fmt.Println("Сканирование завершено")
	fmt.Println("Прошло времени:", time.Since(inittime))
	return nil
}

func FileCount(root string) (count int, err error) {
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".txt") {
			parrentdir := filepath.Base(filepath.Dir(path))
			if parrentdir == Folder { // dir_buf_whith_time
				count++
			}
		}
		return nil
	})
	return count, err
}

// открываем файл передавая в аргумент путь файла ,после чего создаем сканнер и проходимся по каждой строке файла
// строку разделяем на части дата,время и информация через пробелы
// вызываем функцию обработки строки и кладем в ее аргумент нашу строку
func OpenFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		messagecount++
		var date string
		var timehour string
		var dataline string
		line := scanner.Text()
		memoryWithTime := strings.Fields(line)
		if len(memoryWithTime) == 0 {
			log.Println("Пустая строка")
			database.AllErrors = append(database.AllErrors, database.MsgErrors{
				Err:    "пустая строка",
				ErrMsg: "",
			})
			continue
		}
		if len(memoryWithTime) == 1 {
			date = "-"
			timehour = "-"
			dataline = memoryWithTime[0]
			//fmt.Println(dataline)
			MessageHandler(dataline, date, timehour)
			continue
		}
		if len(memoryWithTime) == 2 {
			date = "-"
			timehour = "-"
			dataline = strings.Join(memoryWithTime[0:], " ")
			MessageHandler(dataline, date, timehour)
			continue
		}
		date = memoryWithTime[0]
		if len(date) != 10 {
			date = "wrong date"
		}
		timehour = memoryWithTime[1]
		if len(timehour) < 7 || len(timehour) > 8 {
			timehour = "wrong time"
		}
		dataline = strings.Join(memoryWithTime[2:], " ")

		MessageHandler(dataline, date, timehour)
	}
	return scanner.Err()
}
