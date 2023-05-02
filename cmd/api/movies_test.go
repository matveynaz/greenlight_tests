package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"greenlight.bcc/internal/assert"
)

func TestShowMovie(t *testing.T) {
	app := newTestApplication(t)

	ts := newTestServer(t, app.routesTest())
	defer ts.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody string
	}{
		{
			name:     "Valid ID",
			urlPath:  "/v1/movies/1",
			wantCode: http.StatusOK,
		},
		{
			name:     "Non-existent ID",
			urlPath:  "/v1/movies/2",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Negative ID",
			urlPath:  "/v1/movies/-1",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Decimal ID",
			urlPath:  "/v1/movies/1.23",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "String ID",
			urlPath:  "/v1/movies/foo",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			code, _, body := ts.get(t, tt.urlPath)

			assert.Equal(t, code, tt.wantCode)

			if tt.wantBody != "" {
				assert.StringContains(t, body, tt.wantBody)
			}

		})
	}

}

func TestCreateMovie(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routesTest())
	defer ts.Close()

	const (
		validTitle   = "Test Title"
		validYear    = 2021
		validRuntime = "105 mins"
	)

	validGenres := []string{"comedy", "drama"}

	tests := []struct {
		name     string
		Title    string
		Year     int32
		Runtime  string
		Genres   []string
		wantCode int
	}{
		{
			name:     "Valid submission",
			Title:    validTitle,
			Year:     validYear,
			Runtime:  validRuntime,
			Genres:   validGenres,
			wantCode: http.StatusCreated,
		},
		{
			name:     "Empty Title",
			Title:    "",
			Year:     validYear,
			Runtime:  validRuntime,
			Genres:   validGenres,
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name:     "year < 1888",
			Title:    validTitle,
			Year:     1500,
			Runtime:  validRuntime,
			Genres:   validGenres,
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name:     "test for wrong input",
			Title:    validTitle,
			Year:     validYear,
			Runtime:  validRuntime,
			Genres:   validGenres,
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputData := struct {
				Title   string   `json:"title"`
				Year    int32    `json:"year"`
				Runtime string   `json:"runtime"`
				Genres  []string `json:"genres"`
			}{
				Title:   tt.Title,
				Year:    tt.Year,
				Runtime: tt.Runtime,
				Genres:  tt.Genres,
			}

			b, err := json.Marshal(&inputData)
			if err != nil {
				t.Fatal("wrong input data")
			}
			if tt.name == "test for wrong input" {
				b = append(b, 'a')
			}

			code, _, _ := ts.postForm(t, "/v1/movies", b)

			assert.Equal(t, code, tt.wantCode)

		})
	}
}

func TestDeleteMovie(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routesTest())
	defer ts.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody string
	}{
		{
			name:     "deleting existing movie",
			urlPath:  "/v1/movies/1",
			wantCode: http.StatusOK,
		},
		{
			name:     "Non-existent ID",
			urlPath:  "/v1/movies/2",
			wantCode: http.StatusNotFound,
		}, {
			name:     "invalid ID",
			urlPath:  "/v1/movies/txt",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			code, _, body := ts.deleteReq(t, tt.urlPath)

			assert.Equal(t, code, tt.wantCode)

			if tt.wantBody != "" {
				assert.StringContains(t, body, tt.wantBody)
			}

		})
	}

}

func TestUpdateMovie(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routesTest())
	defer ts.Close()
	const (
		validTitle   = "Test Title"
		validYear    = 2021
		validRuntime = "105 mins"
	)
	validGenres := []string{"comedy", "drama"}

	tests := []struct {
		name    string
		Title   string
		Year    int32
		Runtime string
		Genres  []string
		urlPath string
		WCode   int
	}{
		{
			name:    "Checking existing movie",
			Title:   validTitle,
			Year:    validYear,
			Runtime: validRuntime,
			Genres:  validGenres,
			urlPath: "/v1/movies/1",
			WCode:   http.StatusOK,
		},
		{
			urlPath: "/v1/movies/150",
			name:    "Checking not existing movie",
			WCode:   http.StatusNotFound,
		},
		{
			urlPath: "/v1/movies/text",
			name:    "Invalid ID",
			WCode:   http.StatusNotFound,
		},
		{
			name:    "Wrong input",
			Title:   validTitle,
			urlPath: "/v1/movies/1",
			WCode:   http.StatusBadRequest,
		},
		{
			name:    "Failed validation",
			Title:   validTitle,
			Year:    1337,
			Runtime: validRuntime,
			Genres:  validGenres,
			urlPath: "/v1/movies/1",
			WCode:   http.StatusUnprocessableEntity,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputData := struct {
				Title   string   `json:"title"`
				Year    int32    `json:"year"`
				Runtime string   `json:"runtime"`
				Genres  []string `json:"genres"`
			}{
				Title:   tt.Title,
				Year:    tt.Year,
				Runtime: tt.Runtime,
				Genres:  tt.Genres,
			}

			d, err := json.Marshal(&inputData)
			if err != nil {
				t.Fatal("invalid data")
			}
			code, _, _ := ts.patchForm(t, tt.urlPath, d, http.MethodPatch)
			assert.Equal(t, code, tt.WCode)

		})
	}
}

func TestListMovies(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routesTest())
	defer ts.Close()

	tests := []struct {
		name    string
		urlPath string
		WCode   int
		WBody   string
	}{
		{
			name:    "invalid page size number",
			urlPath: "/v1/movies?page_size=-1",
			WCode:   http.StatusUnprocessableEntity,
		}, {
			name:    "invalid page size input",
			urlPath: "/v1/movies?page_size=txt",
			WCode:   http.StatusUnprocessableEntity,
		},
		{
			name:    "valid test",
			urlPath: "/v1/movies",
			WCode:   http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)

			assert.Equal(t, code, tt.WCode)

			if tt.WBody != "" {
				assert.StringContains(t, body, tt.WBody)
			}

		})
	}

}
