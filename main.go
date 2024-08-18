package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
)

// keyServerAddr is used as a key for storing and retrieving the server address from the context.
const keyServerAddr = "serverAddr"

// HelloRequest defines the structure of the JSON payload expected by the /hello endpoint.
type HelloRequest struct {
	Name  string `json:"Name"`  // Name field from the JSON payload
	Age   int    `json:"Age"`   // Age field from the JSON payload
	Hobby string `json:"Hobby"` // Hobby field from the JSON payload
}

// getRoot handles requests to the root ("/") endpoint.
func getRoot(w http.ResponseWriter, r *http.Request) {
	// Retrieve the context associated with the request.
	ctx := r.Context()

	// Extract query parameters from the URL.
	hasFirst := r.URL.Query().Has("first")
	first := r.URL.Query().Get("first")
	hasSecond := r.URL.Query().Has("second")
	second := r.URL.Query().Get("second")

	// Read the body of the request.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		// If reading the body fails, log the error.
		fmt.Printf("could not read body: %s\n", err)
	}

	// Log details about the request.
	fmt.Printf("%s: got / request. first(%t)=%s, second(%t)=%s, body:\n%s\n",
		ctx.Value(keyServerAddr), // Print server address
		hasFirst, first,          // Query parameter 'first'
		hasSecond, second, // Query parameter 'second'
		body) // Request body

	// Respond with a simple message.
	io.WriteString(w, "This is my website!\n")
}

// getHello handles POST requests to the "/hello" endpoint.
func getHello(w http.ResponseWriter, r *http.Request) {
	// Retrieve the context associated with the request.
	ctx := r.Context()

	// Define a variable to store the parsed JSON data.
	var req HelloRequest

	// Decode the JSON body into the HelloRequest struct.
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		// If JSON decoding fails, respond with a bad request error.
		http.Error(w, "Bad Request: invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate that Age is a positive number.
	if req.Age <= 0 {
		http.Error(w, "Bad Request: Age must be a positive number", http.StatusBadRequest)
		return
	}

	// Log details about the request.
	fmt.Printf("%s: got /hello request\n", ctx.Value(keyServerAddr))
	fmt.Printf("Name: %s, Age: %d, Hobby: %s\n", req.Name, req.Age, req.Hobby)

	// Respond with a personalized greeting message.
	io.WriteString(w, fmt.Sprintf("Hello, %s! You are %d years old and enjoy %s.\n", req.Name, req.Age, req.Hobby))
}

// main initializes and starts the HTTP server.
func main() {
	// Create a new ServeMux (HTTP request multiplexer).
	mux := http.NewServeMux()

	// Register the handlers for the endpoints.
	mux.HandleFunc("/", getRoot)
	mux.HandleFunc("/hello", getHello)

	// Initialize a background context.
	ctx := context.Background()

	// Create a new HTTP server instance.
	server := &http.Server{
		Addr:    ":3333", // Server listens on port 3333
		Handler: mux,
		// BaseContext sets up the context with server address for logging.
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
			return ctx
		},
	}

	// Start the server and handle potential errors.
	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		// If the server was closed gracefully.
		fmt.Printf("server closed\n")
	} else if err != nil {
		// If there was an error starting the server.
		fmt.Printf("error listening for server: %s\n", err)
	}
}
