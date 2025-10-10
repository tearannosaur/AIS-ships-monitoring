package server_html

import (
	"ais/database"
	"ais/handler"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type PageData struct {
	Success string
	Error   string
}

func HtmlErr(err error, w http.ResponseWriter) bool {
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return true
	}
	return false
}

func Reload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Не Post Запрос", http.StatusMethodNotAllowed)
		return
	}
	handler.IsScannig = true
	fmt.Println(handler.ResumeScanPause)
	if !handler.ResumeScanPause {
		handler.ResumeScan <- struct{}{}
	}
	w.WriteHeader(http.StatusOK)
}

func NewFolder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Не Post Запрос", http.StatusMethodNotAllowed)
		return
	}
	log.Println("Новая папка")
	if !handler.NewFolderPause {
		handler.StopScan <- struct{}{}
	}
	w.WriteHeader(http.StatusOK)
}

func AutoScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Не Post Запрос", http.StatusMethodNotAllowed)
		return
	}
	if !handler.AutoScanPause {
		handler.AutoScan <- struct{}{}
	}
	w.WriteHeader(http.StatusOK)
}

func FileScan(w http.ResponseWriter, r *http.Request) {
	html, err := template.ParseFiles("server_html/filescan.html")
	if HtmlErr(err, w) {
		return
	}
	html.Execute(w, nil)
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	html, err := template.ParseFiles("server_html/mainpage.html")
	if HtmlErr(err, w) {
		return
	}
	err = html.Execute(w, nil)
	if HtmlErr(err, w) {
		return
	}
}

var PathChan = make(chan string)

func FolderCheck(w http.ResponseWriter, r *http.Request) {
	path := r.FormValue("path")
	if path != "" {
		PathChan <- path
	}
	parrentdir := r.FormValue("folder")
	if parrentdir != "" {
		handler.Folder = parrentdir
	}
	handler.NewFolderPause = false
	http.Redirect(w, r, "/scan", http.StatusFound)
}

func DbSegments(w http.ResponseWriter, r *http.Request) {
	html, err := template.ParseFiles("server_html/segments.html")
	HtmlErr(err, w)
	err = html.Execute(w, database.NewSegment)
	if HtmlErr(err, w) {
		return
	}
}

func DbObjects(w http.ResponseWriter, r *http.Request) {
	html, err := template.ParseFiles("server_html/objects.html")
	if HtmlErr(err, w) {
		return
	}
	err = html.Execute(w, database.NewObject)
	if HtmlErr(err, w) {
		return
	}
}

func GetMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if handler.AllMessageinFile == nil {
		json.NewEncoder(w).Encode([]interface{}{}) // пустой массив вместо null
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 50
	}

	if offset < 0 {
		offset = 0
	}

	if offset >= len(handler.AllMessageinFile) {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	end := offset + limit
	if end > len(handler.AllMessageinFile) {
		end = len(handler.AllMessageinFile)
	}

	json.NewEncoder(w).Encode(handler.AllMessageinFile[offset:end])
}

func GetObjects(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if database.FindObjects == nil {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	if offset >= len(database.FindObjects) {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	end := offset + limit
	if end > len(database.FindObjects) {
		end = len(database.FindObjects)
	}

	json.NewEncoder(w).Encode(database.FindObjects[offset:end])
}

func GetFileCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(handler.CountFile)
}

func GetErrors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if database.AllErrors == nil {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	if offset >= len(database.AllErrors) {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	end := offset + limit
	if end > len(database.AllErrors) {
		end = len(database.AllErrors)
	}

	json.NewEncoder(w).Encode(database.AllErrors[offset:end])
}

func GetMmsi(w http.ResponseWriter, r *http.Request) {
	html, err := template.ParseFiles("server_html/mmsi.html")
	if HtmlErr(err, w) {
		return
	}
	query := `SELECT DISTINCT mmsi FROM ships;`
	mmsi := make([]int, 0, 60)
	err = database.DB.Select(&mmsi, query)
	if err != nil {
		log.Println("Не удалось достать mmsi из бд", err)
	}
	err = html.Execute(w, mmsi)
	if HtmlErr(err, w) {
		return
	}
}

type MmsiData struct {
	MmsiPart []database.ObjectInfo
	MmsiAll  []database.ObjectInfo
	ShipData
}
type ShipData struct {
	Pages []int
	Mmsi  int64
}

func MmsiCheck(w http.ResponseWriter, r *http.Request) {
	html, err := template.New("mmsicheck.html").Funcs(template.FuncMap{
		"toJson": toJson,
	}).ParseFiles("server_html/mmsicheck.html")
	if HtmlErr(err, w) {
		return
	}
	var mmsidata MmsiData
	path := strings.TrimPrefix(r.URL.Path, "/mmsi/")
	path = strings.TrimSuffix(path, "/")
	//fmt.Println(path)
	pathmmsi, err := strconv.Atoi(path)
	if err != nil {
		log.Println(err)
	}

	query := `SELECT id,mmsi,object_id,segment_km_id,object_name,object_latitude,
		object_longitude,coord_radius,descriptions,date_db FROM ships
		WHERE mmsi=@p1`

	mmsidata.MmsiAll = make([]database.ObjectInfo, 0, 300)
	err = database.DB.Select(&mmsidata.MmsiAll, query, pathmmsi)
	if err != nil {
		log.Println(err)
	}
	allpages := int(math.Ceil(float64(len(mmsidata.MmsiAll)) / 30.0))
	pages := make([]int, 0, allpages)
	for i := 0; i < allpages; i++ {
		pages = append(pages, i+1)
	}
	mmsidata.Pages = pages
	p := r.URL.Query().Get("page")
	//fmt.Println(p)

	mmsidata.Mmsi = int64(pathmmsi)
	if p == "" {
		p = "1"
	}

	page, err := strconv.Atoi(p)
	if err != nil {
		page = 1
	}

	if page < 1 {
		page = 1
	}
	if page > allpages {
		page = allpages
	}

	start := (page - 1) * 30
	end := start + 30
	if end > len(mmsidata.MmsiAll) {
		end = len(mmsidata.MmsiAll)
	}
	mmsidata.MmsiPart = mmsidata.MmsiAll[start:end]
	err = html.Execute(w, mmsidata)
	if HtmlErr(err, w) {
		return
	}
}

func toJson(v interface{}) template.JS {
	b, _ := json.Marshal(v)
	return template.JS(string(b))
}

func NewObject(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		html, err := template.ParseFiles("server_html/newobject.html")
		if HtmlErr(err, w) {
			return
		}
		err = html.Execute(w, PageData{
			Success: r.URL.Query().Get("success"),
			Error:   r.URL.Query().Get("error"),
		})
		if HtmlErr(err, w) {
			return
		}
		return
	}
	var count int
	objkm := r.FormValue("objkm")
	objname := r.FormValue("objname")
	objlatitude := r.FormValue("objlatitude")
	objlongitude := r.FormValue("objlongitude")
	objradius := r.FormValue("objradius")
	objdescription := r.FormValue("objdescription")
	if objdescription == "" {
		objdescription = " "
	}
	query := `SELECT COUNT(*) FROM segments WHERE kilometer_id=@p1`
	err := database.DB.Get(&count, query, objkm)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/objects/new?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	if count == 0 {
		err = errors.New("км реки не существует в базе данных")
		log.Println(err)
		http.Redirect(w, r, "/objects/new?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	query = `INSERT INTO objects (segment_km_id,object_name,object_latitude,object_longitude,coord_radius,descriptions) 
	VALUES(@p1,@p2,@p3,@p4,@p5,@p6)`
	_, err = database.DB.Exec(query, objkm, objname, objlatitude, objlongitude, objradius, objdescription)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/objects/new?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	log.Println("Запись успешно добавлена", objkm, objname, objlatitude, objlongitude, objradius, objdescription)
	database.DbDownload()
	http.Redirect(w, r, "/objects/new?success=1", http.StatusSeeOther)

}

func DeleteObject(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		html, err := template.ParseFiles("server_html/deleteobject.html")
		if HtmlErr(err, w) {
			return
		}
		err = html.Execute(w, PageData{
			Success: r.URL.Query().Get("success"),
			Error:   r.URL.Query().Get("error"),
		})
		if HtmlErr(err, w) {
			return
		}
		return
	}
	var count int
	objid := r.FormValue("objid")
	query := `SELECT COUNT(*) FROM objects WHERE object_id=@p1`
	err := database.DB.Get(&count, query, objid)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/objects/delete?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	if count == 0 {
		err = errors.New("вы ввели несущетсвующий id")
		log.Println(err)
		http.Redirect(w, r, "/objects/delete?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	query = `DELETE FROM objects WHERE object_id=@p1`
	_, err = database.DB.Exec(query, objid)
	if err != nil {
		log.Println("Не удалось удалить объект", err)
		http.Redirect(w, r, "/objects/delete?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	log.Println("Запись удалена по id объекта:", objid)
	database.DbDownload()
	http.Redirect(w, r, "/objects/delete?success=1", http.StatusSeeOther)

}

func EditObject(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		html, err := template.ParseFiles("server_html/editobject.html")
		HtmlErr(err, w)
		err = html.Execute(w, PageData{
			Success: r.URL.Query().Get("success"),
			Error:   r.URL.Query().Get("error"),
		})
		HtmlErr(err, w)
		return
	}
	objid := r.FormValue("objid")
	objkm := r.FormValue("objkm")
	objname := r.FormValue("objname")
	objlatitude := r.FormValue("objlatitude")
	objlongitude := r.FormValue("objlongitude")
	objradius := r.FormValue("objradius")
	objdescription := r.FormValue("objdescription")
	var count int
	object := database.ObjectInfo{}
	query := `SELECT COUNT(*) FROM objects WHERE object_id=@p1`
	err := database.DB.Get(&count, query, objid)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/objects/edit?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	if count == 0 {
		err = errors.New("id объекта не найден в базе данных")
		log.Println(err)
		http.Redirect(w, r, "/objects/edit?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	query = `SELECT * FROM objects WHERE object_id=@p1`
	err = database.DB.Get(&object, query, objid)
	if err != nil {
		log.Println("Ошибка проверки:", err)
		http.Redirect(w, r, "/objects/edit?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	if objdescription == "" {
		objdescription = object.Description
	}

	if objkm == "" {
		objkm = strconv.Itoa(object.SegmendKmId)
	}
	if objname == "" {
		objname = object.ObjectName
	}
	if objlatitude == "" {
		objlatitude = strconv.FormatFloat(object.ObjectLatitude, 'f', -1, 64)
	}
	if objlongitude == "" {
		objlongitude = strconv.FormatFloat(object.ObjectLongitude, 'f', -1, 64)
	}
	if objradius == "" {
		objradius = strconv.FormatFloat(object.CoordRadius, 'f', -1, 64)
	}
	km, err := strconv.ParseInt(objkm, 10, 64)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/objects/edit?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	latitude, err := strconv.ParseFloat(objlatitude, 64)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/objects/edit?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	longitude, err := strconv.ParseFloat(objlongitude, 64)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/objects/edit?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	radius, err := strconv.ParseFloat(objradius, 64)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/objects/edit?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	query = `UPDATE objects
	SET segment_km_id=@p1,object_name=@p2,object_latitude=@p3,
	object_longitude=@p4,coord_radius=@p5,descriptions=@p6 WHERE object_id=@p7`
	_, err = database.DB.Exec(query, km, objname, latitude, longitude, radius, objdescription, objid)
	if err != nil {
		log.Println("Ошибка обновления:", err)
		http.Redirect(w, r, "/objects/edit?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	log.Println("Запись успешного обновлена по id:", objid)
	database.DbDownload()
	http.Redirect(w, r, "/objects/edit?success=1", http.StatusSeeOther)
}

func NewSegment(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		html, err := template.ParseFiles("server_html/newsegment.html")
		if HtmlErr(err, w) {
			return
		}
		err = html.Execute(w, PageData{
			Success: r.URL.Query().Get("success"),
			Error:   r.URL.Query().Get("error"),
		})
		if HtmlErr(err, w) {
			return
		}
		return
	}
	km := r.FormValue("km")
	city := r.FormValue("city")
	depth := r.FormValue("depth")
	latitude := r.FormValue("latitude")
	longitude := r.FormValue("longitude")

	query := `INSERT INTO segments (kilometer_id,city,depth,segment_latitude,segment_longitude) 
	VALUES(@p1,@p2,@p3,@p4,@p5)`
	_, err := database.DB.Exec(query, km, city, depth, latitude, longitude)
	if err != nil {
		log.Println("Ошибка добавления сегмента", err)
		http.Redirect(w, r, "/segments/new?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	database.DbDownload()
	log.Println("запись добавлена по сегменту", km)
	http.Redirect(w, r, "/segments/new?success=1", http.StatusSeeOther)
}

func EditSegment(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		html, err := template.ParseFiles("server_html/editsegment.html")
		if HtmlErr(err, w) {
			return
		}
		err = html.Execute(w, PageData{
			Success: r.URL.Query().Get("success"),
			Error:   r.URL.Query().Get("error"),
		})
		if HtmlErr(err, w) {
			return
		}
		return
	}
	id := r.FormValue("id")
	city := r.FormValue("city")
	depth := r.FormValue("depth")
	latitude := r.FormValue("latitude")
	longitude := r.FormValue("longitude")
	segment := database.Segment{}
	var count int
	query := `SELECT COUNT(*) FROM segments WHERE kilometer_id=@p1`
	err := database.DB.Get(&count, query, id)
	if err != nil {
		log.Println("Ошибка проверки сегмента", err)
		http.Redirect(w, r, "/segments/edit?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	if count == 0 {
		err = errors.New("км сегмента не найден в базе данных")
		http.Redirect(w, r, "/segments/edit?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	query = `SELECT city,depth,segment_latitude,segment_longitude FROM segments
	WHERE kilometer_id=@p1`
	err = database.DB.Get(&segment, query, id)
	if err != nil {
		log.Println("Ошибка получения сегмента", err)
		http.Redirect(w, r, "/segments/edit?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	if city == "" {
		city = segment.City
	}
	if depth == "" {
		depth = strconv.FormatFloat(segment.Depth, 'f', -1, 64)
	}
	if latitude == "" {
		latitude = strconv.FormatFloat(segment.SegmentLatitude, 'f', -1, 64)
	}
	if longitude == "" {
		longitude = strconv.FormatFloat(segment.SegmentLongitude, 'f', -1, 64)
	}

	sdepth, err := strconv.ParseFloat(depth, 64)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/segments/edit?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	slatitude, err := strconv.ParseFloat(latitude, 64)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/segments/edit?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	slongitude, err := strconv.ParseFloat(longitude, 64)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/segments/edit?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	query = `UPDATE segments
	SET city=@p1,depth=@p2,segment_latitude=@p3,segment_longitude=@p4
	WHERE kilometer_id=@p5`
	_, err = database.DB.Exec(query, city, sdepth, slatitude, slongitude, id)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/segments/edit?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	log.Println("запись отредактирована по сегменту", id)
	database.DbDownload()
	http.Redirect(w, r, "/segments/edit?success=1", http.StatusSeeOther)
}

func SegmentDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		html, err := template.ParseFiles("server_html/segmentdelete.html")
		if HtmlErr(err, w) {
			return
		}
		err = html.Execute(w, PageData{
			Success: r.URL.Query().Get("success"),
			Error:   r.URL.Query().Get("error"),
		})
		if HtmlErr(err, w) {
			return
		}
		return
	}
	km := r.FormValue("km")
	var count int
	query := `SELECT COUNT(*) FROM segments
	WHERE kilometer_id=@p1`
	err := database.DB.Get(&count, query, km)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/segments/delete?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	if count == 0 {
		err = errors.New("км сегмента не найден в базе данных")
		http.Redirect(w, r, "/segments/delete?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	query = `DELETE FROM objects 
	WHERE segment_km_id=@p1`
	_, err = database.DB.Exec(query, km)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/segments/delete?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	query = `DELETE FROM segments
	WHERE kilometer_id=@p1`
	_, err = database.DB.Exec(query, km)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/segments/delete?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	database.DbDownload()
	log.Println("запись удалена по сегменту", km)
	http.Redirect(w, r, "/segments/delete?success=1", http.StatusSeeOther)

}
