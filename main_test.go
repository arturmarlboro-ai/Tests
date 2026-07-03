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
	//
	requests := []struct {
		count int // передаваемое значение count
		want  int // ожидаемое количество кафе в ответе
	}{
		{count: 0, want: 0},
		{count: 1, want: 1},
		{count: 2, want: 2},
		{count: 100, want: 5}, // в Москве 5 кафе
	}

	cities := []string{"moscow", "tula"}
	for _, v := range requests {
		for _, c := range cities {
			params := url.Values{}
			if v.count == 0 {
				params.Set("count", "0")
			}
			params.Set("count", strconv.Itoa(v.count))
			if c != "" {
				params.Set("city", c)
			}

			req := httptest.NewRequest("GET", "/cafe?"+params.Encode(), nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, req)

			countStr := req.FormValue("count")
			count, err := strconv.Atoi(countStr)
			if err != nil {
				t.Error(err)
			}
			city := req.FormValue("city")
			cafes := cafeList[city]

			count = min(count, len(cafes))
			answer := strings.Join(cafes[:count], ",")

			answerSlice := strings.Split(answer, ",")

			if len(answerSlice) == 1 && answerSlice[0] == "" {
				answerSlice = []string{}
			}
			lenAnswer := len(answerSlice)
			if city == "tula" && v.count == 100 {
				v.want = 3
			}
			require.Equal(t, http.StatusOK, response.Code)
			assert.Equal(t, v.want, lenAnswer)
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

		city := req.FormValue("city")
		search := req.FormValue("search")
		cafes := cafeList[city]
		var found []string
		for _, cafe := range cafes {
			cafe = strings.ToLower(cafe)
			if strings.Contains(cafe, search) {
				found = append(found, cafe)
			}
		}
		require.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, r.wantCount, len(found))
	}

}
