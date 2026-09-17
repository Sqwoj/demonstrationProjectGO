package main

type Device struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	IP           string `json:"ip"`
	Status       string `json:"status"`
	Model        string `json:"model,omitempty"`
	Location     string `json:"location,omitempty"`
	SerialNumber string `json:"serial_number,omitempty"`
	LastSeen     string `json:"last_seen,omitempty"`
	Uptime       string `json:"uptime,omitempty"`
	Description  string `json:"description,omitempty"`
}

type deviceResponse struct {
	Message string `json:"message"`
}

type Event struct {
	ID          int    `json:"id"`
	DeviceID    int    `json:"device_id"`
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at,omitempty"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}
