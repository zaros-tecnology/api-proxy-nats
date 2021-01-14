package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

//Hash implements root.Hash
type Hash struct{}

// Dispose general
func Dispose() {
	if p := recover(); p != nil {
		stack := debug.Stack()
		fmt.Println(fmt.Sprintf("Panic found: %v\nStack: %v", p, string(stack)))
	}
}

//Generate a salted hash for the input string
func (c *Hash) Generate(s string) (string, error) {
	saltedBytes := []byte(s)
	hashedBytes, err := bcrypt.GenerateFromPassword(saltedBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	hash := string(hashedBytes[:])
	return hash, nil
}

//Compare string to generated hash
func (c *Hash) Compare(hash string, s string) error {
	incoming := []byte(s)
	existing := []byte(hash)
	return bcrypt.CompareHashAndPassword(existing, incoming)
}

// GeneratePrivateKey string to crypto key private
func GeneratePrivateKey(hash string) (hashed string, err error) {
	if len(hash) == 0 {
		return "", errors.New("No input supplied")
	}

	hasher := sha256.New()
	hasher.Write([]byte(hash))

	stringToSHA256 := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	// Cut the length down to 32 bytes and return.
	return stringToSHA256[:32], nil
}

// GetBearer get
func GetBearer(r *http.Request) (accessToken string, ok bool) {
	auth := r.Header.Get("Authorization")
	prefix := "Bearer "

	if auth != "" && strings.HasPrefix(auth, prefix) {
		accessToken = auth[len(prefix):]
	} else {
		accessToken = r.FormValue("access_token")
	}

	if accessToken != "" {
		ok = true
	}

	return
}
