package handler

import (
	"ais/database"
	"fmt"
	"strconv"
	"strings"
)

var memory string
var Message string

//принимаем строку и фльтруем на нужное сообщение
//разделяем все сообщение на части между запятыми

func MessageHandler(line string, date, time string) {
	if strings.Contains(line, "!AIVDM") {
		//сделать обработку ошибок этих функций
		cleanString, err := ClearBeforeLine(line)
		if err != nil {
			//log.Println(err, line)
			database.AllErrors = append(database.AllErrors, database.MsgErrors{
				Err:    err.Error(),
				ErrMsg: line,
			})
			return
		}
		for _, oneOfMessage := range cleanString {
			//	fmt.Println("Сообщение из слайса:", oneOfMessage)
			dataString, err := CleanAfterString(oneOfMessage)
			//	fmt.Println("Сообщение после очищения второй части:", dataString)
			if err != nil {

				//log.Println(err, oneOfMessage)
				database.AllErrors = append(database.AllErrors, database.MsgErrors{
					Err:    err.Error(),
					ErrMsg: oneOfMessage,
				})
				continue
			}
			//тут чек сумму проверяем
			err = CheckSumm(dataString)
			if err != nil {
				//fmt.Println(err, dataString)
				database.AllErrors = append(database.AllErrors, database.MsgErrors{
					Err:    err.Error(),
					ErrMsg: dataString,
				})
				continue
			}
			Message = dataString
			//fmt.Println("Дата и время:", date, time, dataString)
			newlineparts := strings.Split(dataString, ",")
			//если вдруг неккоректное сообщение обрабатываем его
			//	fmt.Println("Длина:", len(newlineparts))
			if len(newlineparts) < 7 { //исправить если что на 6
				//fmt.Println("Не хватает частей через запятую", line)
				database.AllErrors = append(database.AllErrors, database.MsgErrors{
					Err:    "не хватает частей через запятую",
					ErrMsg: line,
				})
				continue
			}
			//если у нас многострочное сообщение нужно его собрать из двух
			//сделать мапой если буду использовать асинхронный способ перебора сообщений

			if newlineparts[1] != "1" {
				//fmt.Println("Многострочное сообщение")

				//если это первая часть сообщения записываем ее в память
				//fmt.Println("Память:", memory)
				if newlineparts[1] == "2" && newlineparts[2] == "1" {
					if memory != "" {
						//	fmt.Println("Предыдущая часть 2 не получена. Сброс памяти.")
						memory = ""
					}
					memory += newlineparts[5]
					continue
				}
				//если это вторая часть сообщения записываем ее в память
				//обрабатываем полное сообщение и очищаем память
				if newlineparts[1] == "2" && newlineparts[2] == "2" {
					if memory == "" {
						//fmt.Println("Второе сообщение без первой части")
						continue
					}
					//исключить если первое сообщение не найдено
					memory += newlineparts[5]
					ConvertToAisCode(memory, date, time)
					//fmt.Println("Память после второго сообщения:", memory)
					memory = ""
					continue
				}
			}
			//если не многострочное сразу вызываем обработчик
			ConvertToAisCode(newlineparts[5], date, time)

		}

	}

}

func CheckSumm(data string) error {
	data = data[1:]
	datafileds := strings.Split(data, "*")
	if len(datafileds) != 2 {
		return fmt.Errorf("некоректный формат сообщения")
	}
	message := datafileds[0]
	checksumm, err := strconv.ParseUint(datafileds[1], 16, 8)
	if err != nil {
		return fmt.Errorf("ошибка перевода чексуммы в uint")
	}
	checksumcalculate := byte(0)
	for _, val := range message {
		checksumcalculate ^= byte(val)
	}
	if byte(checksumm) != checksumcalculate {
		return fmt.Errorf("неверная чек сумма сообщения")
	}
	return nil
}

// Очищаем строку до нужного сообщения

func ClearBeforeLine(line string) ([]string, error) {
	splitline := strings.Split(line, "!AIVDM")

	var dataString []string
	for _, part := range splitline {
		if part == "" {
			continue
		}
		fullLine := "!AIVDM" + part

		if strings.HasPrefix(fullLine, "!AIVDM,") {
			dataString = append(dataString, fullLine)
		}
	}
	if len(dataString) == 0 {
		return nil, fmt.Errorf("не найденого валидного !AIVDM сообщения")
	}
	//	fmt.Println("Сообщения после разделения:", dataString)
	return dataString, nil

}

func CleanAfterString(line string) (string, error) {
	//	fmt.Println("После пришло", line)
	if !strings.Contains(line, "*") {
		return line, fmt.Errorf("некорректная чек сумма")
	}
	splitLine := strings.Split(line, "*")
	//	fmt.Println(splitLine)

	checkSumm := splitLine[1]
	if len(checkSumm) < 2 {
		return checkSumm, fmt.Errorf("некорректная чек сумма")
	}
	fullCLearLine := splitLine[0] + "*" + string(checkSumm[:2])
	return fullCLearLine, nil
}

// принимаем строку и проходимся по ее рунам вычитая 48,если число >=40 то вычитаем еще 8
// тк у нас аис символы от 0 до w в аски таблице 0-48
// добавляем все значения в массив и передаем его в функцию ConvertToBits
func ConvertToAisCode(part string, date, time string) {
	code := make([]int, 0, 10)
	for _, val := range part {
		num := int(val) - 48
		if num >= 40 {
			num -= 8
		}
		code = append(code, num)
	}

	ConvertToBits(code, date, time)
}

//тут собираем битовую строку из массива чисел и записываем  в result

func ConvertToBits(code []int, date, time string) {
	var builder strings.Builder
	for _, val := range code {
		builder.WriteString(fmt.Sprintf("%06b", val))
	}
	result := builder.String()
	BitsToInfo(result, date, time)
}
