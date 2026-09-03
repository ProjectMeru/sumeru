package router

import "net/http"

func NormalizeHTTPMethod(method string) string { return normalizeHTTPMethod(method) }

func IsRegisterableRoute(path string, handler http.HandlerFunc) bool {
	return isRegisterableRoute(path, handler)
}

func GroupRoutesByPath(routes []Route) (pathsInOrder []string, routesByPath map[string][]Route) {
	return groupRoutesByPath(routes)
}

func RequestMatchesRoute(r *http.Request, route Route) bool { return requestMatchesRoute(r, route) }

func UpsertRouteForTest(route Route) { upsertRoute(route) }

func RegisteredRouteCountForTest() int { return len(registeredRoutes) }
