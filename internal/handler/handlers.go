package handler

import (
	"encoding/json"
	"fmt"
	"github.com/454270186/Hotel-booking-web-application/internal/Models"
	"github.com/454270186/Hotel-booking-web-application/internal/config"
	"github.com/454270186/Hotel-booking-web-application/internal/forms"
	"github.com/454270186/Hotel-booking-web-application/internal/render"
	"log"
	"net/http"
)

var Repo *Repository

type Repository struct {
	App *config.AppConfig
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

// NewHandler sets the repository to the handler
func NewHandler(r *Repository) {
	Repo = r
}

// Home is the home page handler
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	remoteIP := r.RemoteAddr
	m.App.Session.Put(r.Context(), "remote_ip", remoteIP)

	render.RenderTemplate(w, r, "home.page.html", &Models.TemplateData{})
}

// About is the about page handler
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	// perform some logic
	stringMap := make(map[string]string)
	stringMap["test"] = "Hello, again!"

	remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIP

	// render the template
	render.RenderTemplate(w, r, "about.page.html", &Models.TemplateData{
		StringMap: stringMap,
	})
}

// Reservation renders a make reservation page and display a form
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "make-reservation.page.html", &Models.TemplateData{
		Form: forms.New(nil), // associates with the form in make-reservation page
	})
}

func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm() // 获得表单post的数据
	if err != nil {
		log.Println(err)
		return
	}

	// store the form data
	reservation := Models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
	}

	// store the data posted by form
	form := forms.New(r.PostForm)

	form.Has("first_name", r)

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation // store the reservation-data and pass it to template

		// re-render this page
		render.RenderTemplate(w, r, "make-reservation.page.html", &Models.TemplateData{
			Form: form,
			Data: data,
		})

		return
	}
}

// Generals renders the room page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "generals.page.html", &Models.TemplateData{})
}

// Majors renders the room page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "majors.page.html", &Models.TemplateData{})
}

// Availability renders the Book Now page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "search-availability.page.html", &Models.TemplateData{})
}

// PostAvailability handle the post from form in search-availability page
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	w.Write([]byte(fmt.Sprintf("The start date is %s and The end date is %s", start, end)))
}

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// AvailabilityJSON handle request to Availability and send JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	resp := jsonResponse{
		OK:      true,
		Message: "Available!",
	}

	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// Contact renders the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "contact.page.html", &Models.TemplateData{})
}
