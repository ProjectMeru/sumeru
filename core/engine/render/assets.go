package render

var ExtraStylesheetURLs []string

var ExtraScriptURLs []string

func SetExtraStylesheetURLs(urls []string) {
	ExtraStylesheetURLs = append([]string(nil), urls...)
}

func SetExtraScriptURLs(urls []string) {
	ExtraScriptURLs = append([]string(nil), urls...)
}
