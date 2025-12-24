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
		clientId := "localhost:5000"
		clientSecret := "a4ea4cfe1be4513a8a65581097c07f29af99d2e602248d9e3f2efc8952e93fae2d08c647cfccfe01f7ec54fd194632b0eb0c20c75e8231e72e610b19876d05c7"
		redirectURI := "http://localhost:5000/authorize"
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
