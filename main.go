package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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
	rel := r.URL.Query().Get("rel") // Specific relation requested

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

	// Mapping of rel values to YAML keys
	relMap := map[string]string{
		"http://webfinger.net/rel/profile-page":      "profile",
		"http://webfinger.net/rel/avatar":           "avatar",
		"http://openid.net/specs/connect/1.0/issuer": "openid",
		"https://tailscale.com/rel":                 "tailscale",
		"https://github.com":                        "github",
		"https://mastodon.social":                   "mastodon",
	}

	// ** Special Case: OpenID Connect Discovery **
	if rel == "http://openid.net/specs/connect/1.0/issuer" {
		if value, exists := userData["openid"]; exists && value != "" {
			u, err := url.Parse(value)
			if err == nil && u.Host != "" {
				w.Header().Set("Host", u.Host)
				w.WriteHeader(http.StatusNoContent) // 204 No Content, no body
				return
			}
		}
		http.Error(w, "Requested rel not found", http.StatusNotFound)
		return
	}

	// If a specific `rel` parameter is provided, return only that value in JSON
	if rel != "" {
		if key, ok := relMap[rel]; ok {
			if value, exists := userData[key]; exists && value != "" {
				response := WebFingerResource{
					Subject: fmt.Sprintf("acct:%s", resource),
					Links: []Link{
						{Rel: rel, Href: value},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
				return
			}
		}
		http.Error(w, "Requested rel not found", http.StatusNotFound)
		return
	}

	// Construct full WebFinger response
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

