package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db             *gorm.DB
	usersLock      sync.Mutex
	smtpServer     = "smtp.gmail.com"
	smtpPort       = "587"
	senderEmail    = "kisuigone@gmail.com"
	senderPassword = "Qw!2#As"
)

type User struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	var err error
	dsn := "user=postgres password=1 dbname=Users port=5432 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Automatically migrate schema
	db.AutoMigrate(&User{})

	// Setup routes
	http.HandleFunc("/", homePage)
	http.HandleFunc("/users", handleUsers)
	http.HandleFunc("/users/", handleUserByID)
	http.HandleFunc("/delete/", handleDeleteUser)
	http.HandleFunc("/messages", handleMessage) // Новая обработка для сообщений

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Handler для обработки POST-запросов с сообщением
func handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var msg map[string]string
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&msg)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		message := msg["message"]
		fmt.Printf("Received message: %s\n", message)

		// Отправка успешного ответа
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]string{
			"status":  "success",
			"message": message,
		}
		json.NewEncoder(w).Encode(response)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

// Домашняя страница
func homePage(w http.ResponseWriter, r *http.Request) {
	html := `
	<!DOCTYPE html>
	<html>
	<head><title>User Management</title></head>
	<body>
		<h1>User Management</h1>
		<form action="/users" method="post">
			<label for="name">Name:</label><br>
			<input type="text" id="name" name="name"><br>
			<label for="email">Email:</label><br>
			<input type="email" id="email" name="email"><br><br>
			<input type="submit" value="Add User">
		</form>
		<h2>Users</h2>
		<table border="1">
			<tr><th>ID</th><th>Name</th><th>Email</th><th>Actions</th></tr>
	`
	var users []User
	db.Find(&users)
	for _, user := range users {
		html += fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%s</td><td><a href='/delete/%d'>Delete</a></td></tr>", user.ID, user.Name, user.Email, user.ID)
	}
	html += `
		</table>
	</body>
	</html>
	`
	w.Write([]byte(html))
}

// Handle /users route
func handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		email := r.FormValue("email")

		// Create user
		user := User{Name: name, Email: email}
		result := db.Create(&user)
		if result.Error != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		// Send email notification
		go sendEmailNotification(user)

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		var users []User
		db.Find(&users)
		usersJSON, _ := json.Marshal(users)
		w.Header().Set("Content-Type", "application/json")
		w.Write(usersJSON)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// Handle /users/{id} route
func handleUserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/users/")
	var user User
	result := db.First(&user, id)
	if result.Error != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	userJSON, _ := json.Marshal(user)
	w.Header().Set("Content-Type", "application/json")
	w.Write(userJSON)
}

// Handle /delete/{id} route
func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/delete/")
	var user User
	result := db.Delete(&user, id)
	if result.Error != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Send email notification
func sendEmailNotification(user User) {
	auth := smtp.PlainAuth("", senderEmail, senderPassword, smtpServer)
	to := []string{user.Email}
	subject := "Welcome to User Management"
	body := fmt.Sprintf("Hello %s,\n\nThank you for registering with our system.", user.Name)
	message := []byte("Subject: " + subject + "\r\n" + "To: " + user.Email + "\r\n" + "MIME-Version: 1.0\r\n" + "Content-Type: text/plain; charset=UTF-8\r\n\r\n" + body)

	err := smtp.SendMail(smtpServer+":"+smtpPort, auth, senderEmail, to, message)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
	}
}
