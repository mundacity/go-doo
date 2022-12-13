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
func Authenticate(keyPath, userPasswordHash string, handlerFunc func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc { //(string, error) {

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

		privateKey, err := getPrivateKey(keyPath)
		if err != nil {
			lg.Logger.Log(lg.Error, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// decrypt clientKey with server private key --> plain text version of user password
		userPasswordPlainText, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, []byte(key), []byte(""))
		if err != nil {
			//return "", err //TODO: label needs to be same as that used when encrypting --> add to client
			lg.Logger.Log(lg.Error, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// hash user password & check against hash stored in app
		h := sha256.New()
		h.Write(userPasswordPlainText)
		n := h.Sum(nil)
		fm := fmt.Sprintf("%x", n)

		if fm == userPasswordHash {
			// then it's me so generate jwt with 8 hour expiry
			jwt, err := generateJWT(privateKey)
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

func getPrivateKey(keyPath string) (*rsa.PrivateKey, error) {

	// get file contents
	contents, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, errors.New("failed to retrieve key")
	}

	for {

		block, rest := pem.Decode(contents)
		if block.Type == "PRIVATE KEY" { // private key should be first so only one loop
			if k, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
				return k.(*rsa.PrivateKey), nil
			}
		}

		contents = rest
	}
}

func GetPublicKey(keyPath string) (string, error) {
	// get file contents
	contents, err := os.ReadFile(keyPath)
	if err != nil {
		return "", errors.New("failed to retrieve key")
	}

	for {

		block, rest := pem.Decode(contents)
		if block.Type == "PUBLIC KEY" {
			_, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err == nil {
				return string(block.Bytes), nil
			}
		}

		contents = rest
	}
}

func generateJWT(key *rsa.PrivateKey) (string, error) {

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour * 8).Unix()

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
			http.Error(w, "unauthorised", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(r.Header["Token"][0], func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				http.Error(w, "unauthorised", http.StatusUnauthorized)
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
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

		if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			handlerFunc(w, r)
			return
		} else {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		}

	})
}
