package handlers

import (
	"backend-form/m/internal/models"
	"html/template"
	"net/http"
)

type FormHandler struct {
	templates *template.Template
}

func NewFormHandler(templates *template.Template) *FormHandler {
	return &FormHandler{templates: templates}
}
func (h *FormHandler) ShowForm(w http.ResponseWriter, r *http.Request) {
	// Your existing form logic here
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if err := h.templates.ExecuteTemplate(w, "form.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (h *FormHandler) SubmitForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form submission", http.StatusBadRequest)
		return
	}
	user := models.User{
		Name:    r.FormValue("name"),
		Email:   r.FormValue("email"),
		Age:     r.FormValue("age"),
		Phone:   r.FormValue("phone"),
		Website: r.FormValue("website"),
		Message: r.FormValue("message"),
	}
	if err := user.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data := map[string]string{
		"Name":    user.Name,
		"Email":   user.Email,
		"Age":     user.Age,
		"Phone":   user.Phone,
		"Website": user.Website,
		"Message": user.Message,
	}

	if err := h.templates.ExecuteTemplate(w, "success.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
