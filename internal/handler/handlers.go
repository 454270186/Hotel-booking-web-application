package handler

import (
	"encoding/json"
	"fmt"
	"github.com/454270186/Hotel-booking-web-application/internal/Models"
	"github.com/454270186/Hotel-booking-web-application/internal/config"
	"github.com/454270186/Hotel-booking-web-application/internal/driver"
	"github.com/454270186/Hotel-booking-web-application/internal/forms"
	"github.com/454270186/Hotel-booking-web-application/internal/helpers"
	"github.com/454270186/Hotel-booking-web-application/internal/render"
	"github.com/454270186/Hotel-booking-web-application/internal/repository"
	"github.com/454270186/Hotel-booking-web-application/internal/repository/dbrepo"
	"net/http"
)

var Repo *Repository

type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(a, db.SQL),
	}
}

// NewHandler sets the repository to the handler
func NewHandler(r *Repository) {
	Repo = r
}

// Home is the home page handler
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	_ = render.Template(w, r, "home.page.html", &Models.TemplateData{})
}

// About is the about page handler
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	// render the template
	_ = render.Template(w, r, "about.page.html", &Models.TemplateData{})
}

// Reservation renders a make-reservation page and display a form
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	var emptyReservation Models.Reservation
	data := make(map[string]interface{})
	data["reservation"] = emptyReservation
	// initialize empty data and form-data to make-reservation page
	// so that it can display blank in every input when first time get in this page
	_ = render.Template(w, r, "make-reservation.page.html", &Models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm() // 获得表单post的数据
	if err != nil {
		helpers.ServeError(w, err)
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

	// Backend validation
	form.Required("first_name", "last_name", "email") // check input is blank or not
	form.MinLength("first_name", 5)                   // specific validation for the first_name
	form.IsEmail("email")                             // check if email address is valid or not

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation // store the reservation-data and pass it to template

		// re-render this page to show some error
		_ = render.Template(w, r, "make-reservation.page.html", &Models.TemplateData{
			Form: form,
			Data: data,
		})

		return
	}

	// if all input is validated, store the input in Session which is for reservation-summary page to use
	m.App.Session.Put(r.Context(), "reservation", reservation)

	// In order to avoid people from accidentally submitting the form twice
	// Everytime receive a POST request, should direct Users to another page
	http.Redirect(w, r, "reservation-summary", http.StatusSeeOther)
}

// Generals renders the room page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	_ = render.Template(w, r, "generals.page.html", &Models.TemplateData{})
}

// Majors renders the room page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	_ = render.Template(w, r, "majors.page.html", &Models.TemplateData{})
}

// Availability renders the Book Now page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	_ = render.Template(w, r, "search-availability.page.html", &Models.TemplateData{})
}

// PostAvailability handle the post from form in search-availability page
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	_, _ = w.Write([]byte(fmt.Sprintf("The start date is %s and The end date is %s", start, end)))
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
		helpers.ServeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// Contact renders the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	_ = render.Template(w, r, "contact.page.html", &Models.TemplateData{})
}

// ReservationSummary renders the reservation-summary page
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(Models.Reservation)
	if !ok {
		m.App.ErrorLog.Println("Cannot get the reservation information items.")
		m.App.Session.Put(r.Context(), "error", "Error to get reservation form session.")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// when have got the reservation in Session, remove it from Session
	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	_ = render.Template(w, r, "reservation-summary.page.html", &Models.TemplateData{
		Data: data,
	})
}
