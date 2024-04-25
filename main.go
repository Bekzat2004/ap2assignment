package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
)

// Struct to hold question-response pairs
type QA struct {
	User     string
	Question string
	Response string
}

// Global variable to hold all requests history
var allRequestsHistory []QA

func main() {
	http.HandleFunc("/", handleRequest)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		renderTemplate(w, "index.html", nil)
		return
	}

	apiKey := "sk-proj-jJW7tPHq7l2SAKbxUkiwT3BlbkFJA5Yf4IlTNeH5ZzXWcHg0"
	apiEndpoint := "https://api.openai.com/v1/chat/completions"

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	userInput := r.FormValue("userInput")

	// Check if the user input contains filter words
	if containsFilterWords(userInput) {
		// Proceed with GPT-3 interaction
		client := resty.New()

		customSettings := map[string]interface{}{
			"model":      "gpt-3.5-turbo",
			"messages":   []interface{}{map[string]interface{}{"role": "system", "content": userInput}},
			"max_tokens": 10, // Change max tokens here
		}

		response, err := client.R().
			SetAuthToken(apiKey).
			SetHeader("Content-Type", "application/json").
			SetBody(customSettings).
			Post(apiEndpoint)

		if err != nil {
			http.Error(w, fmt.Sprintf("Error while sending the request: %v", err), http.StatusInternalServerError)
			return
		}

		body := response.Body()

		var data map[string]interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error while decoding JSON response: %v", err), http.StatusInternalServerError)
			return
		}

		content := data["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)

		
		// Save request history
		user := r.RemoteAddr
		request := QA{
			User:     user,
			Question: userInput,
			Response: content,
		}
		allRequestsHistory = append(allRequestsHistory, request)

		renderTemplate(w, "index.html", map[string]interface{}{
			"userInput": userInput,
			"response":  content,
			"history":   allRequestsHistory,
		})
	} else {

		notification := "Your request was declined because your question is not related to the vision of the touristic company."
		renderTemplate(w, "index.html", map[string]interface{}{
			"notification": notification,
		})
	}
}

func containsFilterWords(input string) bool {
	// Define filter words
	filterWords := []string{"virtual assistant", "tourist", "travel", "destination", "sightseeing"}

	// Check if any filter word exists in the input
	for _, word := range filterWords {
		if strings.Contains(strings.ToLower(input), word) {
			return true
		}
	}
	return false
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing template: %v", err), http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
		return
	}
}
