
let msgPage = 0, msgPageSize = 50, msgLoading = false;
let objPage = 0, objPageSize = 50, objLoading = false;
let errPage = 0, errPageSize = 50, errLoading = false;


function loadMessages() {
    if (msgLoading) return;
    msgLoading = true;

    $.ajax({
        url: `/api/messages?limit=${msgPageSize}&offset=${msgPage * msgPageSize}`,
        success: function(data) {
            let html = '';
            data.forEach(function(msg) {
                html += '<p>Дата: ' + msg.Date + ' Время: ' + msg.Time +
                        ' Сообщение: ' + msg.Message + '</p>' +
                        '<p>Тип сообщения: ' + msg.MessageType +
                        ' MMSI: ' + msg.Mmsi +
                        ' Долгота: ' + msg.Longitude +
                        ' Широта: ' + msg.Latitude + '</p>';
            });

            if (msgPage === 0) {
                $('.block').html(html);
            } else {
                $('.block').append(html);
            }

            if (data.length > 0) msgPage++;
            msgLoading = false;
        },
        error: function() {
            msgLoading = false;
        }
    });
}


function loadObjects() {
    if (objLoading) return;
    objLoading = true;

    $.ajax({
        url: `/api/objects?limit=${objPageSize}&offset=${objPage * objPageSize}`,
        success: function(data) {
            let html = '';
            data.forEach(function(obj) {
                html += '<p>MMSI судна: ' + obj.ObjectMMSI + '</p>' +
                        '<p>ID Объекта: ' + obj.ObjectId + '</p>' +
                        'Км реки: ' + obj.SegmentId + '</p>' +
                        'Имя объекта: ' + obj.ObjectName + '</p>' +
                        '<p>Широта: ' + obj.ObjectLatitude + '</p>' +
                        '<p>Долгота: ' + obj.ObjectLongitude + '</p>' +
                        'Радиус координат: ' + obj.ObjectRadius + '</p>' +
                        'Описание: ' + obj.ObjectDescription + '<p>______________________________________</p>';
            });

            if (objPage === 0) {
                $('.blockfind').html(html);
            } else {
                $('.blockfind').append(html);
            }

            if (data.length > 0) objPage++;
            objLoading = false;
        },
        error: function() {
            objLoading = false;
        }
    });
}


function loadErrors() {
    if (errLoading) return;
    errLoading = true;

    $.ajax({
        url: `/api/errors?limit=${errPageSize}&offset=${errPage * errPageSize}`,
        success: function(data) {
            let html = '';
            data.forEach(function(er) {
                html += '<p>error: ' + er.Err + ' ' + er.ErrMsg + '</p>';
            });

            if (errPage === 0) {
                $('.blockerror').html(html);
            } else {
                $('.blockerror').append(html);
            }

            if (data.length > 0) errPage++;
            errLoading = false;
        },
        error: function() {
            errLoading = false;
        }
    });
}

function triggerScan() {
    $.ajax({
        url: '/reload',
        method: 'POST',
        success: function() {
            console.log('Сканирование запущено');
            msgPage = 0;
            objPage = 0;
            errPage = 0;
            $('.block').html('');
            $('.blockfind').html('');
            $('.blockerror').html('');
            loadMessages();
            loadObjects();
            loadErrors();
            loadCount();
        }
    });
}


function autoScan() {
    $.ajax({
        url: '/reloadauto',
        method: 'POST',
        success: function() {
            console.log('Автосканирование запущено');
        }
    });
}


$('.block').on('scroll', function() {
    if (this.scrollTop + this.clientHeight >= this.scrollHeight - 50) loadMessages();
});
$('.blockfind').on('scroll', function() {
    if (this.scrollTop + this.clientHeight >= this.scrollHeight - 50) loadObjects();
});
$('.blockerror').on('scroll', function() {
    if (this.scrollTop + this.clientHeight >= this.scrollHeight - 50) loadErrors();
});


$(document).ready(function() {
    $('#scanButton').on('click', function(e) {
        e.preventDefault();
        triggerScan();
    });

    $('#autoscanButton').on('click', function(e) {
        e.preventDefault();
        autoScan();
    });

    setInterval(loadMessages, 2000);
    setInterval(loadObjects, 2000);
    setInterval(loadErrors, 2000);
    setInterval(loadCount, 2000);
});

function loadCount() {
     $.ajax({
    url: '/api/count',
    success: function(data) {
        let html = '';
        console.log("Ответ сервера:", data);
        $('.filecount').html('Отсканировано файлов: '+data.FileNow +'/'+data.AllFiles);
        $('.objectcount').html('Найдено совпадений: '+data.FindCount);
        }
    });
}


function newFolder() {
    $.ajax({
        url: '/newfolder',
        method: 'POST',
        success: function() {
            console.log("Выбор папки заново, сканирование прервано");
            window.location.href = "/"; 
        }
    });
}

$(document).ready(function() {
    $('#newFolder').on('click', function(e) {
        e.preventDefault();
        newFolder();
    });
});


document.addEventListener("DOMContentLoaded", () => {
    if (!window.coords || window.coords.length === 0) {
        console.error("Нет координат для отображения");
        return;
    }

    const map = L.map('map').setView(
        [window.coords[0].ObjectLatitude, window.coords[0].ObjectLongitude],
        10
    );

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '&copy; OpenStreetMap contributors'
    }).addTo(map);

    const MAX_GAP_MINUTES = 5; 

  
    const sortedPoints = window.coords.sort((a, b) => new Date(a.DateDB) - new Date(b.DateDB));

    let routes = [];       
    let currentRoute = []; 
    let prevTime = null;

    sortedPoints.forEach((p) => {

        const circle = L.circleMarker([p.ObjectLatitude, p.ObjectLongitude], {
            radius: 5,
            color: 'blue',
            fillColor: 'blue',
            fillOpacity: 0.3
        }).bindPopup(`Id Точки: ${p.Id}<br>Имя объекта: ${p.ObjectName}<br>Описание: ${p.ObjectDescription}<br>Дата/Время: ${p.DateDB}`);
        circle.addTo(map);

        const pointTime = new Date(p.DateDB);


        if (prevTime) {
            const diffMinutes = (pointTime.getTime() - prevTime.getTime()) / (1000 * 60);
            if (diffMinutes > MAX_GAP_MINUTES) {
   
                if (currentRoute.length > 0) {
                    routes.push(currentRoute);
                }
                currentRoute = [];
            }
        }


        const lastCoord = currentRoute[currentRoute.length - 1];
        if (!lastCoord || lastCoord[0] !== p.ObjectLatitude || lastCoord[1] !== p.ObjectLongitude) {
            currentRoute.push([p.ObjectLatitude, p.ObjectLongitude]);
        }

        prevTime = pointTime;
    });

 
    if (currentRoute.length > 0) {
        routes.push(currentRoute);
    }


    routes.forEach((route) => {
        L.polyline(route, { color: 'red' }).addTo(map);
    });

    const allLatLngs = routes.flat();
    map.fitBounds(allLatLngs);
});