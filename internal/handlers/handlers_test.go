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

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//test with non-existent room
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	reservation.RoomID = 100

	rr = httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
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
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
		{
			name:               "Invalid start date",
			start:              "invalid",
			end:                "2030-01-02",
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
		{
			name:               "Invalid end date",
			start:              "2020-01-01",
			end:                "invalid",
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
		{
			name:               "Database query error",
			start:              "2040-01-01",
			end:                "2040-01-02",
			expectedStatusCode: http.StatusTemporaryRedirect,
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
			expectedStatusCode: http.StatusTemporaryRedirect,
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
			expectedStatusCode: http.StatusTemporaryRedirect,
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
			expectedStatusCode: http.StatusTemporaryRedirect,
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
			expectedStatusCode: http.StatusTemporaryRedirect,
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
			expectedStatusCode: http.StatusTemporaryRedirect,
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
			expectedStatusCode: http.StatusTemporaryRedirect,
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
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
		{
			name:               "Missing url parameter",
			roomID:             "ei",
			expectedStatusCode: http.StatusTemporaryRedirect,
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
			expectedStatusCode: http.StatusTemporaryRedirect,
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
			expectedStatusCode: http.StatusTemporaryRedirect,
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

func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}

	return ctx
}
