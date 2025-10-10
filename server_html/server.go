package server_html

import (
	"log"
	"net/http"
)

func ServerRequest() {
	//подгрузка статики
	fs := http.FileServer(http.Dir("./server_html/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", HomePage)
	http.HandleFunc("/foldercheck", FolderCheck)
	http.HandleFunc("/scan", FileScan)
	http.HandleFunc("/reload", Reload)
	http.HandleFunc("/segments", DbSegments)
	http.HandleFunc("/objects", DbObjects)
	http.HandleFunc("/api/messages", GetMessages)
	http.HandleFunc("/api/objects", GetObjects)
	http.HandleFunc("/api/errors", GetErrors)
	http.HandleFunc("/api/count", GetFileCount)
	http.HandleFunc("/mmsi", GetMmsi)
	http.HandleFunc("/mmsi/", MmsiCheck)
	http.HandleFunc("/objects/new", NewObject)
	http.HandleFunc("/objects/delete", DeleteObject)
	http.HandleFunc("/objects/edit", EditObject)
	http.HandleFunc("/newfolder", NewFolder)
	http.HandleFunc("/reloadauto", AutoScan)
	http.HandleFunc("/segments/new", NewSegment)
	http.HandleFunc("/segments/edit", EditSegment)
	http.HandleFunc("/segments/delete", SegmentDelete)
	err := http.ListenAndServe(":8080", nil)
	log.Fatal(err)
}
