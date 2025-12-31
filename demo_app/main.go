package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"

	"github.com/aritradeveops/porichoy/internal/core/jwtutil"
)

//go:embed public/*
var publicFS embed.FS

type TokenResponse struct {
	Message string `json:"message"`
	Data    struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	} `json:"data"`
}

func main() {
	subFS, err := fs.Sub(publicFS, "public")
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", http.FileServer(http.FS(subFS)))
	http.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		clientId := "demo.local"
		clientSecret := "6fcb20a444f96990ddd5d54a5471b2b41d55329ee97e6e18e0e370096c7eb3990e4c0468ed7006dd80c6aecd811daec1b1a7db4ff67da4b172c42efc125bebbc"
		redirectURI := "http://demo.local:5000/authorize"
		grantType := "authorization_code"

		query := url.Values{
			"client_id":     {clientId},
			"client_secret": {clientSecret},
			"grant_type":    {grantType},
			"code":          {code},
			"redirect_uri":  {redirectURI},
		}

		resp, err := http.Post("http://localhost:8080/api/v1/auth/token?"+query.Encode(), "application/x-www-form-urlencoded", nil)

		if err != nil || resp.StatusCode != http.StatusOK {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			w.Header().Set("Location", "/authorize/error")
			return
		}
		defer resp.Body.Close()

		var tokenResponse TokenResponse
		if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
			w.WriteHeader(http.StatusPermanentRedirect)
			w.Header().Set("Location", "/authorize/error")
			return
		}
		w.Header().Set("Set-Cookie", "access_token="+tokenResponse.Data.AccessToken)
		w.Header().Set("Location", "/profile")
		w.WriteHeader(http.StatusPermanentRedirect)
	})

	http.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("access_token")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return
		}
		userData, err := jwtutil.Verify(c.Value)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return
		}

		w.Write(fmt.Appendf(nil, "Hello %s! You are logged in.", userData.Email))
	})

	err = http.ListenAndServe(":5000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
