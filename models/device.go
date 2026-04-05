package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Type represents the kind of device feature (controller or sensor).
type Type string

// Supported feature types.
const (
	Controller Type = "controller"
	Sensor     Type = "sensor"
)

// Feature describes a single capability of a device.
type Feature struct {
	UUID   string `json:"uuid" bson:"uuid"`
	Type   Type   `json:"type" bson:"type"`
	Name   string `json:"name" bson:"name"`
	Enable bool   `json:"enable" bson:"enable"`
	Order  int    `json:"order" bson:"order"`
	Unit   string `json:"unit" bson:"unit"`
}

// Device represents a registered IoT device with its features and metadata.
type Device struct {
	//swagger:ignore
	ID           bson.ObjectID `json:"id" bson:"_id"`
	UUID         string        `json:"uuid" bson:"uuid"`
	Mac          string        `json:"mac" bson:"mac"`
	Manufacturer string        `json:"manufacturer" bson:"manufacturer"`
	Model        string        `json:"model" bson:"model"`
	Features     []Feature     `json:"features" bson:"features"`
	CreatedAt    time.Time     `json:"createdAt" bson:"createdAt"`
	ModifiedAt   time.Time     `json:"modifiedAt" bson:"modifiedAt"`
}
