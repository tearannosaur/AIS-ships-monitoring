package database

import (
	"fmt"
	"log"
	"math"
	"time"
)

type Segment struct {
	KilometerId      int     `db:"kilometer_id"`
	City             string  `db:"city"`
	Depth            float64 `db:"depth"`
	SegmentLatitude  float64 `db:"segment_latitude"`
	SegmentLongitude float64 `db:"segment_longitude"`
}

type ObjectInfo struct {
	Id              int       `db:"id" json:"Id"`
	ObjectId        int       `db:"object_id" json:"ObjectId"`
	SegmendKmId     int       `db:"segment_km_id" json:"SegmentId"`
	ObjectName      string    `db:"object_name" json:"ObjectName"`
	ObjectLatitude  float64   `db:"object_latitude" json:"ObjectLatitude"`
	ObjectLongitude float64   `db:"object_longitude" json:"ObjectLongitude"`
	CoordRadius     float64   `db:"coord_radius" json:"ObjectRadius"`
	Description     string    `db:"descriptions" json:"ObjectDescription"`
	OMMSI           int64     `db:"mmsi" json:"ObjectMMSI"`
	DateDB          time.Time `db:"date_db" json:"DateDB"`
}

type MsgErrors struct {
	Err    string `json:"Err"`
	ErrMsg string `json:"ErrMsg"`
}

var AllErrors []MsgErrors

var ObjMMSI int64

type DataBase struct {
	Obj   map[int][]ObjectInfo
	Sgmnt map[int][]Segment
}

var NewSegment []Segment
var NewObject []ObjectInfo
var Data DataBase

func DbDownload() error {
	data := DataBase{}
	data.Obj = make(map[int][]ObjectInfo)
	data.Sgmnt = make(map[int][]Segment)
	query := "SELECT * FROM objects;"

	err := DB.Select(&NewObject, query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	for _, val := range NewObject {
		data.Obj[val.SegmendKmId] = append(data.Obj[val.SegmendKmId], val)
	}
	query = "SELECT * FROM segments;"
	err = DB.Select(&NewSegment, query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	for _, val := range NewSegment {
		data.Sgmnt[val.KilometerId] = append(data.Sgmnt[val.KilometerId], val)
	}
	Data = data
	return nil

}

var FindObjects []ObjectInfo
var Date string
var Time string

func TakeCoordFromDB(latitude, longitude float64) error {

	objects := Data.Obj
	for _, val := range objects {
		for _, value := range val {
			value.ObjectLatitude = CoordConvert(value.ObjectLatitude, value.ObjectId)
			value.ObjectLongitude = CoordConvert(value.ObjectLongitude, value.ObjectId)
			maxLatitude, minLatitude := RadiusLatitude(value.ObjectLatitude, value.CoordRadius)
			maxLongitude, minLongitude := RadiusLongitude(value.ObjectLongitude, value.CoordRadius)
			//fmt.Println("Из бд после:", value.ObjectLatitude, value.ObjectLongitude)
			//fmt.Println("Из сообщения после после:", latitude, longitude)
			if latitude <= maxLatitude && latitude >= minLatitude && longitude <= maxLongitude && longitude >= minLongitude {
				//fmt.Printf("Найденный участок:\nID:%d\nКм реки: %d\nИмя объекта: %s\nШирота: %f\nДолгота:%f\nРадиус координат:%f\nОписание%s\n", value.ObjectId, value.SegmendKmId, value.ObjectName, value.ObjectLatitude, value.ObjectLongitude, value.CoordRadius, value.Description)
				value.OMMSI = ObjMMSI
				datetime, err := TimeToNormal(Date, Time)
				if err != nil {
					log.Println(err)
					continue
				}
				value.DateDB = datetime
				FindObjects = append(FindObjects, value)
				MmsiFound(value)
				return nil
			}
		}
	}
	//fmt.Println("Не найдено координат в бд")
	return fmt.Errorf("не найдено совпадений по коориднатам:%f %f", latitude, longitude)
}

func TimeToNormal(date string, timeh string) (time.Time, error) {
	datestr := date + " " + timeh
	t, err := time.Parse("02.01.2006 15:04:05", datestr)
	if err != nil {
		return t, err
	}
	return t, nil
}

// сделать форматирование
func CoordConvert(coords float64, id int) float64 {
	//fmt.Println(coords, id)
	degrees := float64(int(coords) / 100)
	min := (coords - (degrees * 100)) / 60
	result := degrees + min
	//fmt.Println("из бд до:")
	result = Round(result)
	//fmt.Println(result)
	return result
}
func Round(value float64) float64 {
	valInt := int64(value * 1000000)
	value = float64(valInt) / 1000000
	return value
}

func RadiusLatitude(latitude, radius float64) (float64, float64) {
	difference := radius / 111.133
	maxlatitude := latitude + difference
	minlatitude := latitude - difference
	return maxlatitude, minlatitude
}

func RadiusLongitude(longitude, radius float64) (float64, float64) {
	radians := longitude * math.Pi / 180
	degrees := 111.319 * math.Cos(radians)
	difference := radius / degrees
	maxlongitude := longitude + difference
	minlongitude := longitude - difference
	return maxlongitude, minlongitude
}
