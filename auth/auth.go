package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	lg "github.com/mundacity/quick-logger"
)

type InvalidDurationError struct{}

func (e *InvalidDurationError) Error() string {
	return "duration 1 or less"
}

type AuthConfig struct {
	keyPath       string
	passwordHash  string
	durationHours int
	keyFunc       func(string) (*rsa.PrivateKey, error)
}

func NewAuthConfig(path, pwHash string, duration int) *AuthConfig {
	return &AuthConfig{
		keyPath:       path,
		passwordHash:  pwHash,
		durationHours: duration,
		keyFunc:       getPrivateKey,
	}
}

// Authenticate checks the request for an encoded password hash in {conf}. If the hash
// matches, a JWT is generated to give access for a duration specified in {conf}
func Authenticate(conf *AuthConfig, handlerFunc func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get encrypted password from header
		b64clientKey := r.Header.Get("Auth")
		if b64clientKey == "" {
			msg := "no token in header"
			lg.Logger.Log(lg.Warning, msg)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		key, err := base64.StdEncoding.DecodeString(b64clientKey)
		if err != nil {
			lg.Logger.Log(lg.Error, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		privateKey, err := conf.keyFunc(conf.keyPath)
		if err != nil {
			lg.Logger.Log(lg.Error, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// decrypt clientKey with server private key --> plain text version of user password
		userPasswordPlainText, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, []byte(key), []byte(""))
		if err != nil {
			lg.Logger.Log(lg.Error, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// hash user password & check against hash stored in app
		h := sha256.New()
		h.Write(userPasswordPlainText)
		n := h.Sum(nil)
		res := fmt.Sprintf("%x", n)

		if res == conf.passwordHash {
			jwt, err := generateJWT(privateKey, conf.durationHours)
			if err != nil {
				lg.Logger.Log(lg.Error, err.Error())
				http.Error(w, "jwt generation error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Auth", jwt)
			handlerFunc(w, r)
		}
		msg := "authentication error"
		lg.Logger.Log(lg.Info, msg)
		http.Error(w, msg, http.StatusInternalServerError)
	})
}

// RequestAuthentication is called by the client if the first request receives
// an unauthorised status code.
//
// It encrypts <userPw> using the server's public key and returns the encrypted
// password in base64 encoding
func RequestAuthentication(keyPath, userPw string, getPubKey func(string) (*rsa.PublicKey, error)) (string, error) {

	if getPubKey == nil {
		getPubKey = getPublicKey
	}

	pub, err := getPubKey(keyPath)
	if err != nil {
		return "", err
	}

	encryptedPw, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, []byte(userPw), []byte(""))
	if err != nil {
		return "", err
	}

	s := base64.StdEncoding.EncodeToString(encryptedPw)
	return s, nil
}

func getPrivateKey(keyPath string) (*rsa.PrivateKey, error) {

	// get file contents
	contents, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, errors.New("failed to retrieve key")
	}

	return encodePrivateKey(contents)
}

func encodePrivateKey(input []byte) (*rsa.PrivateKey, error) {
	for {

		block, rest := pem.Decode(input)
		if block.Type == "PRIVATE KEY" { // private key should be first so only one loop
			if k, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
				return k.(*rsa.PrivateKey), nil
			}
		}

		input = rest
	}
}

func getPublicKey(keyPath string) (*rsa.PublicKey, error) {
	// get file contents
	contents, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, errors.New("failed to retrieve key")
	}
	return encodePublicKey([]byte(contents))
}

func encodePublicKey(input []byte) (*rsa.PublicKey, error) {
	for {

		block, rest := pem.Decode(input)
		if block.Type == "PUBLIC KEY" {
			k, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err == nil {
				return k.(*rsa.PublicKey), nil
			}
		}

		input = rest
	}
}

func generateJWT(key *rsa.PrivateKey, dur int) (string, error) {

	if dur < 1 {
		return "", &InvalidDurationError{}
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(dur)).Unix()

	bs, _ := x509.MarshalPKCS8PrivateKey(key)
	pemdata := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: bs,
		},
	)

	tokenStr, err := token.SignedString(pemdata)
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

func ValidateJwt(keyPath string, handlerFunc func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Token"] == nil || r.Header["Token"][0] == "" {
			http.Error(w, "unauthorised", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(r.Header["Token"][0], func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				http.Error(w, "unauthorised", http.StatusUnauthorized)
				return nil, errors.New("unexpected signing method")
			}

			k, err := getPrivateKey(keyPath)
			if err != nil {
				http.Error(w, "key retrieval error", http.StatusInternalServerError)
			}

			bs, _ := x509.MarshalPKCS8PrivateKey(k)
			pemdata := pem.EncodeToMemory(
				&pem.Block{
					Type:  "PRIVATE KEY",
					Bytes: bs,
				},
			)

			return pemdata, nil
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		}

		if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			handlerFunc(w, r)
			return
		} else {
			http.Error(w, "invalid", http.StatusUnauthorized)
		}

	})
}
