package web

import (
	"net/http"

	"sumeru/core/server/api"
)

func writeWebJSONError(w http.ResponseWriter, status int, code, message string, details map[string]interface{}) {
	api.WriteResponse(w, status, api.Fail(code, message, details))
}
