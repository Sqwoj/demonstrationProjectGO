package main

import (
	"database/sql"
)

var devices []Device
var events []Event
var users []User
var db *sql.DB
