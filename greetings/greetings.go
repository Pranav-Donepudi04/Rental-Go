// Package greetings provides greeting functionality
package greetings

import "fmt"

// Hello returns a greeting message for the given name
func Hello(name string) string {
	message := fmt.Sprintf("Hi, %v. Welcome!", name)
	return message
}

// Goodbye returns a farewell message for the given name
func Goodbye(name string) string {
	message := fmt.Sprintf("Goodbye, %v. See you later!", name)
	return message
}
