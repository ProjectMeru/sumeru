module sumeru

go 1.26.2

require (
	github.com/gorilla/websocket v1.5.3
	github.com/gpdf-dev/gpdf v1.0.11
	github.com/lib/pq v1.12.3
	golang.org/x/crypto v0.55.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require github.com/DATA-DOG/go-sqlmock v1.5.2 // indirect

replace github.com/gpdf-dev/gpdf => github.com/ProjectMeru/gpdf v1.0.11

replace github.com/gorilla/websocket => github.com/ProjectMeru/websocket v1.5.3
