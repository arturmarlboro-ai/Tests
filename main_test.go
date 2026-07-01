package main

import (
	"net/http"
	"net/http/httptest"
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
	counts := []string{
		"count=0",
		"count=1",
		"count=2",
		"count=100",
	}
	cities := []string{"moscow", "tula"}
	for _, v := range counts {
		for _, c := range cities {
			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/cafe?"+v+"&city="+c, nil)

			handler.ServeHTTP(response, req)
			countStr := req.FormValue("count")
			count, err := strconv.Atoi(countStr)
			if err != nil {
				t.Errorf("Failed to convert count to integer: %s", countStr)
			}
			city := req.FormValue("city")
			cafe, ok := cafeList[city]
			if !ok {
				t.Errorf("Unknown city: %s", city)
			}
			count = min(count, len(cafe))
			answer := strings.Join(cafe[:count], ",")
			//fmt.Printf("Request: %s, City: %s, Count: %d, Expected Answer: %s\n", v, city, count, answer)
			answerSlice := strings.Split(answer, ",")
			if len(answerSlice) == 1 && answerSlice[0] == "" {
				answerSlice = []string{}
			}
			lenAnswer := len(answerSlice)

			require.Equal(t, http.StatusOK, response.Code)
			assert.Equal(t, count, lenAnswer)
		}
	}
}
func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	c := "city=moscow"
	look4 := []struct {
		search    string
		wantCount int
	}{
		{"search=фасоль", 0},
		{"search=кофе", 2},
		{"search=вилка", 1},
	}

	for _, r := range look4 {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/cafe?"+c+"&"+r.search, nil)
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
