package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

type WebFingerResource struct {
	Subject string `json:"subject"`
	Links   []Link `json:"links"`
}

var (
	config      map[string]map[string]string
	configLock  sync.RWMutex
	configFile  = "config.yaml"
	reloadEvery = 30 * time.Second
	defaultUser string
)

func loadConfig() error {
	configLock.Lock()
	defer configLock.Unlock()

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Extract default user from config
	if val, ok := config["default"]["user"]; ok {
		defaultUser = val
		delete(config, "default") // Ensure "default" is not treated as a user entry
	} else {
		defaultUser = ""
	}

	return nil
}

func autoReloadConfig() {
	for {
		time.Sleep(reloadEvery)
		if err := loadConfig(); err != nil {
			log.Println("Error reloading config:", err)
		}
	}
}

func webfingerHandler(w http.ResponseWriter, r *http.Request) {
	resource := r.URL.Query().Get("resource")

	// If no resource is provided, use the default user
	if resource == "" {
		if defaultUser == "" {
			http.Error(w, "No default user specified", http.StatusNotFound)
			return
		}
		resource = defaultUser
	}

	// Remove "acct:" prefix if present
	resource = strings.TrimPrefix(resource, "acct:")

	configLock.RLock()
	userData, exists := config[resource]
	configLock.RUnlock()

	if !exists {
		http.Error(w, "Resource not found", http.StatusNotFound)
		return
	}

	// Construct WebFinger response
	response := WebFingerResource{
		Subject: fmt.Sprintf("acct:%s", resource),
		Links: []Link{
			{Rel: "http://webfinger.net/rel/profile-page", Href: userData["profile"]},
			{Rel: "http://webfinger.net/rel/avatar", Href: userData["avatar"]},
			{Rel: "http://openid.net/specs/connect/1.0/issuer", Href: userData["openid"]},
			{Rel: "https://tailscale.com/rel", Href: userData["tailscale"]},
			{Rel: "https://github.com", Href: userData["github"]},
			{Rel: "https://mastodon.social", Href: userData["mastodon"]},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Load config on startup
	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Start automatic config reloading
	go autoReloadConfig()

	http.HandleFunc("/.well-known/webfinger", webfingerHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("WebFinger server is running on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

