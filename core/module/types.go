package module

import (
	"sumeru/core/engine/parser"
)

const KernelModule = "base"

type Manifest struct {
	Name         string              `json:"name"`
	DisplayName  string              `json:"display_name"` // optional; Apps / shell label
	Version      string              `json:"version"`
	Depends      []string            `json:"depends"`
	Author       string              `json:"author"`
	Description  string              `json:"description"`
	Icon         string              `json:"icon"`        // optional relative path under addon, e.g. static/icon.png
	Data         []string            `json:"data"`        // XML files to load
	Application  *bool               `json:"application"` // nil = true (show in Apps)
	AutoImport   *bool               `json:"auto_import"` // nil = true; false = omit from generated zimports blank imports
	SwcEntry     string              `json:"swc_entry"`   // optional SWC addon entry module URL
	Assets       []string            `json:"assets"`      // optional CSS/JS paths relative to addon root
	AssetBundles map[string][]string `json:"asset_bundles"` // named lazy bundles, e.g. swc.backend
}

func (manifest *Manifest) IsAutoImport() bool {
	if manifest.AutoImport == nil {
		return true
	}
	return *manifest.AutoImport
}

func (manifest *Manifest) IsApplication() bool {
	if manifest.Application == nil {
		return true
	}
	return *manifest.Application
}

type Addon struct {
	Manifest Manifest
	Path     string
	Menus    []parser.MenuItem
}
