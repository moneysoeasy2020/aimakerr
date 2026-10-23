package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// webhook is a HTTP Handler for Gitlab Webhook events.
type webhook struct {
	Secret         string
	EventsToAccept []gitlab.EventType
}

// webhookExample shows how to create a Webhook server to parse Gitlab events.
func webhookExample() {
	wh := webhook{
		Secret:         "your-gitlab-secret",
		EventsToAccept: []gitlab.EventType{gitlab.EventTypeMergeRequest, gitlab.EventTypePipeline},
	}

	mux := http.NewServeMux()
	mux.Handle("/webhook", wh)

	server := &http.Server{
		Addr:         "127.0.0.1:8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
}

// ServeHTTP tries to parse Gitlab events sent and calls handle function
// with the successfully parsed events.
func (hook webhook) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	event, err := hook.parse(request)
	if err != nil {
		writer.WriteHeader(500)
		fmt.Fprintf(writer, "could not parse the webhook event: %v", err)
		return
	}

	// Handle the event before we return.
	if err := hook.handle(event); err != nil {
		writer.WriteHeader(500)
		fmt.Fprintf(writer, "error handling the event: %v", err)
		return
	}

	// Write a response when were done.
	writer.WriteHeader(204)
}

func (hook webhook) handle(event any) error {
	str, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("could not marshal json event for logging: %v", err)
	}

	// Just write the event for this example.
	fmt.Println(string(str))

	return nil
}

// parse verifies and parses the events specified in the request and
// returns the parsed event or an error.
func (hook webhook) parse(r *http.Request) (any, error) {
	defer func() {
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			log.Printf("could discard request body: %v", err)
		}
		if err := r.Body.Close(); err != nil {
			log.Printf("could not close request body: %v", err)
		}
	}()

	if r.Method != http.MethodPost {
		return nil, errors.New("invalid HTTP Method")
	}

	// If we have a secret set, we should check if the request matches it.
	if len(hook.Secret) > 0 {
		signature := r.Header.Get("X-Gitlab-Token")
		if signature != hook.Secret {
			return nil, errors.New("token validation failed")
		}
	}

	event := r.Header.Get("X-Gitlab-Event")
	if strings.TrimSpace(event) == "" {
		return nil, errors.New("missing X-Gitlab-Event Header")
	}

	eventType := gitlab.EventType(event)
	if !isEventSubscribed(eventType, hook.EventsToAccept) {
		return nil, errors.New("event not defined to be parsed")
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil || len(payload) == 0 {
		return nil, errors.New("error reading request body")
	}

	return gitlab.ParseWebhook(eventType, payload)
}

func isEventSubscribed(event gitlab.EventType, events []gitlab.EventType) bool {
	return slices.Contains(events, event)
}
