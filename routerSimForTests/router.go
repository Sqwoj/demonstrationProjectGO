package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Device struct {
	Name         string `json:"name"`
	IP           string `json:"ip"`
	Model        string `json:"model,omitempty"`
	Location     string `json:"location,omitempty"`
	SerialNumber string `json:"serial_number,omitempty"`
	LastSeen     string `json:"last_seen,omitempty"`
	Uptime       string `json:"uptime,omitempty"`
	Description  string `json:"description,omitempty"`
}

func main() {
	device := Device{
		Name:         "Router-Sim",
		IP:           "192.168.1.50",
		Model:        "Router-Model-01",
		Location:     "Russia",
		SerialNumber: "SIM-ROUTER-001",
		Description:  "Demo router simulator",
	}

	endpoint := "http://localhost:8080/devices/identify"

	for true {
		body, err := json.Marshal(device)
		if err != nil {
			panic(err)
		}

		resp, err := http.Post(endpoint, "application/json", bytes.NewReader(body))
		if err != nil {
			fmt.Println("Ошибка запроса:", err)
			time.Sleep(5 * time.Second)
			continue
		}

		fmt.Println("Статус:", resp.Status)
		resp.Body.Close()

		time.Sleep(10 * time.Second)
	}

	fmt.Println("Симулятор завершил работу")
}
