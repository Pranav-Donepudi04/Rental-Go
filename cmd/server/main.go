package main

import (
	"backend-form/m/internal/config"
	"backend-form/m/internal/handlers"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/backend-form/greetings"
)

var templates = template.Must(template.ParseFiles(
	"templates/form.html",
	"templates/success.html",
))

func main() {
	fmt.Println("Starting Go Backend Server...")

	// Load configuration
	config.LoadEnvFile()
	cfg := config.Load()

	// Use our custom greetings function
	greeting := greetings.Hello("Surya pranav")
	fmt.Println(greeting)

	// Print configuration
	fmt.Printf("Server will start on port: %s\n", cfg.Port)
	fmt.Printf("Database URL: %s\n", cfg.DatabaseURL)
	fmt.Printf("Log Level: %s\n", cfg.LogLevel)

	// Set up HTTP routes
	formHandler := handlers.NewFormHandler(templates)
	http.HandleFunc("/", formHandler.ShowForm)
	http.HandleFunc("/submit", formHandler.SubmitForm)

	// Start the server using config
	fmt.Printf("Server starting on port %s...\n", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
