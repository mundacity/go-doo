package auth

import (
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"

	lg "github.com/mundacity/quick-logger"
)

var secretKey = "this is a super secret key"
var storedUserPasswordHash = "13bc6c9df8c126a8d276d92959ef1f5dc27d3fba077b6e93730d249d74ff0904"

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

var fakePrivRsa string = `-----BEGIN PRIVATE KEY-----
MIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBANJ7daSoC0CJQXhH1/IFlPsNesJhsSZgajOORhDFjyqUhvZh1Il0yV9EEL4DVeOz1UebFaCNjfEt5udWUSyVwukwIlTqWRzxkzz2aPsB8nBcLrx15+1OkbU055tVVlcQjATytY6WrTl6AN18i/xwQwddO0iOTxe4RsnsG4onyLaPAgMBAAECgYAWIasmDBFa0NPUfOFk7ldS6oDs7W6+FUc1cpFFdDBwjrt+Lp01ctU1sid8g0dFsQQNCm6Euj2hjW0JCBdy87BRubb/xhAEL3XzhkqUrV81sw12Y/kUDsNxbrlgAN7Cj7U5Gsh0rxVbpNIC7yvJ/vRyeCRj/175tvze+Xxvhmwz4QJBAPdRdGgHzqpq18i42GpJepAzIZ2C7B6I9+aZ28P8gKlz/RMdkci3Iu4rEVF1KVeqJHjSLAgyOcNFmzGtcphlAr8CQQDZ3vnOTAVySKRTBQ+cTtTyUMXsShtiWolKL+1TzWUT/17ETgq3OVbrntMlTdZf8hrH6/w/3QDOFj94NrN7yNAxAkADAAb0eBvGr3McqTle2LNW6nfe7Eam/CxdrMIgt4BsDc8lGze4gpg24WjdXxl4ScUVfh8wnkNbHg4K5Tq9pIQLAkEAzlT/G0KfvdXR2dXnLM7zmPCqINcmDAVWE+5DwqO4YDHvG9YVC+S/zrFBogiPR5pPhpqU8B5rDsG/JigX3tkVYQJBAKHbJZZ46KMvQOxnC4W7wX7sq2abJvVa2Um4EXI/gR5kNIILDM2raZbMGj5nPIplUc4wSxLNcJ3fGe4hAM0ceOk=
-----END PRIVATE KEY-----`

var fakePubRsa string = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDSe3WkqAtAiUF4R9fyBZT7DXrCYbEmYGozjkYQxY8qlIb2YdSJdMlfRBC+A1Xjs9VHmxWgjY3xLebnVlEslcLpMCJU6lkc8ZM89mj7AfJwXC68deftTpG1NOebVVZXEIwE8rWOlq05egDdfIv8cEMHXTtIjk8XuEbJ7BuKJ8i2jwIDAQAB
-----END PUBLIC KEY-----`

func TestAuthenticate_WrongHeader(t *testing.T) { //TODO

	lg.Logger = lg.NewDummyLogger()

	wr := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/authenticate", nil)

	// invalid header
	r.Header.Set("test", "blah")

	cfg := NewAuthConfig("", storedUserPasswordHash, 1)
	cfg.keyFunc = getFakePrivRsa

	f := func(w http.ResponseWriter, r *http.Request) {
		// will only change if authentication succeeds
		w.WriteHeader(http.StatusOK)
	}

	f2 := Authenticate(cfg, f)
	f2(wr, r)

	exp := http.StatusBadRequest
	if wr.Code == exp {
		t.Logf(">>>>PASS: expected <%v> and got <%v>", exp, wr.Code)
	} else {
		t.Errorf(">>>>FAIL: expected <%v> and got <%v>", exp, wr.Code)
	}

	// valid header, garbage password
	wr = httptest.NewRecorder()
	r.Header.Set("Auth", "blah")
	f2(wr, r)

	exp = http.StatusInternalServerError
	if wr.Code == exp {
		t.Logf(">>>>PASS: expected <%v> and got <%v>", exp, wr.Code)
	} else {
		t.Errorf(">>>>FAIL: expected <%v> and got <%v>", exp, wr.Code)
	}

	// valid header & password
	wr = httptest.NewRecorder()
	k, _ := RequestAuthentication("", secretKey, getFakePubRsa)
	r.Header.Set("Auth", k)

	f2(wr, r)

	exp = http.StatusOK
	if wr.Code == exp {
		t.Logf(">>>>PASS: expected <%v> and got <%v>", exp, wr.Code)
	} else {
		t.Errorf(">>>>FAIL: expected <%v> and got <%v>", exp, wr.Code)
	}

}

func getFakePrivRsa(dud string) (*rsa.PrivateKey, error) {
	val := []byte(fakePrivRsa)
	return encodePrivateKey(val)
}

func getFakePubRsa(dud string) (*rsa.PublicKey, error) {
	val := []byte(fakePubRsa)
	return encodePublicKey(val)
}
