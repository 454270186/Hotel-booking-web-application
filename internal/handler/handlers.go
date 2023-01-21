package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/454270186/Hotel-booking-web-application/internal/Models"
	"github.com/454270186/Hotel-booking-web-application/internal/config"
	"github.com/454270186/Hotel-booking-web-application/internal/driver"
	"github.com/454270186/Hotel-booking-web-application/internal/forms"
	"github.com/454270186/Hotel-booking-web-application/internal/helpers"
	"github.com/454270186/Hotel-booking-web-application/internal/render"
	"github.com/454270186/Hotel-booking-web-application/internal/repository"
	"github.com/454270186/Hotel-booking-web-application/internal/repository/dbrepo"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"time"
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

// NewTestRepo creates a new repository for test
func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestingRepo(a),
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
	res, ok := m.App.Session.Get(r.Context(), "reservation").(Models.Reservation)
	if !ok {
		helpers.ServeError(w, errors.New("cannot get reservation from Session"))
		return
	}

	// populate room name
	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		helpers.ServeError(w, err)
	}
	res.Room.RoomName = room.RoomName
	m.App.Session.Put(r.Context(), "reservation", res)

	// parse time-object to string
	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = res
	// initialize empty data and form-data to make-reservation page
	// so that it can display blank in every input when first time get in this page
	_ = render.Template(w, r, "make-reservation.page.html", &Models.TemplateData{
		Data:      data,
		Form:      forms.New(nil),
		StringMap: stringMap,
	})
}

func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(Models.Reservation)
	if !ok {
		helpers.ServeError(w, errors.New("cannot get reservation from Session"))
	}
	err := r.ParseForm() // 获得表单post的数据
	if err != nil {
		helpers.ServeError(w, err)
		return
	}

	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")

	//// parse string to time-object
	//layout := "2006-01-02"
	//startDate, err := time.Parse(layout, sd)
	//if err != nil {
	//	helpers.ServeError(w, err)
	//	return
	//}
	//endDate, err := time.Parse(layout, ed)
	//if err != nil {
	//	helpers.ServeError(w, err)
	//	return
	//}

	// convert room_id(string) to int
	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "invalid data!")
		return
	}

	// update the reservation got from Session
	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Email = r.Form.Get("email")
	reservation.Phone = r.Form.Get("phone")
	reservation.RoomID = roomID

	// store the data posted by form
	form := forms.New(r.PostForm)

	// Backend validation
	form.Required("first_name", "last_name", "email") // check input is blank or not
	form.MinLength("first_name", 5)                   // specific validation for the first_name
	form.IsEmail("email")                             // check if email address is valid or not

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation // store the reservation-data and pass it to template

		stringMap := make(map[string]string)
		stringMap["start_date"] = sd
		stringMap["end_date"] = ed

		// re-render this page to show some error
		_ = render.Template(w, r, "make-reservation.page.html", &Models.TemplateData{
			Form:      form,
			Data:      data,
			StringMap: stringMap,
		})

		return
	}

	// if form is valid, insert data into database
	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		helpers.ServeError(w, err)
		return
	}

	restriction := Models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        roomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}
	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		helpers.ServeError(w, err)
		return
	}

	// send notifications by email - first to guest
	htmlMSG := fmt.Sprintf(`
		<strong>Reservation Confirmation</strong><br>
		Dear %s:<br>
		This is confirm your reservation from %s to %s.
`, reservation.FirstName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msg := Models.MailData{
		To:      reservation.Email,
		From:    "me@here.com",
		Subject: "Reservation Confirmation",
		Content: htmlMSG,
	}
	m.App.MailChan <- msg

	// send notifications by email - second to property owner
	htmlMSGForOwner := fmt.Sprintf(`
	<strong>Reservation Confirmation</strong><br>
	You have a reservation of %s from %s to %s.
`, reservation.Room.RoomName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msgToOwner := Models.MailData{
		To:      "Owner@ow.com",
		From:    "me@here.com",
		Subject: "Reservation Notification",
		Content: htmlMSGForOwner,
	}
	m.App.MailChan <- msgToOwner

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

	// parse string to time-object
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, start)
	if err != nil {
		helpers.ServeError(w, err)
		return
	}
	endDate, err := time.Parse(layout, end)
	if err != nil {
		helpers.ServeError(w, err)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		helpers.ServeError(w, err)
	}

	// if not room is available
	if len(rooms) == 0 {
		m.App.Session.Put(r.Context(), "error", "No Availability")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	// if there is rooms available
	data := make(map[string]interface{})
	data["rooms"] = rooms

	// store the start date and end date berfore rendering this page
	res := Models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}
	m.App.Session.Put(r.Context(), "reservation", res)

	_ = render.Template(w, r, "choose-room.page.html", &Models.TemplateData{
		Data: data,
	})
}

type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// AvailabilityJSON handle request to Availability and send JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	// parse string to time.Time
	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		helpers.ServeError(w, err)
	}

	isAvailable, _ := m.DB.SearchAvailabilityByDateByRoomID(startDate, endDate, roomID)
	resp := jsonResponse{
		OK:        isAvailable,
		Message:   "",
		StartDate: sd,
		EndDate:   ed,
		RoomID:    strconv.Itoa(roomID),
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

	startDate := reservation.StartDate.Format("2006-01-02")
	endDate := reservation.EndDate.Format("2006-01-02")
	stringMap := make(map[string]string)
	stringMap["start_date"] = startDate
	stringMap["end_date"] = endDate

	_ = render.Template(w, r, "reservation-summary.page.html", &Models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

// ChooseRoom takes URL parameter of room_id, and pass it to make_reservation page
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	// using Chi helper function
	roomID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServeError(w, err)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(Models.Reservation)
	if !ok {
		helpers.ServeError(w, errors.New("Cannot get reservation from Session"))
		return
	}
	res.RoomID = roomID                                // update room id
	m.App.Session.Put(r.Context(), "reservation", res) // put it back into Session

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// BookRoom takes URL paramters, builds a sessional variable, and take users to make_reservation page
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	// put some data in 'res' and pass 'res' to make_reservation page for using
	var res Models.Reservation

	// grab id, s, e from URL
	roomID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	sd := r.URL.Query().Get("s")
	ed := r.URL.Query().Get("e")

	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		helpers.ServeError(w, err)
	}

	res.RoomID = roomID
	res.StartDate = startDate
	res.EndDate = endDate
	res.Room.RoomName = room.RoomName

	m.App.Session.Put(r.Context(), "reservation", res)
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}
