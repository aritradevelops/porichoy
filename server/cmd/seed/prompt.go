package main

import (
	"bufio"
	"fmt"
	"net/mail"
	"os"
	"strings"

	"github.com/aritradevelops/porichoy/server/internal/identity"
	"golang.org/x/term"
)

// promptString reads a single non-empty line, re-prompting until one is given.
func promptString(r *bufio.Reader, label string) (string, error) {
	for {
		fmt.Print(label + ": ")
		line, err := r.ReadString('\n')
		if err != nil {
			return "", err
		}
		v := strings.TrimSpace(line)
		if v == "" {
			fmt.Println("This field is required.")
			continue
		}
		return v, nil
	}
}

// promptEmail is promptString plus a net/mail syntax check, re-prompting on an invalid
// address.
func promptEmail(r *bufio.Reader, label string) (string, error) {
	for {
		v, err := promptString(r, label)
		if err != nil {
			return "", err
		}
		if _, err := mail.ParseAddress(v); err != nil {
			fmt.Println("That doesn't look like a valid email address.")
			continue
		}
		return v, nil
	}
}

// promptDomain is promptString plus stripping a pasted-in scheme/path — TenantResolution
// matches on bare Host header values (e.g. "admin.example.com"), not a full URL, and pasting
// "https://admin.example.com/" is an easy mistake to make when copying from a browser.
func promptDomain(r *bufio.Reader, label string) (string, error) {
	for {
		v, err := promptString(r, label)
		if err != nil {
			return "", err
		}
		v = strings.TrimPrefix(v, "https://")
		v = strings.TrimPrefix(v, "http://")
		v = strings.TrimSuffix(v, "/")
		if v == "" {
			fmt.Println("This field is required.")
			continue
		}
		return v, nil
	}
}

// promptPassword reads a masked password (golang.org/x/term — no stdlib equivalent), asks
// for it twice, and re-prompts until both entries match and fall within the same length
// bounds the signup REST endpoint enforces (auth_handlers.go's signupRequest, min=8;
// identity.MaxPasswordBytes, bcrypt's own input limit).
func promptPassword(label string) (string, error) {
	for {
		fmt.Print(label + ": ")
		pw1, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return "", err
		}
		if len(pw1) < 8 {
			fmt.Println("Password must be at least 8 characters.")
			continue
		}
		if len(pw1) > identity.MaxPasswordBytes {
			fmt.Printf("Password must be %d characters or fewer.\n", identity.MaxPasswordBytes)
			continue
		}

		fmt.Print("Confirm " + label + ": ")
		pw2, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return "", err
		}
		if string(pw1) != string(pw2) {
			fmt.Println("Passwords didn't match, try again.")
			continue
		}
		return string(pw1), nil
	}
}
