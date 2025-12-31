package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Porichoy",
	Long:  `Login to Porichoy`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Flags().StringP("username", "u", "", "username (email) of the root user.")
		cmd.Flags().StringP("password", "p", "", "password of the root user.")
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		if username == "" {
			survey.AskOne(&survey.Input{Message: "Username(email):"}, &username)
		}
		if password == "" {
			survey.AskOne(&survey.Password{Message: "Password:"}, &password)
		}

		resp, err := http.Post("http://localhost:8080/api/v1/auth/login", "application/json", bytes.NewReader(fmt.Appendf(nil, `{"email": "%s", "password": "%s"}`, username, password)))
		if err != nil {
			fmt.Println("Error logging in:", err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Println("Error logging in:", resp.Status)
			return
		}
		fmt.Println("Login successful")
		type LoginResponse struct {
			Message string `json:"message"`
			Data    struct {
				AccessToken string `json:"access_token"`
			} `json:"data"`
		}
		var loginResponse LoginResponse
		if err := json.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
			fmt.Println("Error decoding response:", err)
			return
		}
		fmt.Println("Login successful")
		fmt.Println("Access token:", loginResponse.Data.AccessToken)
		fmt.Println("Storing access token in keychain")
		err = keyring.Set("porichoy", "access_token", loginResponse.Data.AccessToken)
		if err != nil {
			fmt.Println("Error storing access token in keychain:", err)
			return
		}
		fmt.Println("Access token stored in keychain")
	},
}
