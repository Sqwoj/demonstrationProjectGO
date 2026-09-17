package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func getSecretKey() []byte {
	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		log.Fatal("JWT_SECRET_KEY is not set!")
	}
	return []byte(secret)
}

func main() {
	err := initDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{"message": "it's a start"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	http.HandleFunc("/devices", authMiddleware(deviceHandler))
	http.HandleFunc("/events", authMiddleware(eventHandler))
	http.HandleFunc("/devices/identify", deviceIdentifyHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			markDevicesOffline()
		}
	}()

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func deviceExists(deviceID int) bool {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM devices WHERE id = $1)", deviceID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking device existence: %v", err)
		return false
	}
	return exists
}

func userExists(username string, password string) bool {
	var hashedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username = $1", username).Scan(&hashedPassword)
	if err != nil {
		log.Printf("Error checking user existence: %v", err)
		return false
	}
	if bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil {
		return true
	}
	return false
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return getSecretKey(), nil
		})
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func generateJWT(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(getSecretKey())
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func findUserByUsername(username string) *User {
	var user User
	err := db.QueryRow("SELECT id, username, password FROM users WHERE username = $1", username).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		return nil
	}
	return &user
}

func markDevicesOffline() {
	rows, err := db.Query(`
		SELECT id
		FROM devices
		WHERE status <> 'offline'
		  AND last_seen < NOW() - INTERVAL '2 minutes'
	`)
	if err != nil {
		log.Println("markDevicesOffline select error:", err)
		return
	}
	defer rows.Close()

	var offlineIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			log.Println("markDevicesOffline scan error:", err)
			continue
		}
		offlineIDs = append(offlineIDs, id)
	}
	if err := rows.Err(); err != nil {
		log.Println("markDevicesOffline rows error:", err)
	}

	for _, id := range offlineIDs {
		_, err := db.Exec(`
			UPDATE devices
			SET status = 'offline'
			WHERE id = $1
		`, id)
		if err != nil {
			log.Println("markDevicesOffline update error:", err)
			continue
		}

		_, err = db.Exec(`
			INSERT INTO events (device_id, type, severity, description, created_at)
			VALUES ($1, 'device_offline', 'warning', 'Device missed heartbeat', NOW())
		`, id)
		if err != nil {
			log.Println("markDevicesOffline insert event error:", err)
		}
	}
}
