package web_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sumeru/core/server/web"
	"testing"
)

func TestHttpStatusFromWorkspaceError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{name: "invalid id", err: errors.New("invalid id"), wantStatus: http.StatusBadRequest},
		{name: "record not found", err: errors.New("record 5 not found"), wantStatus: http.StatusNotFound},
		{name: "no view", err: errors.New("No view for model core.user"), wantStatus: http.StatusNotFound},
		{name: "internal", err: errors.New("list load: timeout"), wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := web.HTTPStatusFromWorkspaceError(tt.err); got != tt.wantStatus {
				t.Fatalf("got status %d, want %d", got, tt.wantStatus)
			}
		})
	}
}

func TestWorkspaceQueryParams(t *testing.T) {
	req := httptest.NewRequest("GET", web.TestWorkspaceRoute+"?"+web.TestWorkspaceActionParam+"=3&"+web.TestWorkspaceMenuIDParam+"=9", nil)
	actionQuery, menuQuery := web.WorkspaceQueryParams(req)
	if actionQuery != "3" || menuQuery != "9" {
		t.Fatalf("got action=%q menu=%q", actionQuery, menuQuery)
	}
}

func TestRedirectIfMenuAccessDenied_emptyMenu(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", web.TestWorkspaceRoute, nil)
	if web.RedirectIfMenuAccessDenied(recorder, request, "") {
		t.Fatal("empty menu should not redirect")
	}
}
