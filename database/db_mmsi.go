package database

import (
	"log"
)

func MmsiFound(val ObjectInfo) {
	query := `INSERT INTO ships(
		object_id, segment_km_id, object_name, object_latitude, object_longitude,
		coord_radius, descriptions, mmsi, date_db
	)
	SELECT @p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9
	WHERE NOT EXISTS (
		SELECT 1 FROM ships WHERE mmsi=@p8 AND date_db=@p9
	);`

	res, err := DB.Exec(query,
		val.ObjectId,
		val.SegmendKmId,
		val.ObjectName,
		val.ObjectLatitude,
		val.ObjectLongitude,
		val.CoordRadius,
		val.Description,
		val.OMMSI,
		val.DateDB,
	)

	if err != nil {
		log.Println("не удалось вставить запись совпадений", err)
		AllErrors = append(AllErrors, MsgErrors{
			Err:    err.Error(),
			ErrMsg: "",
		})
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected > 0 {
		log.Println("Запись успешно вставлена:", val)
	}
}
