module sumeru

go 1.26.9

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/gorilla/websocket v1.5.3
	github.com/gpdf-dev/gpdf v1.0.13
	github.com/lib/pq v1.12.3
	github.com/piglig/go-qr v1.1.0
	golang.org/x/crypto v0.57.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

replace github.com/gpdf-dev/gpdf => github.com/ProjectMeru/gpdf v1.0.13

replace github.com/gorilla/websocket => github.com/ProjectMeru/websocket v1.5.3

replace github.com/piglig/go-qr => github.com/ProjectMeru/go-qr v1.1.0
