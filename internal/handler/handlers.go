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
	"log"
	"net/http"
	"strconv"
	"strings"
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

// ShowLogin renders the login page
func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.html", &Models.TemplateData{
		Form: forms.New(nil),
	})
}

// PostShowLogin handles logging the user in
func (m *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	// for safe
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")
	if !form.Valid() {
		render.Template(w, r, "login.page.html", &Models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		log.Println(err)

		m.App.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id) // the key "user_id" is used to authenticate
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout logs user out
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// AdminDashboard shows the admin dashboard
func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.html", &Models.TemplateData{})
}

// AdminNewReservations shows all new reservations in admin tool
func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllNewReservations()
	if err != nil {
		helpers.ServeError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, r, "admin-new-reservations.page.html", &Models.TemplateData{
		Data: data,
	})
}

// AdminAllReservations shows all reservations in admin tool
func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllReservations()
	if err != nil {
		helpers.ServeError(w, err)
		return
	}
	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, r, "admin-all-reservations.page.html", &Models.TemplateData{
		Data: data,
	})
}

// AdminShowReservation shows the reservation detail
func (m *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")

	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServeError(w, err)
		return
	}

	src := exploded[3]
	stringMap := make(map[string]string)
	stringMap["src"] = src

	// get reservation from database
	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServeError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "admin-reservations-show.page.html", &Models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      forms.New(nil),
	})
}

// AdminPostShowReservation handles the post for reservation updates
func (m *Repository) AdminPostShowReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServeError(w, err)
		return
	}

	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServeError(w, err)
		return
	}
	src := exploded[3]

	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServeError(w, err)
		return
	}

	// update reservation
	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	err = m.DB.UpdateReservation(res)
	if err != nil {
		helpers.ServeError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Changes Saved")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)

}

// AdminReservationsCalendar displays the reservations calendar
func (m *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	// assume that there is no month/year specified
	now := time.Now()

	if r.URL.Query().Get("y") != "" {
		year, _ := strconv.Atoi(r.URL.Query().Get("y"))
		month, _ := strconv.Atoi(r.URL.Query().Get("m"))
		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	data := make(map[string]interface{})
	data["now"] = now

	next := now.AddDate(0, 1, 0)
	last := now.AddDate(0, -1, 0)

	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")

	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	stringMap := make(map[string]string)
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear

	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	// get the first and last days of the month
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	// get all rooms from database
	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServeError(w, err)
		return
	}
	data["rooms"] = rooms

	for _, room := range rooms {
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		// init these two map
		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format("2006-01-2")] = 0
			blockMap[d.Format("2006-01-2")] = 0
		}

		// get all restrictions for the current room
		restrictions, err := m.DB.GetRestrictionsForRoomByDate(room.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			helpers.ServeError(w, err)
			return
		}

		// loop over resrtictions and distribute them in either reservation or block
		for _, y := range restrictions {
			if y.ReservationID > 0 {
				// it's a reservation
				for d := y.StartDate; d.After(y.EndDate) == false; d = d.AddDate(0, 0, 1) {
					reservationMap[d.Format("2006-01-2")] = y.ReservationID
				}
			} else {
				// it's a block
				blockMap[y.StartDate.Format("2006-01-2")] = y.ID
			}
		}

		data[fmt.Sprintf("reservation_map_%d", room.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", room.ID)] = blockMap

		m.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", room.ID), blockMap)
	}

	render.Template(w, r, "admin-reservations-calendar.page.html", &Models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		IntMap:    intMap,
	})
}

// AdminProcessReservation marks a reservation as processed
func (m *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	err := m.DB.UpdateProcessedForReservation(id, 1)
	if err != nil {
		helpers.ServeError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Reservation marked as Processed")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
}

// AdminProcessReservation marks a reservation as processed
func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	err := m.DB.DeleteReservation(id)
	if err != nil {
		helpers.ServeError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Reservation Deleted")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
}
