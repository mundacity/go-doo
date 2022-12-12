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

// authenticate user and generate jwt
func Authenticate(r *http.Request, keyPath, userPasswordHash string) (string, error) {

	// get encrypted password from header
	clientKey := r.Header.Get("Authorization")
	if clientKey == "" {
		msg := "no token in header"
		lg.Logger.Log(lg.Info, msg)
		return msg, errors.New(msg)
	}

	privateKey, err := getPrivateKey(keyPath)
	if err != nil {
		return err.Error(), err
	}

	// decrypt clientKey with server private key --> plain text version of user password
	userPasswordPlainText, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, []byte(clientKey), []byte(""))
	if err != nil {
		return "", err //TODO: label needs to be same as that used when encrypting --> add to client
	}

	// hash user password & check against hash stored in app
	hashPw := sha256.Sum256(userPasswordPlainText)
	if string(hashPw[:]) == userPasswordHash {
		// then it's me so generate jwt with 8 hour expiry
		return generateJWT(privateKey)
	}
	msg := "authentication error"
	return msg, errors.New(msg)
}

func RequestAuthentication(r *http.Request, keyPath, userPw string) error {

	pub, err := GetPublicKey(keyPath)
	if err != nil {
		return err
	}

	encryptedPw, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, []byte(userPw), []byte(""))
	if err != nil {
		return err
	}

	s := base64.StdEncoding.EncodeToString(encryptedPw)

	r.Header.Set("Auth", s)
	return nil
}

func getPrivateKey(keyPath string) (*rsa.PrivateKey, error) {

	// get file contents
	contents, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, errors.New("failed to retrieve key")
	}

	for {

		block, rest := pem.Decode(contents)
		if block.Type == "RSA PRIVATE KEY" { // private key should be first so only one loop
			if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
				return k, nil
			}
		}

		contents = rest
	}
}

func GetPublicKey(keyPath string) (*rsa.PublicKey, error) {
	// get file contents
	contents, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, errors.New("failed to retrieve key")
	}

	for {

		block, rest := pem.Decode(contents)
		if block.Type == "PUBLIC KEY" {
			k, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err == nil {
				return k.(*rsa.PublicKey), nil
			}
		}

		contents = rest
	}
}

func generateJWT(key *rsa.PrivateKey) (string, error) {

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour * 8).Unix()

	tokenStr, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

func ValidateToken(tokenString, keyPath string) error {

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		k, err := getPrivateKey(keyPath)
		if err != nil {
			return nil, errors.New("key retrieval error")
		}

		return k, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println(claims["foo"], claims["nbf"])
	} else {
		fmt.Println(err)
	}
	return nil
}

func ValidateJwt(keyPath string, handlerFunc func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Token"] == nil || r.Header["Token"][0] == "" {
			http.Error(w, "missing header/s", http.StatusBadRequest)
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

			return k, nil
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
