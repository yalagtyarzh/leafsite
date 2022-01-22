package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yalagtyarzh/leafsite/internal/models"
)

type postData struct {
	key   string
	value string
}

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
		// {
		// 	name:               "ma",
		// 	url:                "/make-reservation",
		// 	method:             "GET",
		// 	params:             []postData{},
		// 	expectedStatusCode: http.StatusOK,
		// },
		// {
		// 	name:               "rs",
		// 	url:                "/reservation-summary",
		// 	method:             "GET",
		// 	params:             []postData{},
		// 	expectedStatusCode: http.StatusOK,
		// },
		// {
		// 	name:   "post-sa",
		// 	url:    "/search-availability",
		// 	method: "POST",
		// 	params: []postData{
		// 		{key: "start", value: "2021-12-12"},
		// 		{key: "end", value: "2021-12-13"},
		// 	},
		// 	expectedStatusCode: http.StatusOK,
		// },
		// {
		// 	name:   "post-sa-json",
		// 	url:    "/search-availability-json",
		// 	method: "POST",
		// 	params: []postData{
		// 		{key: "start", value: "2021-12-12"},
		// 		{key: "end", value: "2021-12-13"},
		// 	},
		// 	expectedStatusCode: http.StatusOK,
		// },
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
			startDate:          "start_date=2030-01-01",
			endDate:            "end_date=2030-01-02",
			firstName:          "first_name=Alister",
			lastName:           "last_name=Azimuth",
			email:              "email=silhouetteAG@gmail.com",
			phone:              "phone=7777777777",
			roomID:             "room_id=1",
			expectedStatusCode: http.StatusSeeOther,
		},
		{
			name:               "Missing post body",
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
		{
			name:               "Invalid start date",
			startDate:          "start_date=invalid",
			endDate:            "end_date=2030-01-02",
			firstName:          "first_name=Alister",
			lastName:           "last_name=Azimuth",
			email:              "email=silhouetteAG@gmail.com",
			phone:              "phone=7777777777",
			roomID:             "room_id=1",
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
		{
			name:               "Invalid end date",
			startDate:          "start_date=2030-01-01",
			endDate:            "end_date=invalid",
			firstName:          "first_name=Alister",
			lastName:           "last_name=Azimuth",
			email:              "email=silhouetteAG@gmail.com",
			phone:              "phone=7777777777",
			roomID:             "room_id=1",
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
		{
			name:               "Invalid room id",
			startDate:          "start_date=2030-01-01",
			endDate:            "end_date=2030-01-02",
			firstName:          "first_name=Alister",
			lastName:           "last_name=Azimuth",
			email:              "email=silhouetteAG@gmail.com",
			phone:              "phone=7777777777",
			roomID:             "room_id=invalid",
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
		{
			name:               "Invalid data",
			startDate:          "start_date=2030-01-01",
			endDate:            "end_date=2030-01-02",
			firstName:          "first_name=D",
			lastName:           "last_name=Azimuth",
			email:              "email=silhouetteAG@gmail.com",
			phone:              "phone=7777777777",
			roomID:             "room_id=1",
			expectedStatusCode: http.StatusSeeOther,
		},
		{
			name:               "Failure to insert reservation",
			startDate:          "start_date=2030-01-01",
			endDate:            "end_date=2030-01-02",
			firstName:          "first_name=Alister",
			lastName:           "last_name=Azimuth",
			email:              "email=silhouetteAG@gmail.com",
			phone:              "phone=7777777777",
			roomID:             "room_id=2",
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
		{
			name:               "Failure to insert restriction",
			startDate:          "start_date=2030-01-01",
			endDate:            "end_date=2030-01-02",
			firstName:          "first_name=Alister",
			lastName:           "last_name=Azimuth",
			email:              "email=silhouetteAG@gmail.com",
			phone:              "phone=7777777777",
			roomID:             "room_id=2",
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
	}

	for _, tt := range theTests {
		var reqBody string
		var req *http.Request

		if tt.name == "Missing post body" {
			req, _ = http.NewRequest("POST", "/make-reservation", nil)
		} else {
			reqBody = tt.startDate
			reqBody = fmt.Sprintf("%s&%s", reqBody, tt.endDate)
			reqBody = fmt.Sprintf("%s&%s", reqBody, tt.firstName)
			reqBody = fmt.Sprintf("%s&%s", reqBody, tt.lastName)
			reqBody = fmt.Sprintf("%s&%s", reqBody, tt.email)
			reqBody = fmt.Sprintf("%s&%s", reqBody, tt.phone)
			reqBody = fmt.Sprintf("%s&%s", reqBody, tt.roomID)

			req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
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
			startDate:       "start=2030-01-01",
			endDate:         "end=2030-01-02",
			isAvailable:     false,
			expectedMessage: "",
		},
		{
			name:            "Rooms are available",
			startDate:       "start=2029-01-01",
			endDate:         "end=2029-01-02",
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
			startDate:       "start=2040-01-01",
			endDate:         "end=2040-01-02",
			isAvailable:     false,
			expectedMessage: "Error connecting to database",
		},
	}

	for _, tt := range theTests {
		var reqBody string
		var req *http.Request
		if tt.name == "Missing post body" {
			req, _ = http.NewRequest("POST", "/search-availability-json", nil)
		} else {
			reqBody = tt.startDate
			reqBody = fmt.Sprintf("%s&%s", reqBody, tt.endDate)
			reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")

			req, _ = http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody))
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

func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}

	return ctx
}
