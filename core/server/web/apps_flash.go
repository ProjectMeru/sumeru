package web

import (
	"fmt"
	"strings"

	"sumeru/core/engine/render"
	"sumeru/core/module"
)

func buildModuleDisplayNameMap(modules []appsModule) map[string]string {
	names := make(map[string]string, len(modules))
	for _, mod := range modules {
		names[mod.Name] = mod.DisplayName
	}
	return names
}

func moduleDisplayNameFromMap(name string, displayNames map[string]string) string {
	if dn, ok := displayNames[name]; ok && strings.TrimSpace(dn) != "" {
		return dn
	}
	if addon, ok := module.DiscoveredAddons[name]; ok {
		return moduleDisplayName(name, addon.Manifest.DisplayName)
	}
	return name
}

// appsFlashFromMessage converts Apps page ?msg= values into structured flash messages.
func appsFlashFromMessage(msg string, displayNames map[string]string) (render.FlashMessage, bool) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return render.FlashMessage{}, false
	}

	lifecycle := []struct {
		prefix string
		title  string
		body   string
		toast  bool
	}{
		{"installed_", "App installed", "%s is ready to use.", true},
		{"uninstalled_", "App removed", "%s was uninstalled.", true},
		{"activated_", "App activated", "%s menus are visible again.", true},
		{"deactivated_", "App deactivated", "%s menus are hidden.", true},
		{"upgraded_", "App updated", "%s was updated.", true},
	}
	for _, entry := range lifecycle {
		if !strings.HasPrefix(msg, entry.prefix) {
			continue
		}
		modName := strings.TrimPrefix(msg, entry.prefix)
		display := moduleDisplayNameFromMap(modName, displayNames)
		return render.FlashMessage{
			Kind:      "success",
			Title:     entry.title,
			Body:      fmt.Sprintf(entry.body, display),
			ToastOnly: entry.toast,
		}, true
	}

	switch msg {
	case moduleMsgSaved:
		return render.FlashMessage{
			Kind: "success", Title: "Saved", Body: "Module details were saved.", ToastOnly: true,
		}, true
	case moduleMsgMissingModule:
		return render.FlashMessage{Kind: "error", Title: "Error", Body: "No module was specified."}, true
	case moduleMsgUnknownAction:
		return render.FlashMessage{Kind: "error", Title: "Error", Body: "Unknown action."}, true
	case moduleMsgInvalidModuleRow, moduleMsgModuleNotFound, moduleMsgModuleMismatch:
		return render.FlashMessage{Kind: "error", Title: "Error", Body: strings.ReplaceAll(msg, "_", " ")}, true
	}

	if flash, ok := flashFromQueryMessage(msg); ok {
		return flash, true
	}
	return render.FlashMessage{Kind: "error", Title: "Action failed", Body: msg}, msg != ""
}

func splitAppsPageFlashes(msg string, displayNames map[string]string) (inline, toast []render.FlashMessage) {
	flash, ok := appsFlashFromMessage(msg, displayNames)
	if !ok {
		return nil, nil
	}
	if flash.ToastOnly {
		return nil, []render.FlashMessage{flash}
	}
	return []render.FlashMessage{flash}, nil
}

func formatAppsActionError(message string) string {
	message = strings.TrimSpace(message)
	if message == "" {
		return ""
	}
	if strings.HasPrefix(message, "error:") || strings.HasPrefix(message, "save_error:") {
		return message
	}
	return "error:" + message
}
