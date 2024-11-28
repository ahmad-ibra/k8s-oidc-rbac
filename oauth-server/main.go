package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const (
	clientID     = "0oalgq0kxjvutVEaR5d7"
	clientSecret = "dXgdX2rSTSCaorjBEQ4UbEInf8S_-YgnSM9jMOb1oaA4C735gVxhtQMTujGX_xGy"
	redirectURI  = "http://localhost:8080/authorization-code/callback"
	issuerURL    = "https://dev-04886319.okta.com/oauth2/default"
)

var (
	provider *oidc.Provider
	config   oauth2.Config
)

func main() {
	// Step 1: Initialize OIDC provider
	var err error
	provider, err = oidc.NewProvider(context.Background(), issuerURL)
	if err != nil {
		log.Fatalf("Failed to get OIDC provider: %v", err)
	}

	// Step 2: Configure OAuth2 client
	config = oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess},
	}

	// Step 3: Start local HTTP server for callback
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/authorization-code/callback", handleCallback)
	fmt.Println("Starting server at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Step 4: Serve the authorization URL
func handleIndex(w http.ResponseWriter, r *http.Request) {
	state := "randomstate123" // Use a secure random state in production
	authURL := config.AuthCodeURL(state, oauth2.SetAuthURLParam("prompt", "login"))
	fmt.Fprintf(w, "Visit the following URL to authenticate:\n\n%s", authURL)
}

// Step 5: Handle the OAuth2 callback
func handleCallback(w http.ResponseWriter, r *http.Request) {
	// Validate state (optional, but recommended)
	state := r.URL.Query().Get("state")
	if state != "randomstate123" {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// Exchange authorization code for tokens
	code := r.URL.Query().Get("code")
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Decode ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No ID token found", http.StatusInternalServerError)
		return
	}

	idToken, err := provider.Verifier(&oidc.Config{ClientID: clientID}).Verify(context.Background(), rawIDToken)
	if err != nil {
		http.Error(w, "Failed to verify ID token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract and display claims
	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, "Failed to parse claims: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Display tokens and claims
	fmt.Fprintf(w, "Access Token:\n%s\n\nRefresh Token:\n%s\n\nID Token:\n%s\n\nClaims:\n%v\n", token.AccessToken, token.RefreshToken, rawIDToken, claims)
}

// Utility: Encode client secret in Base64 (if needed for Basic Auth header)
func encodeClientSecret() string {
	creds := fmt.Sprintf("%s:%s", clientID, clientSecret)
	return base64.StdEncoding.EncodeToString([]byte(creds))
}
