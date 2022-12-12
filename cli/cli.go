package cli

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/auth"
	lg "github.com/mundacity/quick-logger"
	"golang.org/x/term"
)

var CliContext godoo.ICliContext

// Feature in progress... TODO
// relates to priority queue and how remote storage returns the 'next' item
type priorityMode string

const (
	deadline priorityMode = "d"
	none     priorityMode = "n"
	low      priorityMode = "l"
	medium   priorityMode = "m"
	high     priorityMode = "h"
)

// if user is using a date range, get the upper bound of that range
func getUpperDateBound(dateText string, dateLayout string) time.Time {
	splt := splitDates(dateText)
	var d time.Time

	if len(splt) > 1 {
		d, _ = time.Parse(dateLayout, splt[1])
	}

	return d
}

func splitDates(s string) []string {
	return strings.Split(s, ":")
}

func sendRequest(r *http.Request, c *http.Client) (*http.Response, error) {
	resp, err := c.Do(r)
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("error receiving response: %v", err), runtime.Caller)
		return nil, err
	}

	return resp, nil
}

func authenticateUser(pubKeyPath string, c *http.Client, r *http.Request) (string, error) {
	pw, err := requestPassword()
	if err != nil {
		return "", err
	}
	auth.RequestAuthentication(r, pubKeyPath, pw)

	rsp, err := sendRequest(r, c)

	if err != nil {
		return "", err
	}

	jwt := rsp.Header.Get("Auth")
	if jwt == "" {
		return jwt, errors.New("jwt blank")
	}

	return jwt, nil
}

func requestPassword() (string, error) {

	lg.Logger.Log(lg.Info, "requesting user password")
	fmt.Print("\nEnter password to authenticate: \n")
	pwd, err := term.ReadPassword(int(os.Stdin.Fd()))

	if err != nil {
		return "", err
	}

	return string(pwd), nil
}
