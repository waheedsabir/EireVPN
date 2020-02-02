package test

import (
	"bytes"
	"eirevpn/api/config"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"time"
)

func TestGetPlanRoute(t *testing.T) {
	conf := config.GetConfig()
	makeRequest := func(t *testing.T, authToken, csrfToken string, planId uint) int {
		t.Helper()
		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/protected/plans/%d", planId)
		req, _ := http.NewRequest("GET", url, nil)
		authCookie := http.Cookie{Name: conf.App.AuthCookieName, Value: authToken, Expires: time.Now().Add(time.Minute * 5)}
		req.Header.Set("X-CSRF-Token", csrfToken)
		req.AddCookie(&authCookie)
		r.ServeHTTP(w, req)
		return w.Code
	}

	t.Run("Retrieve plan by ID", func(t *testing.T) {
		user := CreateAdminUser()
		authToken, csrfToken := GetToken(user)
		plan := CreatePlan()
		want := 200
		got := makeRequest(t, authToken, csrfToken, plan.ID)
		assertCorrectStatus(t, want, got)
		CreateCleanDB()
	})

	t.Run("Plan not found", func(t *testing.T) {
		user := CreateAdminUser()
		authToken, csrfToken := GetToken(user)
		want := 400
		planID := uint(999)
		got := makeRequest(t, authToken, csrfToken, planID)
		assertCorrectStatus(t, want, got)
		CreateCleanDB()
	})
}

func TestCreatePlanRoute(t *testing.T) {
	conf := config.GetConfig()
	makeRequest := func(t *testing.T, authToken, csrfToken string, plan map[string]interface{}) int {
		t.Helper()
		w := httptest.NewRecorder()
		j, _ := json.Marshal(plan)
		req, _ := http.NewRequest("POST", "/api/protected/plans/create", bytes.NewBuffer(j))
		authCookie := http.Cookie{Name: conf.App.AuthCookieName, Value: authToken, Expires: time.Now().Add(time.Minute * 5)}
		req.Header.Set("X-CSRF-Token", csrfToken)
		req.AddCookie(&authCookie)
		r.ServeHTTP(w, req)
		return w.Code
	}

	t.Run("Successful Plan Creation", func(t *testing.T) {
		plan := map[string]interface{}{
			"name":           "Test Product in test mode",
			"amount":         500,
			"interval":       "month",
			"interval_count": 1,
			"currency":       "EUR",
			"plan_type":      "PAYG",
		}
		user := CreateAdminUser()
		authToken, csrfToken := GetToken(user)
		want := 200
		got := makeRequest(t, authToken, csrfToken, plan)
		assertCorrectStatus(t, want, got)
		CreateCleanDB()
	})

	t.Run("Invalid form", func(t *testing.T) {
		halfFilledPlan := map[string]interface{}{
			"name":           "",
			"amount":         "",
			"interval":       "month",
			"interval_count": 1,
			"currency":       "EUR",
		}
		user := CreateAdminUser()
		authToken, csrfToken := GetToken(user)
		want := 400
		got := makeRequest(t, authToken, csrfToken, halfFilledPlan)
		assertCorrectStatus(t, want, got)
		CreateCleanDB()
	})

	t.Run("Drop table - Internal Server Error", func(t *testing.T) {
		plan := map[string]interface{}{
			"name":           "Test Product in test mode",
			"amount":         500,
			"interval":       "month",
			"interval_count": 1,
			"currency":       "EUR",
			"plan_type":      "PAYG",
		}
		user := CreateAdminUser()
		authToken, csrfToken := GetToken(user)
		want := 500
		DropPlanTable()
		got := makeRequest(t, authToken, csrfToken, plan)
		assertCorrectStatus(t, want, got)
		CreateCleanDB()
	})
}

func TestAllPlansRoute(t *testing.T) {
	conf := config.GetConfig()
	makeRequest := func(t *testing.T, authToken, csrfToken string) int {
		t.Helper()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/plans", nil)
		authCookie := http.Cookie{Name: conf.App.AuthCookieName, Value: authToken, Expires: time.Now().Add(time.Minute * 5)}
		req.Header.Set("X-CSRF-Token", csrfToken)
		req.AddCookie(&authCookie)
		r.ServeHTTP(w, req)
		return w.Code
	}

	t.Run("Successful get all plans", func(t *testing.T) {
		_ = CreatePlan()
		user := CreateUser()
		authToken, csrfToken := GetToken(user)
		want := 200
		got := makeRequest(t, authToken, csrfToken)
		assertCorrectStatus(t, want, got)
		CreateCleanDB()
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		user := CreateUser()
		authToken, csrfToken := GetToken(user)
		want := 500
		DropPlanTable()
		got := makeRequest(t, authToken, csrfToken)
		assertCorrectStatus(t, want, got)
		CreateCleanDB()
	})
}

func TestUpdatePlanRoute(t *testing.T) {
	conf := config.GetConfig()
	makeRequest := func(t *testing.T, authToken, csrfToken string, plan map[string]interface{}, id uint) int {
		t.Helper()
		w := httptest.NewRecorder()
		j, _ := json.Marshal(plan)
		url := fmt.Sprintf("/api/protected/plans/update/%d", id)
		req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(j))
		authCookie := http.Cookie{Name: conf.App.AuthCookieName, Value: authToken, Expires: time.Now().Add(time.Minute * 5)}
		req.Header.Set("X-CSRF-Token", csrfToken)
		req.AddCookie(&authCookie)
		r.ServeHTTP(w, req)
		return w.Code
	}

	t.Run("Successful Update Plan", func(t *testing.T) {
		plan := map[string]interface{}{
			"name": "Update test plan",
		}
		p := CreatePlan()
		user := CreateAdminUser()
		authToken, csrfToken := GetToken(user)
		want := 200
		got := makeRequest(t, authToken, csrfToken, plan, p.ID)
		assertCorrectStatus(t, want, got)
		CreateCleanDB()
	})

	t.Run("Invalid form", func(t *testing.T) {
		halfFilledPlan := map[string]interface{}{
			"name": "",
		}
		p := CreatePlan()
		user := CreateAdminUser()
		authToken, csrfToken := GetToken(user)
		want := 400
		got := makeRequest(t, authToken, csrfToken, halfFilledPlan, p.ID)
		assertCorrectStatus(t, want, got)
		CreateCleanDB()
	})
}

func TestDeletePlanRoute(t *testing.T) {
	conf := config.GetConfig()
	makeRequest := func(t *testing.T, authToken, csrfToken string, id uint) int {
		t.Helper()
		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/protected/plans/delete/%d", id)
		req, _ := http.NewRequest("DELETE", url, nil)
		authCookie := http.Cookie{Name: conf.App.AuthCookieName, Value: authToken, Expires: time.Now().Add(time.Minute * 5)}
		req.Header.Set("X-CSRF-Token", csrfToken)
		req.AddCookie(&authCookie)
		r.ServeHTTP(w, req)
		return w.Code
	}

	t.Run("Successful Delete Plan", func(t *testing.T) {
		plan := CreatePlan()
		user := CreateAdminUser()
		authToken, csrfToken := GetToken(user)
		want := 200
		got := makeRequest(t, authToken, csrfToken, plan.ID)
		assertCorrectStatus(t, want, got)
		CreateCleanDB()
	})

	t.Run("Plan not found", func(t *testing.T) {
		user := CreateAdminUser()
		authToken, csrfToken := GetToken(user)
		want := 400
		planID := uint(999)
		got := makeRequest(t, authToken, csrfToken, planID)
		assertCorrectStatus(t, want, got)
		CreateCleanDB()
	})
}
