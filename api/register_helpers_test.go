package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"admission/models"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestProfileOwnsDevice(t *testing.T) {
	deviceID := bson.NewObjectID()
	profile := models.Profile{
		Devices: []bson.ObjectID{bson.NewObjectID(), deviceID},
	}

	if !profileOwnsDevice(&profile, deviceID) {
		t.Fatal("profileOwnsDevice() = false, want true")
	}
}

func TestProfileOwnsDeviceReturnsFalseWhenMissing(t *testing.T) {
	profile := models.Profile{
		Devices: []bson.ObjectID{bson.NewObjectID()},
	}

	if profileOwnsDevice(&profile, bson.NewObjectID()) {
		t.Fatal("profileOwnsDevice() = true, want false")
	}
}

func TestRegisterSensorsViaHTTPReturnsErrorWhenKeepAliveFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	handler := Register{
		keepAliveSensorURL: server.URL,
		registerSensorURL:  server.URL + "/register/",
	}

	err := handler.registerSensorsViaHTTP(testDevice(), []models.Feature{testSensorFeature()}, testProfile(), "token")

	if err == nil {
		t.Fatal("registerSensorsViaHTTP() error = nil, want error")
	}
}

func TestRegisterSensorsViaHTTPReturnsErrorWhenRegisterFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/keepalive" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	handler := Register{
		keepAliveSensorURL: server.URL + "/keepalive",
		registerSensorURL:  server.URL + "/register/",
	}

	err := handler.registerSensorsViaHTTP(testDevice(), []models.Feature{testSensorFeature()}, testProfile(), "token")

	if err == nil {
		t.Fatal("registerSensorsViaHTTP() error = nil, want error")
	}
}

func TestRegisterSensorsViaHTTPRegistersEverySensor(t *testing.T) {
	var registered []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/keepalive" {
			w.WriteHeader(http.StatusOK)
			return
		}
		registered = append(registered, r.URL.Path)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	handler := Register{
		keepAliveSensorURL: server.URL + "/keepalive",
		registerSensorURL:  server.URL + "/register/",
	}
	sensors := []models.Feature{
		testSensorFeature(),
		{UUID: "feature-2", Name: "humidity"},
		{UUID: "feature-3", Name: "mode"},
	}

	err := handler.registerSensorsViaHTTP(testDevice(), sensors, testProfile(), "token")

	if err != nil {
		t.Fatalf("registerSensorsViaHTTP() error = %v", err)
	}
	if len(registered) != 3 {
		t.Fatalf("registered calls = %d, want 3", len(registered))
	}
	if registered[0] != "/register/temperature" || registered[1] != "/register/humidity" || registered[2] != "/register/mode" {
		t.Fatalf("registered paths = %#v, want temperature, humidity, and mode", registered)
	}
}

func testDevice() *models.Device {
	return &models.Device{
		UUID:         "device-uuid",
		Mac:          "11:22:33:44:55:66",
		Manufacturer: "test",
		Model:        "test-model",
	}
}

func testProfile() *models.Profile {
	return &models.Profile{ID: bson.NewObjectID()}
}

func testSensorFeature() models.Feature {
	return models.Feature{
		UUID: "feature-1",
		Type: models.Sensor,
		Name: "temperature",
	}
}
