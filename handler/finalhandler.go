package handler

import (
	"ais/database"
	"fmt"
	"log"
	"strconv"
)

type Coordinates struct {
	Longitude    float64
	Latitude     float64
	LonDirection string
	LatDirection string
}

type DataMessage struct {
	MessageStr  string  `json:"Message"`
	Date        string  `json:"Date"`
	Time        string  `json:"Time"`
	MessageType int64   `json:"MessageType"`
	Mmsi        int64   `json:"Mmsi"`
	Longitude   float64 `json:"Longitude"`
	Latitude    float64 `json:"Latitude"`
}

var Data DataMessage
var AllMessageinFile []DataMessage

func BitsToInfo(code string, date, time string) {
	Data = DataMessage{}
	if !SafeParse(code, 0, 6) {
		fmt.Println("неккоректный тип сообщения")
		database.AllErrors = append(database.AllErrors, database.MsgErrors{
			Err:    "ошибка перевода типа сообщения",
			ErrMsg: code,
		})
		return
	}
	messageType, err := strconv.ParseInt(code[:6], 2, 64)
	if err != nil {
		return
	}
	if !SafeParse(code, 8, 38) {
		fmt.Println("ошибка перевода строки типа судна")
		database.AllErrors = append(database.AllErrors, database.MsgErrors{
			Err:    "ошибка перевода строки типа судна",
			ErrMsg: code,
		})
		return
	}
	mmsi, err := strconv.ParseInt(code[8:38], 2, 64)
	if err != nil {
		fmt.Println("Ошибка перевода строки mmsi судна")
		database.AllErrors = append(database.AllErrors, database.MsgErrors{
			Err:    "ошибка перевода строки mmsi судна",
			ErrMsg: code,
		})
		return
	}
	database.ObjMMSI = mmsi
	switch messageType {
	case 1, 2, 3:
		coord, err := LatLongParse(code, 61, 89, 89, 116)
		if err != nil {
			fmt.Println(err)
			return
		}
		//fmt.Println("Дата и время:", date, time, "Тип Сообщения:", messageType, "MMSI судна:", mmsi, "Долгота", Check(coord.Longitude), coord.LonDirection, "Широта:", Check(coord.Latitude), coord.LatDirection)
		dataTake(date, time, messageType, mmsi, Check(coord.Longitude), Check(coord.Latitude))
		database.TakeCoordFromDB(coord.Latitude, coord.Longitude)

	case 18:
		coord, err := LatLongParse(code, 57, 85, 85, 112)
		if err != nil {
			fmt.Println(err)
			database.AllErrors = append(database.AllErrors, database.MsgErrors{
				Err:    err.Error(),
				ErrMsg: code,
			})
			return
		}
		//fmt.Println("Дата и время:", date, time, "Тип Сообщения:", messageType, "MMSI судна:", mmsi, "Долгота", Check(coord.Longitude), coord.LonDirection, "Широта:", Check(coord.Latitude), coord.LatDirection)
		dataTake(date, time, messageType, mmsi, Check(coord.Longitude), Check(coord.Latitude))
		database.TakeCoordFromDB(coord.Latitude, coord.Longitude)

	case 19:
		coord, err := LatLongParse(code, 61, 89, 89, 116)
		if err != nil {
			fmt.Println(err)
			database.AllErrors = append(database.AllErrors, database.MsgErrors{
				Err:    err.Error(),
				ErrMsg: code,
			})
			return
		}
		//fmt.Println("Дата и время:", date, time, "Тип Сообщения:", messageType, "MMSI судна:", mmsi, "Долгота", Check(coord.Longitude), coord.LonDirection, "Широта:", Check(coord.Latitude), coord.LatDirection)
		dataTake(date, time, messageType, mmsi, Check(coord.Longitude), Check(coord.Latitude))
		database.TakeCoordFromDB(coord.Latitude, coord.Longitude)

	case 4:
		coord, err := LatLongParse(code, 79, 107, 89, 116)
		if err != nil {
			fmt.Println(err)
			database.AllErrors = append(database.AllErrors, database.MsgErrors{
				Err:    err.Error(),
				ErrMsg: code,
			})
			return
		}
		//fmt.Println("Дата и время:", date, time, "Тип Сообщения:", messageType, "MMSI судна:", mmsi, "Долгота", Check(coord.Longitude), coord.LonDirection, "Широта:", Check(coord.Latitude), coord.LatDirection)
		dataTake(date, time, messageType, mmsi, Check(coord.Longitude), Check(coord.Latitude))
		database.TakeCoordFromDB(coord.Latitude, coord.Longitude)

	case 21:
		coord, err := LatLongParse(code, 132, 160, 160, 187)
		if err != nil {
			fmt.Println(err)
			database.AllErrors = append(database.AllErrors, database.MsgErrors{
				Err:    err.Error(),
				ErrMsg: code,
			})
			return
		}
		//fmt.Println("Дата и время:", date, time, "Тип Сообщения:", messageType, "MMSI судна:", mmsi, "Долгота", Check(coord.Longitude), coord.LonDirection, "Широта:", Check(coord.Latitude), coord.LatDirection)
		dataTake(date, time, messageType, mmsi, Check(coord.Longitude), Check(coord.Latitude))
		database.TakeCoordFromDB(coord.Latitude, coord.Longitude)

	case 27:
		coord, err := LatLongParse(code, 41, 62, 62, 79)
		if err != nil {
			fmt.Println(err)
			database.AllErrors = append(database.AllErrors, database.MsgErrors{
				Err:    err.Error(),
				ErrMsg: code,
			})
			return
		}
		//fmt.Println("Дата и время:", date, time, "Тип Сообщения:", messageType, "MMSI судна:", mmsi, "Долгота", Check(coord.Longitude), coord.LonDirection, "Широта:", Check(coord.Latitude), coord.LatDirection)
		dataTake(date, time, messageType, mmsi, Check(coord.Longitude), Check(coord.Latitude))
		database.TakeCoordFromDB(coord.Latitude, coord.Longitude)

	default:
		//fmt.Println("Дата и время:", date, time, "Тип Сообщения:", messageType, "MMSI судна:", mmsi, "Сообщение не поддерживает координаты")

	}
	IsScannig = false
}

func dataTake(date string, time string, messageType int64, mmsi int64, longitude float64, latitude float64) {
	Data = DataMessage{
		MessageStr:  Message,
		Date:        date,
		Time:        time,
		MessageType: messageType,
		Mmsi:        mmsi,
		Latitude:    latitude,
		Longitude:   longitude,
	}
	AllMessageinFile = append(AllMessageinFile, Data)
	database.Date = date
	database.Time = time
}

func Check(val float64) float64 {
	if val < 0 {
		return -val
	}
	return val
}

func LatLongParse(code string, Lonstart, Lonend, Latstart, Latend int) (Coordinates, error) {
	allCoord := Coordinates{}
	allCoord.LonDirection = "E"
	allCoord.LatDirection = "N"
	if len(code) < max(Lonend, Latend) {
		return allCoord, fmt.Errorf("ошибка извлечения бит,короткое сообщение")
	}

	loncode, err := strconv.ParseInt(code[Lonstart:Lonend], 2, 32)
	if err != nil {
		return allCoord, fmt.Errorf("ошибка перевода долготы")
	}
	latcode, err := strconv.ParseInt(code[Latstart:Latend], 2, 32)
	if err != nil {
		return allCoord, fmt.Errorf("ошибка перевод широты")
	}
	if loncode >= (1 << 27) {
		loncode -= (1 << 28)
	}
	if latcode >= (1 << 26) {
		latcode -= (1 << 27)
	}
	allCoord.Longitude = float64(loncode) / 600000.0
	allCoord.Latitude = float64(latcode) / 600000.0
	//fmt.Println("из сообщения до:", allCoord.Longitude, allCoord.Latitude)
	allCoord.Longitude = database.Round(allCoord.Longitude)
	allCoord.Latitude = database.Round(allCoord.Latitude)
	//fmt.Println("из сообщения после:", allCoord.Longitude, allCoord.Latitude)

	if allCoord.Longitude < 0 {
		allCoord.LonDirection = "W"
	}
	if allCoord.Latitude < 0 {
		allCoord.LatDirection = "S"
	}
	return allCoord, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func SafeParse(code string, start, end int) bool {
	if len(code) < end || len(code) < start {
		log.Printf("Недостаточная длина строки битов для расшифровки:%s,начальный бит и конечный бит:%d,%d\n", code, start, end)
		return false
	}
	return true
}
