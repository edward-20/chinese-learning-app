package templates

import (
	"html/template"
	"log"
	"os"
)

var AboutTemplate, ContactTemplate, StartTestTemplate, ResumeTestTemplate, TestQuestionTemplate, TestSolutionTemplate, TestReviewTemplate *template.Template

func init() {
	mode := os.Getenv("CHINESE_LEARNING_APP_MODE")
	if "" == mode {
		mode = "develop"
	}

	var templatesDirectory string
	if mode == "production" || mode == "develop" {
		templatesDirectory = "./templates/"
	} else if mode == "test" {
		templatesDirectory = "../templates/"
	} else {
		log.Fatal("CHINESE_LEARNING_APP_MODE was not set in the OS")
	}

	AboutTemplate = template.Must(template.ParseFiles(templatesDirectory+"base.html", templatesDirectory+"about.html"))
	ContactTemplate = template.Must(template.ParseFiles(templatesDirectory+"base.html", templatesDirectory+"contact.html"))
	StartTestTemplate = template.Must(template.ParseFiles(templatesDirectory+"base.html", templatesDirectory+"test-start.html"))
	ResumeTestTemplate = template.Must(template.ParseFiles(templatesDirectory+"base.html", templatesDirectory+"test-resume.html"))
	TestQuestionTemplate = template.Must(template.ParseFiles(templatesDirectory + "single-character-question.html"))
	TestSolutionTemplate = template.Must(template.ParseFiles(templatesDirectory + "check-answer.html"))
	TestReviewTemplate = template.Must(template.ParseFiles(templatesDirectory + "review.html"))
}
