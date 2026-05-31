package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"admission/models"

	"github.com/gin-gonic/gin"
)

func TestGetKeepAliveReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/keepalive", GetKeepAlive)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/keepalive", nil)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	var body map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	if body["message"] != "ok" {
		t.Fatalf("message = %q, want %q", body["message"], "ok")
	}
}

func TestSpecReqToModelMapsListItems(t *testing.T) {
	min := 0.0
	max := 10.0
	step := 1.0
	first := 1
	second := 2

	spec := specReqToModel(SpecReq{
		Format: models.List,
		Min:    &min,
		Max:    &max,
		Step:   &step,
		List: []SpecListItemReq{
			{Value: &first, Text: "one"},
			{Value: &second, Text: "two"},
		},
	})

	if spec.Format != models.List {
		t.Fatalf("Format = %q, want %q", spec.Format, models.List)
	}
	if spec.Min != &min || spec.Max != &max || spec.Step != &step {
		t.Fatal("range pointers were not preserved")
	}
	if len(spec.List) != 2 {
		t.Fatalf("len(List) = %d, want 2", len(spec.List))
	}
	if spec.List[0].Value != 1 || spec.List[0].Text != "one" {
		t.Fatalf("first list item = %#v, want value 1 text one", spec.List[0])
	}
	if spec.List[1].Value != 2 || spec.List[1].Text != "two" {
		t.Fatalf("second list item = %#v, want value 2 text two", spec.List[1])
	}
}
