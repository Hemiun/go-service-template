package dto

// Ping - dto for displaying service resources state
type Ping struct {
	Status      string `json:"status"`
	StatusDB    string `json:"statusDB"` //nolint:tagliatelle
	StatusKafka string `json:"statusKafka"`
	Message     string `json:"message"`
}
