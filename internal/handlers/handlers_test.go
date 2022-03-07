package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/yalagtyarzh/leafsite/internal/models"
)

func TestHandlers(t *testing.T) {
	var theTests = []struct {
		name               string
		url                string
		method             string
		expectedStatusCode int
	}{
		{
			name:               "home",
			url:                "/",
			method:             "GET",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "about",
			url:                "/about",
			method:             "GET",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "gq",
			url:                "/generals-quarters",
			method:             "GET",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "ms",
			url:                "/majors-suite",
			method:             "GET",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "sa",
			url:                "/search-availability",
			method:             "GET",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "contact",
			url:                "/contact",
			method:             "GET",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "non-existent",
			url:                "/aboba",
			method:             "GET",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "login",
			url:                "/user/login",
			method:             "GET",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "logout",
			url:                "/user/logout",
			method:             "GET",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "dashboard",
			url:                "/admin/dashboard",
			method:             "GET",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "new res",
			url:                "/admin/reservations-new",
			method:             "GET",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "all res",
			url:                "/admin/reservations-all",
			method:             "GET",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "show res",
			url:                "/admin/reservations/1/2/show",
			method:             "GET",
			expectedStatusCode: http.StatusOK,
		},
	}
	routes := getRoutes()
	ts := httptest.NewServer(routes)
	defer ts.Close()

	for _, tt := range theTests {
		resp, err := ts.Client().Get(ts.URL + tt.url)
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}

		if resp.StatusCode != tt.expectedStatusCode {
			t.Errorf("for %s, expected %d but got %d", tt.name, tt.expectedStatusCode, resp.StatusCode)
		}
	}
}

func TestRepository_Reservation(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/make-reservation", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.Reservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
	}

	//test case where reservation is not in session
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	//test with non-existent room
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	reservation.RoomID = 100

	rr = httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}
}

func TestRepository_PostAvailabilities(t *testing.T) {
	var theTests = []struct {
		name               string
		start              string
		end                string
		expectedStatusCode int
	}{
		{
			name:               "Rooms aren't available",
			start:              "2030-01-01",
			end:                "2030-01-02",
			expectedStatusCode: http.StatusSeeOther,
		},
		{
			name:               "Rooms are available",
			start:              "2020-01-01",
			end:                "2020-01-02",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Missing post body",
			expectedStatusCode: http.StatusSeeOther,
		},
		{
			name:               "Invalid start date",
			start:              "invalid",
			end:                "2030-01-02",
			expectedStatusCode: http.StatusSeeOther,
		},
		{
			name:               "Invalid end date",
			start:              "2020-01-01",
			end:                "invalid",
			expectedStatusCode: http.StatusSeeOther,
		},
		{
			name:               "Database query error",
			start:              "2040-01-01",
			end:                "2040-01-02",
			expectedStatusCode: http.StatusSeeOther,
		},
	}

	for _, tt := range theTests {
		postedData := url.Values{}
		var req *http.Request

		if tt.name == "Missing post body" {
			req, _ = http.NewRequest("POST", "/make-reservation", nil)
		} else {
			postedData.Add("start", tt.start)
			postedData.Add("end", tt.end)

			req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(postedData.Encode()))
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.PostAvailability)

		handler.ServeHTTP(rr, req)

		if rr.Code != tt.expectedStatusCode {
			t.Errorf("Post availability when no rooms available gave wrong status code: got %d, wanted %d", rr.Code, tt.expectedStatusCode)
		}
	}
}

func TestRepository_PostReservation(t *testing.T) {
	var theTests = []struct {
		name               string
		startDate          string
		endDate            string
		firstName          string
		lastName           string
		email              string
		phone              string
		roomID             string
		expectedStatusCode int
	}{
		{
			name:               "Ok",
			startDate:          "2030-01-01",
			endDate:            "2030-01-02",
			firstName:          "Alister",
			lastName:           "Azimuth",
			email:              "silhouetteAG@gmail.com",
			phone:              "7777777777",
			roomID:             "1",
			expectedStatusCode: http.StatusSeeOther,
		},
		{
			name:               "Missing post body",
			expectedStatusCode: http.StatusSeeOther,
		},
		{
			name:               "Invalid start date",
			startDate:          "invalid",
			endDate:            "2030-01-02",
			firstName:          "Alister",
			lastName:           "Azimuth",
			email:              "silhouetteAG@gmail.com",
			phone:              "7777777777",
			roomID:             "1",
			expectedStatusCode: http.StatusSeeOther,
		},
		{
			name:               "Invalid end date",
			startDate:          "2030-01-01",
			endDate:            "invalid",
			firstName:          "Alister",
			lastName:           "Azimuth",
			email:              "silhouetteAG@gmail.com",
			phone:              "7777777777",
			roomID:             "1",
			expectedStatusCode: http.StatusSeeOther,
		},
		{
			name:               "Invalid room id",
			startDate:          "2030-01-01",
			endDate:            "2030-01-02",
			firstName:          "Alister",
			lastName:           "Azimuth",
			email:              "silhouetteAG@gmail.com",
			phone:              "7777777777",
			roomID:             "invalid",
			expectedStatusCode: http.StatusSeeOther,
		},
		{
			name:               "Invalid data",
			startDate:          "2030-01-01",
			endDate:            "2030-01-02",
			firstName:          "D",
			lastName:           "Azimuth",
			email:              "silhouetteAG@gmail.com",
			phone:              "7777777777",
			roomID:             "1",
			expectedStatusCode: http.StatusSeeOther,
		},
		{
			name:               "Failure to insert reservation",
			startDate:          "2030-01-01",
			endDate:            "2030-01-02",
			firstName:          "Alister",
			lastName:           "Azimuth",
			email:              "silhouetteAG@gmail.com",
			phone:              "7777777777",
			roomID:             "2",
			expectedStatusCode: http.StatusSeeOther,
		},
		{
			name:               "Failure to insert restriction",
			startDate:          "2030-01-01",
			endDate:            "2030-01-02",
			firstName:          "Alister",
			lastName:           "Azimuth",
			email:              "silhouetteAG@gmail.com",
			phone:              "7777777777",
			roomID:             "1000",
			expectedStatusCode: http.StatusSeeOther,
		},
	}

	for _, tt := range theTests {
		postedData := url.Values{}
		var req *http.Request

		if tt.name == "Missing post body" {
			req, _ = http.NewRequest("POST", "/make-reservation", nil)
		} else {
			postedData.Add("start_date", tt.startDate)
			postedData.Add("end_date", tt.endDate)
			postedData.Add("first_name", tt.firstName)
			postedData.Add("last_name", tt.lastName)
			postedData.Add("email", tt.email)
			postedData.Add("phone", tt.phone)
			postedData.Add("room_id", tt.roomID)

			req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.PostReservation)

		handler.ServeHTTP(rr, req)

		if rr.Code != tt.expectedStatusCode {
			t.Errorf("PostReservation failed \"%s\" test: got %d, wanted %d", tt.name, rr.Code, tt.expectedStatusCode)
		}
	}
}

func TestRepository_AvailabilityJSON(t *testing.T) {
	var theTests = []struct {
		name            string
		startDate       string
		endDate         string
		isAvailable     bool
		expectedMessage string
	}{
		{
			name:            "Rooms are not available",
			startDate:       "2030-01-01",
			endDate:         "2030-01-02",
			isAvailable:     false,
			expectedMessage: "",
		},
		{
			name:            "Rooms are available",
			startDate:       "2029-01-01",
			endDate:         "2029-01-02",
			isAvailable:     true,
			expectedMessage: "",
		},
		{
			name:            "Missing post body",
			isAvailable:     false,
			expectedMessage: "Internal server error",
		},
		{
			name:            "Database error",
			startDate:       "2040-01-01",
			endDate:         "2040-01-02",
			isAvailable:     false,
			expectedMessage: "Error connecting to database",
		},
	}

	for _, tt := range theTests {
		postedData := url.Values{}
		var req *http.Request

		if tt.name == "Missing post body" {
			req, _ = http.NewRequest("POST", "/search-availability-json", nil)
		} else {
			postedData.Add("start", tt.startDate)
			postedData.Add("end", tt.endDate)
			postedData.Add("room_id", "1")

			req, _ = http.NewRequest("POST", "/search-availability-json", strings.NewReader(postedData.Encode()))
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AvailabilityJSON)

		handler.ServeHTTP(rr, req)

		var j jsonResponse
		err := json.Unmarshal(rr.Body.Bytes(), &j)
		if err != nil {
			t.Error("failed to parse json")
		}

		if j.OK != tt.isAvailable || j.Message != tt.expectedMessage {
			t.Errorf("Got \"%t\" availability when expected \"%t\", got \"%s\" message when expected \"%s\"", j.OK, tt.isAvailable, j.Message, tt.expectedMessage)
		}
	}
}

func TestRepository_ChooseRoom(t *testing.T) {
	var theTests = []struct {
		name               string
		roomID             string
		expectedStatusCode int
	}{
		{
			name:               "Ok",
			roomID:             "1",
			expectedStatusCode: http.StatusSeeOther,
		},
		{
			name:               "Reservation not in session",
			roomID:             "1",
			expectedStatusCode: http.StatusSeeOther,
		},
		{
			name:               "Missing url parameter",
			roomID:             "ei",
			expectedStatusCode: http.StatusSeeOther,
		},
	}

	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	for _, tt := range theTests {
		req, _ := http.NewRequest("GET", "/choose-room/1", nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.RequestURI = fmt.Sprintf("/choose-room/%s", tt.roomID)

		rr := httptest.NewRecorder()

		if tt.name == "Ok" {
			session.Put(ctx, "reservation", reservation)
		}

		handler := http.HandlerFunc(Repo.ChooseRoom)

		handler.ServeHTTP(rr, req)

		if rr.Code != tt.expectedStatusCode {
			t.Errorf("ChooseRoom handler returned wrong response code: got %d, wanted %d", rr.Code, tt.expectedStatusCode)
		}
	}
}

func TestRepository_BookRoom(t *testing.T) {
	var theTests = []struct {
		name               string
		id                 string
		expectedStatusCode int
	}{
		{
			name:               "Ok",
			id:                 "id=1",
			expectedStatusCode: http.StatusSeeOther,
		},
		{
			name:               "Database error",
			id:                 "id=4",
			expectedStatusCode: http.StatusSeeOther,
		},
	}

	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "Generals' Quarters",
		},
	}

	for _, tt := range theTests {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/book-room?s=2040-01-01&e=2040-01-02&%s", tt.id), nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		session.Put(ctx, "reservation", reservation)

		handler := http.HandlerFunc(Repo.BookRoom)

		handler.ServeHTTP(rr, req)

		if rr.Code != tt.expectedStatusCode {
			t.Errorf("BookRoom handler returned wrong response code: got %d, wanted %d", rr.Code, tt.expectedStatusCode)
		}
	}
}

func TestRepository_ReservationSummary(t *testing.T) {
	var theTests = []struct {
		name               string
		expectedStatusCode int
	}{
		{
			name:               "Ok",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Reservation is not in session",
			expectedStatusCode: http.StatusSeeOther,
		},
	}

	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "Generals' Quarters",
		},
	}

	for _, tt := range theTests {
		req, _ := http.NewRequest("GET", "/reservation-summary", nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		if tt.name == "Ok" {
			session.Put(ctx, "reservation", reservation)
		}

		handler := http.HandlerFunc(Repo.ReservationSummary)

		handler.ServeHTTP(rr, req)

		if rr.Code != tt.expectedStatusCode {
			t.Errorf("ReservationSummary handler returned wrong response code: got %d, wanted %d", rr.Code, tt.expectedStatusCode)
		}
	}
}

func TestRepository_Login(t *testing.T) {
	var theTests = []struct {
		name               string
		email              string
		expectedStatusCode int
		expectedHTML       string
		expectedLocation   string
	}{
		{
			name:               "valid-credentials",
			email:              "sera@gmail.com",
			expectedStatusCode: http.StatusSeeOther,
			expectedHTML:       "",
			expectedLocation:   "/",
		},
		{
			name:               "invalid-credentials",
			email:              "wasnever@pepega.meme",
			expectedStatusCode: http.StatusSeeOther,
			expectedHTML:       "",
			expectedLocation:   "/user/login",
		},
		{
			name:               "invalid-data",
			email:              "?",
			expectedStatusCode: http.StatusOK,
			expectedHTML:       `action="/user/login"`,
			expectedLocation:   "",
		},
	}

	for _, tt := range theTests {
		postedData := url.Values{}
		postedData.Add("email", tt.email)
		postedData.Add("password", "password")

		req, _ := http.NewRequest("POST", "/user/login", strings.NewReader(postedData.Encode()))
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.PostShowLogin)
		handler.ServeHTTP(rr, req)

		if rr.Code != tt.expectedStatusCode {
			t.Errorf("failed %s, expected code %d, but got %d", tt.name, tt.expectedStatusCode, rr.Code)
		}

		if tt.expectedLocation != "" {
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != tt.expectedLocation {
				t.Errorf("failed %s, expected location %s, but got location %s", tt.name, tt.expectedLocation, actualLoc.String())
			}
		}

		if tt.expectedHTML != "" {
			html := rr.Body.String()
			if !strings.Contains(html, tt.expectedHTML) {
				t.Errorf("failed %s, expected to find %s, but got %s", tt.name, tt.expectedHTML, rr.Body.String())
			}
		}
	}
}

func TestRepository_AdminPostShowReservation(t *testing.T) {
	var theTests = []struct {
		name               string
		url                string
		postedData         url.Values
		expectedStatusCode int
		expectedLocation   string
		expectedHTML       string
	}{
		{
			name: "valid-data-from-new",
			url:  "/admin/reservations/new/1/show",
			postedData: url.Values{
				"first_name": {"Kaedehara"},
				"last_name":  {"Kazuha"},
				"email":      {"LeavesintheWind@gmail.com"},
				"phone":      {"8-800-555-35-35"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation:   "/admin/reservations-new",
			expectedHTML:       "",
		},
		{
			name: "valid-data-from-all",
			url:  "/admin/reservations/all/1/show",
			postedData: url.Values{
				"first_name": {"Kaedehara"},
				"last_name":  {"Kazuha"},
				"email":      {"LeavesintheWind@gmail.com"},
				"phone":      {"8-800-555-35-35"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation:   "/admin/reservations-all",
			expectedHTML:       "",
		},
		{
			name: "valid-data-from-cal",
			url:  "/admin/reservations/cal/1/show",
			postedData: url.Values{
				"first_name": {"Kaedehara"},
				"last_name":  {"Kazuha"},
				"email":      {"LeavesintheWind@gmail.com"},
				"phone":      {"8-800-555-35-35"},
				"year":       {"2022"},
				"month":      {"03"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation:   "/admin/reservations-calendar?y=2022&m=03",
			expectedHTML:       "",
		},
	}

	for _, tt := range theTests {
		var req *http.Request
		if tt.postedData != nil {
			req, _ = http.NewRequest("POST", "/user/login", strings.NewReader(tt.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/user/login", nil)
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.RequestURI = tt.url

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AdminPostShowReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != tt.expectedStatusCode {
			t.Errorf("failed %s, expected code %d, but got %d", tt.name, tt.expectedStatusCode, rr.Code)
		}

		if tt.expectedLocation != "" {
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != tt.expectedLocation {
				t.Errorf("failed %s: expected location %s, but got location %s", tt.name, tt.expectedLocation, actualLoc.String())
			}
		}

		if tt.expectedHTML != "" {
			html := rr.Body.String()
			if !strings.Contains(html, tt.expectedHTML) {
				t.Errorf("failed %s: expected to find %s but did not", tt.name, tt.expectedHTML)
			}
		}
	}
}

func TestRepository_AdminPostReservationCalendar(t *testing.T) {
	var theTests = []struct {
		name               string
		postedData         url.Values
		expectedStatusCode int
		expectedLocation   string
		expectedHTML       string
		blocks             int
		reservations       int
	}{
		{
			name: "cal",
			postedData: url.Values{
				"year":  {time.Now().Format("2006")},
				"month": {time.Now().Format("01")},
				fmt.Sprintf("add_block_1_%s", time.Now().AddDate(0, 0, 2).Format("2006-01-2")): {"1"},
			},
			expectedStatusCode: http.StatusSeeOther,
		},
		{
			name:               "cal-blocks",
			postedData:         url.Values{},
			expectedStatusCode: http.StatusSeeOther,
			blocks:             1,
		},
		{
			name:               "cal-res",
			postedData:         url.Values{},
			expectedStatusCode: http.StatusSeeOther,
			reservations:       1,
		},
	}

	for _, tt := range theTests {
		var req *http.Request
		if tt.postedData != nil {
			req, _ = http.NewRequest("POST", "/admin/reservations-calendar", strings.NewReader(tt.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/admin/reservations-calendar", nil)
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		now := time.Now()
		bm := make(map[string]int)
		rm := make(map[string]int)

		currentYear, currentMonth, _ := now.Date()
		currentLocation := now.Location()

		firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
		lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

		for d := firstOfMonth; !d.After(lastOfMonth); d = d.AddDate(0, 0, 1) {
			rm[d.Format("2006-01-02")] = 0
			bm[d.Format("2006-01-02")] = 0
		}

		if tt.blocks > 0 {
			bm[firstOfMonth.Format("2006-01-2")] = tt.blocks
		}

		if tt.reservations > 0 {
			rm[lastOfMonth.Format("2006-01-2")] = tt.reservations
		}

		session.Put(ctx, "block_map_1", bm)
		session.Put(ctx, "reservation_map_1", rm)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AdminPostReservationsCalendar)
		handler.ServeHTTP(rr, req)

		if rr.Code != tt.expectedStatusCode {
			t.Errorf("failed %s: expected code %d, but got %d", tt.name, tt.expectedStatusCode, rr.Code)
		}
	}
}

func TestRepository_AdminProcessReservation(t *testing.T) {
	var theTests = []struct {
		name               string
		queryParams        string
		expectedStatusCode int
		expectedLocation   string
	}{
		{
			name:               "process-reservation",
			queryParams:        "",
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation:   "",
		},
		{
			name:               "process-reservation-back-to-cal",
			queryParams:        "?y=2021&m=12",
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation:   "",
		},
	}

	for _, tt := range theTests {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/admin/process-reservation/cal/1/do%s", tt.queryParams), nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AdminProcessReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusSeeOther {
			t.Errorf("failed %s: expected status code %d, but got %d", tt.name, tt.expectedStatusCode, rr.Code)
		}
	}
}

func TestRepository_AdminDeleteReservation(t *testing.T) {
	var theTests = []struct {
		name               string
		queryParams        string
		expectedStatusCode int
		expectedLocation   string
	}{
		{
			name:               "delete-reservation",
			queryParams:        "",
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation:   "",
		},
		{
			name:               "delete_reservation-back-to-cal",
			queryParams:        "?y=2021&m=12",
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation:   "",
		},
	}

	for _, tt := range theTests {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/admin/delete-reservation/cal/1/do%s", tt.queryParams), nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AdminDeleteReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusSeeOther {
			t.Errorf("failed %s: expected code %d, but got %d", tt.name, tt.expectedStatusCode, rr.Code)
		}
	}
}

func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}

	return ctx
}
