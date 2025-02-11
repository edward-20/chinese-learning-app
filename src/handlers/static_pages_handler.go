package handler

import (
	"fmt"
	"net/http"

	"github.com/edward-20/chinese-learning-app/templates"
	"github.com/edward-20/chinese-learning-app/utils"
)

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("/about")
	utils.RenderTemplate(w, templates.AboutTemplate, nil)
}

func ContactHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("/contact")
	utils.RenderTemplate(w, templates.ContactTemplate, nil)
}
