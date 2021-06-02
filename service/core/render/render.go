package render

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bogdanrat/web-server/contracts/models"
	"github.com/bogdanrat/web-server/service/core/config"
	"github.com/bogdanrat/web-server/service/core/lib"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

const (
	pathToTemplates = "./templates"
)

var (
	functions = template.FuncMap{
		"formatSize": lib.FormatSize,
	}
)

func Template(w http.ResponseWriter, r *http.Request, tmpl string, templateData *models.TemplateData) error {
	var templateCache map[string]*template.Template
	var err error

	if !config.AppConfig.Server.DevelopmentMode {
		templateCache = config.AppConfig.TemplateCache
	} else {
		templateCache, err = CreateTemplateCache()
		if err != nil {
			log.Println("cannot create template cache:", err.Error())
			return errors.New(fmt.Sprintf("cannot create template cache: %s", err.Error()))
		}
	}

	t, ok := templateCache[tmpl]
	if !ok {
		log.Println("template not found in cache")
		return errors.New("template not found in cache")
	}

	buf := new(bytes.Buffer)
	_ = t.Execute(buf, templateData)
	_, err = buf.WriteTo(w)

	if err != nil {
		log.Println("error writing template: ", err)
		return errors.New(fmt.Sprintf("error writing template: %s", err.Error()))
	}

	w.WriteHeader(http.StatusOK)

	return nil
}

func CreateTemplateCache() (map[string]*template.Template, error) {
	templateCache := map[string]*template.Template{}

	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {
			return nil, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
			if err != nil {
				return nil, err
			}
		}

		templateCache[name] = ts
	}

	return templateCache, nil
}
