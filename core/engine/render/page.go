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
	if len(data.SWCBootstrapJSON) == 0 {
		data.SWCBootstrapJSON = BuildSWCBootstrapJSON(ctx, data, bootstrapWorkspace)
	}
	tmpl, err := template.ParseFiles(
		filepath.Join(templatesDir, "base.html"),
		filepath.Join(templatesDir, "shell_partials.html"),
	)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	// Must execute the layout by name: the partials file only defines {{template "sumMenuIcon"}}.
	if err := tmpl.ExecuteTemplate(&buf, "base.html", data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
