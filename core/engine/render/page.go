package render

import (
	"bytes"
	"context"
	"html/template"
	"path/filepath"
)

// RenderPage executes base.html with the given shell + content data.
func RenderPage(ctx context.Context, templatesDir string, data PageData, bootstrapWorkspace *SWCBootstrapWorkspace) (string, error) {
	EnrichShellPageData(ctx, &data)
	if 0 == len(data.SWCBootstrapJSON) {
		data.SWCBootstrapJSON = BuildSWCBootstrapJSON(ctx, data, bootstrapWorkspace)
	}
	paths := ShellLayoutTemplateFiles(templatesDir)
	tmpl, err := template.ParseFiles(paths...)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	// Must execute the layout by name: the partials file only defines {{template "sumMenuIcon"}}.
	if err := tmpl.ExecuteTemplate(&buf, filepath.Base(TemplateShellBase), data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
