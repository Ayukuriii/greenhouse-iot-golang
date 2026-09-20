package dto

type DeviceControlRequest struct {
	DeviceID string `json:"device_id" validate:"required"`
	Command  string `json:"command" validate:"required,oneof=ON OFF"`
}

type MQTTDevicePayload struct {
	DeviceID  string `json:"device_id"`
	Command   string `json:"command"`
	Timestamp string `json:"timestamp"`
}
