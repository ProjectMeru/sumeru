package web

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/module"
	"sumeru/core/orm"
)

// ModuleActionHandler handles POST install, uninstall, activate, deactivate, and save on sys.module.
func ModuleActionHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLoginAndPOST(w, r) {
		return
	}
	if !requireModelAccess(w, r, appsModuleModel, "write") {
		return
	}

	form := parseModuleActionForm(r)
	if form.ModuleName == "" {
		redirectToAppsList(w, r, moduleMsgMissingModule, form.Browse)
		return
	}

	switch form.Action {
	case moduleActionSaveModule:
		handleModuleSaveAction(w, r, form)
	case moduleActionInstall, moduleActionUninstall, moduleActionDeactivate, moduleActionActivate:
		handleModuleLifecycleAction(w, r, form)
	default:
		redirectToAppsList(w, r, moduleMsgUnknownAction, form.Browse)
	}
}

type moduleActionForm struct {
	Action     string
	ModuleName string
	Browse     appsBrowseState
}

func parseModuleActionForm(r *http.Request) moduleActionForm {
	return moduleActionForm{
		Action:     strings.TrimSpace(r.FormValue(moduleActionField)),
		ModuleName: strings.TrimSpace(r.FormValue(moduleNameField)),
		Browse:     parseAppsBrowseStateFromForm(r),
	}
}

func redirectToAppsList(w http.ResponseWriter, r *http.Request, message string, browse appsBrowseState) {
	http.Redirect(w, r, appsRedirectURL(message, browse), http.StatusSeeOther)
}

func redirectToAppsDetail(w http.ResponseWriter, r *http.Request, message string, browse appsBrowseState) {
	http.Redirect(w, r, appsRedirectURL(message, browse), http.StatusSeeOther)
}

func handleModuleSaveAction(w http.ResponseWriter, r *http.Request, form moduleActionForm) {
	if err := saveModuleFromForm(r, form.ModuleName); err != nil {
		redirectToAppsList(w, r, formatAppsActionError(err.Error()), form.Browse)
		return
	}
	redirectToAppsDetail(w, r, moduleMsgSaved, withModuleName(form.Browse, form.ModuleName))
}

func handleModuleLifecycleAction(w http.ResponseWriter, r *http.Request, form moduleActionForm) {
	flashMessage, err := runModuleLifecycleAction(r.Context(), form.Action, form.ModuleName)
	if err != nil {
		redirectToAppsList(w, r, formatAppsActionError(err.Error()), form.Browse)
		return
	}
	WebLogNavigation(r.Context(), moduleActionRoute, "module_action", "Module action completed", map[string]interface{}{
		"do":     form.Action,
		"module": form.ModuleName,
	})
	browse := form.Browse
	if form.Action == moduleActionUninstall {
		redirectToAppsList(w, r, flashMessage, browse)
		return
	}
	redirectToAppsDetail(w, r, flashMessage, withModuleName(browse, form.ModuleName))
}

func runModuleLifecycleAction(ctx context.Context, action, moduleName string) (flashMessage string, err error) {
	var run func(context.Context, string) error
	var outcomeVerb string

	switch action {
	case moduleActionInstall:
		run = module.InstallModuleByName
		outcomeVerb = "installed"
	case moduleActionUninstall:
		run = module.UninstallModuleByName
		outcomeVerb = "uninstalled"
	case moduleActionDeactivate:
		run = func(ctx context.Context, name string) error {
			return module.SetModuleActive(ctx, name, false)
		}
		outcomeVerb = "deactivated"
	case moduleActionActivate:
		run = func(ctx context.Context, name string) error {
			return module.SetModuleActive(ctx, name, true)
		}
		outcomeVerb = "activated"
	default:
		return "", errors.New(moduleMsgUnknownAction)
	}

	if err = run(ctx, moduleName); err != nil {
		return "", err
	}
	return outcomeVerb + "_" + moduleName, nil
}

func saveModuleFromForm(r *http.Request, moduleName string) error {
	recordID, err := strconv.Atoi(strings.TrimSpace(r.FormValue(moduleRowIDField)))
	if err != nil || recordID <= 0 {
		return errors.New(moduleMsgInvalidModuleRow)
	}

	row, err := orm.SearchOne(r.Context(), appsModuleModel, map[string]interface{}{"id": recordID})
	if err != nil {
		return errors.New(moduleMsgModuleNotFound)
	}

	parsed, ok := parseModuleRow(row)
	if !ok || parsed.Name != moduleName {
		return errors.New(moduleMsgModuleMismatch)
	}

	return orm.UpdateRecordByID(r.Context(), appsModuleModel, recordID, map[string]interface{}{
		"display_name": strings.TrimSpace(r.FormValue(moduleDisplayNameField)),
		"author":       strings.TrimSpace(r.FormValue(moduleAuthorField)),
		"description":  strings.TrimSpace(r.FormValue(moduleDescriptionField)),
	})
}
