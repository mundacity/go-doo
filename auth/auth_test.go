package auth

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	lg "github.com/mundacity/quick-logger"
)

var secretKey = "this is a super secret key"
var storedUserPasswordHash = ""

type test_case struct {
	secretKey   string
	testName    string
	expectedErr error
	statusCode  int
	conf        *AuthConfig
}

func getTests() []test_case {

	standard := getPrivateKey

	return []test_case{{
		secretKey:   secretKey,
		testName:    "wrong password",
		expectedErr: nil,
		statusCode:  http.StatusInternalServerError,
		conf: &AuthConfig{
			keyPath:       "",
			passwordHash:  "",
			durationHours: 8,
			keyFunc:       standard},
	}, {}, {}}
}

func TestAuthenticate_WrongPassword(t *testing.T) {

	lg.Logger = lg.NewDummyLogger()

	wr := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/authenticate", nil)

	r.Header.Set("Auth", "blah")

	cfg := NewAuthConfig("", "", 1)
	cfg.keyFunc = func(string) (*rsa.PrivateKey, error) {
		return nil, nil
	}
	f := func(w http.ResponseWriter, r *http.Request) {
		fmt.Print("test")
	}

	//f(wr, r)

	f2 := Authenticate(cfg, f)
	f2(wr, r)

}
