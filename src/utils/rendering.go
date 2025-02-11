package utils

import (
	"html/template"
	"net/http"
)

func RenderTemplate(w http.ResponseWriter, temp *template.Template, data any) {
	w.Header().Set("Content-Type", "text/html")
	err := temp.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		err = temp.Execute(w, data)
		if err == nil {
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
