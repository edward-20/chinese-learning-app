package templates

import "html/template"

var AboutTemplate = template.Must(template.ParseFiles("templates/base.html", "templates/about.html"))
var ContactTemplate = template.Must(template.ParseFiles("templates/base.html", "templates/contact.html"))
var StartTestTemplate = template.Must(template.ParseFiles("templates/base.html", "templates/test-start.html"))
var ResumeTestTemplate = template.Must(template.ParseFiles("templates/base.html", "templates/test-resume.html"))
var TestQuestionTemplate = template.Must(template.ParseFiles("templates/single-character-question.html"))
var TestSolutionTemplate = template.Must(template.ParseFiles("templates/check-answer.html"))
var TestReviewTemplate = template.Must(template.ParseFiles("templates/review.html"))
