package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

func deviceHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		deviceHandlerPost(w, r)
	case http.MethodGet:
		deviceHandlerGet(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
func deviceHandlerPost(w http.ResponseWriter, r *http.Request) {
	var newDevice Device
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewDecoder(r.Body).Decode(&newDevice); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if newDevice.Name == "" || newDevice.IP == "" {
		http.Error(w, "IP or Name is missing", http.StatusBadRequest)
		return
	}
	if newDevice.Status == "" {
		newDevice.Status = "inactive"
	}
	query := `
		INSERT INTO devices (
			name, ip, status, model, location, serial_number, last_seen, uptime, description
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	if err := db.QueryRow(
		query,
		newDevice.Name,
		newDevice.IP,
		newDevice.Status,
		newDevice.Model,
		newDevice.Location,
		newDevice.SerialNumber,
		newDevice.LastSeen,
		newDevice.Uptime,
		newDevice.Description,
	).Scan(&newDevice.ID); err != nil {
		http.Error(w, "Failed to add device", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newDevice)
	fmt.Printf("Device added: \n ID=%d,\n Name=%s,\n IP=%s\n", newDevice.ID, newDevice.Name, newDevice.IP)

}
func deviceHandlerGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rows, err := db.Query(`
		SELECT id, name, ip, status,
		       COALESCE(model, ''),
		       COALESCE(location, ''),
		       COALESCE(serial_number, ''),
		       COALESCE(last_seen::text, ''),
		       COALESCE(uptime, ''),
		       COALESCE(description, '')
		FROM devices
	`)
	if err != nil {
		http.Error(w, "Failed to fetch devices", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var devices []Device
	for rows.Next() {
		var d Device
		if err := rows.Scan(
			&d.ID,
			&d.Name,
			&d.IP,
			&d.Status,
			&d.Model,
			&d.Location,
			&d.SerialNumber,
			&d.LastSeen,
			&d.Uptime,
			&d.Description,
		); err != nil {
			http.Error(w, "Failed to scan devices", http.StatusInternalServerError)
			return
		}
		devices = append(devices, d)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Failed to read devices", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(devices)
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		eventHandlerPost(w, r)
	case http.MethodGet:
		eventHandlerGet(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
func eventHandlerPost(w http.ResponseWriter, r *http.Request) {
	var newEvent Event
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewDecoder(r.Body).Decode(&newEvent); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if !deviceExists(newEvent.DeviceID) {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}
	if newEvent.Description == "" {
		http.Error(w, "Description is missing", http.StatusBadRequest)
		return
	}
	if newEvent.Type == "" {
		newEvent.Type = "info"
	}
	if newEvent.Severity == "" {
		newEvent.Severity = "info"
	}
	query := `
		INSERT INTO events (device_id, type, severity, description, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id
	`
	if err := db.QueryRow(query, newEvent.DeviceID, newEvent.Type, newEvent.Severity, newEvent.Description).Scan(&newEvent.ID); err != nil {
		http.Error(w, "Failed to add event", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newEvent)
}
func eventHandlerGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	deviceIDText := r.URL.Query().Get("device_id")
	if deviceIDText == "" {
		http.Error(w, "device_id is required", http.StatusBadRequest)
		return
	}
	deviceID, err := strconv.Atoi(deviceIDText)
	if err != nil {
		http.Error(w, "Invalid device_id", http.StatusBadRequest)
		return
	}

	rows, err := db.Query(`
		SELECT id, device_id, type, severity, description, created_at
		FROM events
		WHERE device_id = $1
	`, deviceID)
	if err != nil {
		http.Error(w, "Failed to retrieve events", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var deviceEvent []Event
	for rows.Next() {
		var event Event
		if err := rows.Scan(&event.ID, &event.DeviceID, &event.Type, &event.Severity, &event.Description, &event.CreatedAt); err != nil {
			http.Error(w, "Failed to scan events", http.StatusInternalServerError)
			return
		}
		deviceEvent = append(deviceEvent, event)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Failed to read events", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(deviceEvent)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		registerHandlerPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func registerHandlerPost(w http.ResponseWriter, r *http.Request) {
	var newUser User
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if newUser.Username == "" || newUser.Password == "" {
		http.Error(w, "Username or Password is missing", http.StatusBadRequest)
		return
	}
	if findUserByUsername(newUser.Username) != nil {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	newUser.Password = string(hashedPassword)
	query := "INSERT INTO users (username, password) VALUES ($1, $2)"
	_, err = db.Exec(query, newUser.Username, newUser.Password)
	if err != nil {
		http.Error(w, "Failed to add user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	response := map[string]string{
		"message": "User registered successfully",
		"status":  "success",
	}
	json.NewEncoder(w).Encode(response)
	log.Printf("User registered: \n ID=%d,\n Username=%s\n", newUser.ID, newUser.Username)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		loginHandlerPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func loginHandlerPost(w http.ResponseWriter, r *http.Request) {
	var loginUser User
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewDecoder(r.Body).Decode(&loginUser); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		log.Printf("Failed login attempt: \n Username=%s\n", loginUser.Username)
		return
	}
	if loginUser.Username == "" || loginUser.Password == "" {
		http.Error(w, "Username or Password is missing", http.StatusBadRequest)
		log.Printf("Failed login attempt: \n Username=%s\n", loginUser.Username)
		return
	}
	if !userExists(loginUser.Username, loginUser.Password) {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		log.Printf("Failed login attempt: \n Username=%s\n", loginUser.Username)
		return
	}
	token, err := generateJWT(loginUser.Username)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		log.Printf("Failed to generate token for user: \n Username=%s\n", loginUser.Username)
		return
	}
	response := map[string]string{
		"message":  "Login successful",
		"status":   "success",
		"token":    token,
		"username": loginUser.Username}
	log.Printf("User logged in: \n Username=%s\n", loginUser.Username)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func deviceIdentifyHandler(w http.ResponseWriter, r *http.Request) {
	var nameDevice Device
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewDecoder(r.Body).Decode(&nameDevice); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if nameDevice.Name == "" || nameDevice.SerialNumber == "" {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if nameDevice.Status == "" {
		nameDevice.Status = "online"
	}

	err := db.QueryRow("SELECT name, serial_number FROM devices WHERE serial_number = $1", nameDevice.SerialNumber).Scan(&nameDevice.Name, &nameDevice.SerialNumber)
	switch err {
	case nil:
		query := `
			UPDATE devices
			SET last_seen = NOW(),
			    status = 'online'
			WHERE serial_number = $1
		`
		_, updErr := db.Exec(query,
			nameDevice.SerialNumber,
		)
		if updErr != nil {
			http.Error(w, "Failed to update device", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "online", "serial_number": nameDevice.SerialNumber})
		return
	case sql.ErrNoRows:
		query := `
			INSERT INTO devices (name, ip, status, model, location, serial_number, last_seen, uptime, description)
			VALUES ($1, $2, 'online', $3, $4, $5, NOW(), $6, $7)
		`
		_, insErr := db.Exec(query,
			nameDevice.Name,
			nameDevice.IP,
			nameDevice.Model,
			nameDevice.Location,
			nameDevice.SerialNumber,
			nameDevice.Uptime,
			nameDevice.Description,
		)
		if insErr != nil {
			http.Error(w, "Failed to create device", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "online", "serial_number": nameDevice.SerialNumber})
		return
	default:
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
}
