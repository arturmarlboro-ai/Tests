package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}
func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		count  int // передаваемое значение count
		want   int // ожидаемое количество кафе в ответе
		cities []string
	}{
		{count: 0, want: 0, cities: []string{"moscow", "tula"}},
		{count: 1, want: 1, cities: []string{"moscow", "tula"}},
		{count: 2, want: 2, cities: []string{"moscow", "tula"}},
		{count: 100, want: min(len(cafeList["moscow"]), 100), cities: []string{"moscow"}}, // в Москве 5 кафе
	}

	for _, v := range requests {
		for _, c := range v.cities {
			params := url.Values{}
			params.Set("count", strconv.Itoa(v.count))
			params.Set("city", c)

			req := httptest.NewRequest("GET", "/cafe?"+params.Encode(), nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)

			body := response.Body.String()
			var count int
			if body != "" {
				count = len(strings.Split(body, ","))
			}
			assert.Equal(t, v.want, count)
		}
	}
}
func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	requests := []struct {
		search    string // передаваемое значение search
		wantCount int    // ожидаемое количество кафе в ответе
	}{
		{search: "фасоль", wantCount: 0},
		{search: "кофе", wantCount: 2},
		{search: "вилка", wantCount: 1},
	}
	c := "moscow"

	for _, r := range requests {
		params := url.Values{}
		params.Set("city", c)
		params.Set("search", r.search)

		req := httptest.NewRequest("GET", "/cafe?"+params.Encode(), nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code)

		body := response.Body.String()
		var count int
		if body != "" {
			count = len(strings.Split(body, ","))
		}

		assert.Equal(t, r.wantCount, count)
	}

}
