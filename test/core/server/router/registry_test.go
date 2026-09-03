package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"sumeru/core/server/router"
)

func TestNormalizeHTTPMethod(t *testing.T) {
	if got := router.NormalizeHTTPMethod(" post "); got != http.MethodPost {
		t.Fatalf("got %q want POST", got)
	}
	if got := router.NormalizeHTTPMethod(""); got != http.MethodGet {
		t.Fatalf("empty method should default to GET, got %q", got)
	}
}

func TestIsRegisterableRoute(t *testing.T) {
	if !router.IsRegisterableRoute("/ok", func(http.ResponseWriter, *http.Request) {}) {
		t.Fatal("valid route should be registerable")
	}
	if router.IsRegisterableRoute("  ", func(http.ResponseWriter, *http.Request) {}) {
		t.Fatal("blank path should be rejected")
	}
	if router.IsRegisterableRoute("/ok", nil) {
		t.Fatal("nil handler should be rejected")
	}
}

func TestGroupRoutesByPath_preservesFirstSeenOrder(t *testing.T) {
	routes := []router.Route{
		{Method: http.MethodGet, Path: "/b"},
		{Method: http.MethodPost, Path: "/a"},
		{Method: http.MethodPut, Path: "/b"},
	}
	pathsInOrder, routesByPath := router.GroupRoutesByPath(routes)

	wantPaths := []string{"/b", "/a"}
	if len(pathsInOrder) != len(wantPaths) {
		t.Fatalf("paths = %v, want %v", pathsInOrder, wantPaths)
	}
	for i, path := range wantPaths {
		if pathsInOrder[i] != path {
			t.Fatalf("paths[%d] = %q, want %q", i, pathsInOrder[i], path)
		}
	}
	if len(routesByPath["/b"]) != 2 || len(routesByPath["/a"]) != 1 {
		t.Fatalf("unexpected grouped routes: %#v", routesByPath)
	}
}

func TestRequestMatchesRoute(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/t", nil)
	route := router.Route{Method: http.MethodPost, Path: "/t"}
	if !router.RequestMatchesRoute(request, route) {
		t.Fatal("matching method should match route")
	}
	route.Method = http.MethodGet
	if router.RequestMatchesRoute(request, route) {
		t.Fatal("different method should not match")
	}
	route.Method = ""
	if !router.RequestMatchesRoute(request, route) {
		t.Fatal("empty route method should match any request method")
	}
}

func TestUpsertRoute_replacesSamePathAndMethod(t *testing.T) {
	router.Clear()
	t.Cleanup(router.Clear)

	first := func(http.ResponseWriter, *http.Request) {}
	second := func(http.ResponseWriter, *http.Request) {}

	router.UpsertRouteForTest(router.Route{Method: http.MethodGet, Path: "/t", Handler: first})
	router.UpsertRouteForTest(router.Route{Method: http.MethodGet, Path: "/t", Handler: second})

	if router.RegisteredRouteCountForTest() != 1 {
		t.Fatalf("expected one route, got %d", router.RegisteredRouteCountForTest())
	}
}
