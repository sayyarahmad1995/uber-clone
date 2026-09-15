package marketplace

import "github.com/google/uuid"

type OperationFare struct {
	AmountMinor int64  `json:"amount_minor"`
	Currency    string `json:"currency"`
}

type OperationContext struct {
	VehicleID    uuid.UUID     `json:"vehicle_id"`
	ServiceCode  string        `json:"service_code"`
	ServiceName  string        `json:"service_name"`
	DriverName   string        `json:"driver_name"`
	Make         string        `json:"make"`
	Model        string        `json:"model"`
	ModelYear    int           `json:"model_year"`
	Color        string        `json:"color"`
	LicensePlate string        `json:"license_plate"`
	Fare         OperationFare `json:"fare"`
}
