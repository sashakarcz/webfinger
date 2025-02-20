package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Define a struct for WebFinger responses
type WebFingerResponse struct {
	Subject    string            `json:"subject"`
	Links      []WebFingerLink   `json:"links,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`
}

// Define a struct for each WebFinger link
type WebFingerLink struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

// Global config map
var (
	configFile  = "config.yaml"
	webFingerData map[string]map[string]string
	configMutex   sync.RWMutex
)

// Load YAML config
func loadConfig() error {
	configMutex.Lock()
	defer configMutex.Unlock()

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(yamlFile, &webFingerData)
	if err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	log.Println("✔ Config reloaded successfully")
	return nil
}

// Watch the YAML file for changes
func watchConfig() {
	for {
		time.Sleep(10 * time.Second) // Check every 10 seconds

		err := loadConfig()
		if err != nil {
			log.Println("❌ Failed to reload config:", err)
		}
	}
}

// WebFinger Handler
func webFingerHandler(w http.ResponseWriter, r *http.Request) {
	resource := r.URL.Query().Get("resource")
	if resource == "" {
		http.Error(w, "Missing resource parameter", http.StatusBadRequest)
		return
	}

	// Strip "acct:" prefix if present
	if strings.HasPrefix(resource, "acct:") {
		resource = strings.TrimPrefix(resource, "acct:")
	}

	configMutex.RLock()
	data, found := webFingerData[resource]
	configMutex.RUnlock()

	if !found {
		log.Printf("404 Not Found for resource: %s\n", resource)
		http.Error(w, "Resource not found", http.StatusNotFound)
		return
	}

	// Build WebFinger response
	response := WebFingerResponse{
		Subject:    "acct:" + resource,
		Properties: make(map[string]string),
	}

	// Map YAML attributes to WebFinger links
	for key, value := range data {
		if value == "" {
			continue
		}
		switch key {
		case "avatar":
			response.Links = append(response.Links, WebFingerLink{Rel: "http://webfinger.net/rel/avatar", Href: value})
		case "openid":
			response.Links = append(response.Links, WebFingerLink{Rel: "http://specs.openid.net/auth/2.0/provider", Href: value})
		case "github":
			response.Links = append(response.Links, WebFingerLink{Rel: "https://github.com/", Href: value})
		case "mastodon":
			response.Links = append(response.Links, WebFingerLink{Rel: "http://joinmastodon.org/", Href: value})
		case "tailscale":
			response.Links = append(response.Links, WebFingerLink{Rel: "https://login.tailscale.com/", Href: value})
		case "profile":
			response.Links = append(response.Links, WebFingerLink{Rel: "http://webfinger.net/rel/profile-page", Href: value})
		default:
			response.Properties[key] = value
		}
	}

	// Encode response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Load config initially
	if err := loadConfig(); err != nil {
		log.Fatalf("❌ Error loading config: %v", err)
	}

	// Start config watcher in the background
	go watchConfig()

	http.HandleFunc("/.well-known/webfinger", webFingerHandler)
	log.Println("🌐 WebFinger server running on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

