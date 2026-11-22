package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	requests := []struct {
		count int
		want  int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{100, 0},
	}
	handler := http.HandlerFunc(mainHandle)
	for city := range cafeList {
		for _, v := range requests {
			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/cafe?city=%s&count=%d", city, v.count), nil)
			handler.ServeHTTP(response, req)
			body := strings.TrimSpace(response.Body.String())
			var actualCountCafes int
			if body != "" {
				actualCountCafes = len(strings.Split(body, ","))
			}
			want := min(v.count, len(cafeList[city]))
			require.Equal(t, http.StatusOK, response.Code)
			assert.Equal(t, want, actualCountCafes)
		}
	}
}

func TestCafeSearch(t *testing.T) {
	requests := []struct {
		search    string
		wantCount int
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}

	handler := http.HandlerFunc(mainHandle)
	for _, v := range requests {
		var actualCountCafes int
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/cafe?city=moscow&search="+url.QueryEscape(v.search), nil)
		handler.ServeHTTP(response, req)

		body := strings.TrimSpace(response.Body.String())
		if body != "" {
			cafes := strings.Split(body, ",")
			actualCountCafes = len(cafes)
			for _, cafe := range cafes {
				contains := strings.Contains(strings.ToLower(cafe), strings.ToLower(v.search))
				assert.True(t, contains)
			}
		}
		require.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, v.wantCount, actualCountCafes)
	}
}
