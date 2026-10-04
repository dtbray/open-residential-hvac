//go:build desktop

// SPDX-License-Identifier: AGPL-3.0-only
package main

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"git.thomas-bray.com/thomas/open-residential-hvac/app"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/loads"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/project"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

type Desktop struct {
	ctx     context.Context
	service app.Service
}

func (d *Desktop) Calculate(data string) (*loads.Result, error) { return d.service.Calculate(data) }
func (d *Desktop) OpenProject() (string, error) {
	path, err := runtime.OpenFileDialog(d.ctx, runtime.OpenDialogOptions{Title: "Open HVAC project", Filters: []runtime.FileFilter{{DisplayName: "HVAC project", Pattern: "*.json;*.yaml;*.yml"}}})
	if err != nil || path == "" {
		return "", err
	}
	p, err := project.Open(path)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(p)
	return string(data), err
}
func (d *Desktop) SaveProject(data string) (bool, error) {
	p, err := project.Decode([]byte(data), "json")
	if err != nil {
		return false, err
	}
	path, err := runtime.SaveFileDialog(d.ctx, runtime.SaveDialogOptions{Title: "Save HVAC project", DefaultFilename: "house.yaml", Filters: []runtime.FileFilter{{DisplayName: "YAML", Pattern: "*.yaml;*.yml"}, {DisplayName: "JSON", Pattern: "*.json"}}})
	if err != nil || path == "" {
		return false, err
	}
	if filepath.Ext(path) == "" {
		path += ".yaml"
	}
	encoded, err := project.Encode(*p, strings.TrimPrefix(filepath.Ext(path), "."))
	if err != nil {
		return false, err
	}
	// Write beside the destination, then atomically rename; preserve the old file
	// if writing fails. Temp files never contain credentials.
	f, err := os.CreateTemp(filepath.Dir(path), ".hvac-*.tmp")
	if err != nil {
		return false, err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, err = f.Write(encoded); err != nil {
		f.Close()
		return false, err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return false, err
	}
	if err = f.Close(); err != nil {
		return false, err
	}
	if err = os.Rename(tmp, path); err != nil {
		return false, err
	}
	return true, nil
}
func main() {
	d := &Desktop{}
	frontend, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		panic(err)
	}
	err = wails.Run(&options.App{Title: "Open Residential HVAC", Width: 1200, Height: 850, AssetServer: &assetserver.Options{Assets: frontend}, OnStartup: func(ctx context.Context) { d.ctx = ctx }, Bind: []interface{}{d}})
	if err != nil {
		panic(err)
	}
}
