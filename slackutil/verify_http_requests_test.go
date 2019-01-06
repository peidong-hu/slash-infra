package slackutil

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	// Taken from: https://api.slack.com/docs/verifying-requests-from-slack
	SlackTutorialSecret    = "8f742231b10e8888abcd99yyyzzz85a5"
	SlackTutorialTimestamp = 1531420618
	SlackTutorialSignature = "v0=a2114d57b48eac39b9ad189dd8316235a7b4a8d21a10bd27519666489c69b503"
	SlackTutorialBody      = "token=xyzz0WbapA4vBCDEFasx0q6G&team_id=T1DC2JH3J&team_domain=testteamnow&channel_id=G8PSS9T3V&channel_name=foobar&user_id=U2CERLKJA&user_name=roadrunner&command=%2Fwebhook-collect&text=&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2FT1DC2JH3J%2F397700885554%2F96rGlfmibIGlgcZRskXaIFfN&trigger_id=398738663015.47445629121.803a0bc887a14d10d2c447fce8b6703c"
)

func fixedTimeNow() time.Time {
	return time.Unix(SlackTutorialTimestamp, 0)
}

func mustMakeRequest(r *http.Response, err error) *http.Response {
	if err != nil {
		panic(err)
	}

	return r
}

func makeValidSlackRequest(url string) *http.Request {
	r, err := http.NewRequest("POST", url, strings.NewReader(SlackTutorialBody))
	if err != nil {
		panic(err)
	}
	r.Header.Set("X-Slack-Signature", SlackTutorialSignature)
	r.Header.Set("X-Slack-Request-Timestamp", fmt.Sprintf("%d", SlackTutorialTimestamp))
	return r
}
func TestSignatureVerification(t *testing.T) {
	getNowTime = fixedTimeNow
	defer func() { getNowTime = time.Now }()
	middleware := VerifyRequestSignature(SlackTutorialSecret)

	t.Run("When the signature is valid the request is passed to the handler", func(t *testing.T) {
		called := false
		handler := func(w http.ResponseWriter, r *http.Request) {
			called = true
		}

		server := httptest.NewServer(middleware(http.HandlerFunc(handler)))
		defer server.Close()

		r := makeValidSlackRequest(server.URL)
		resp := mustMakeRequest(http.DefaultClient.Do(r))

		if !called {
			t.Error("internal http handler was not called")
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("unexpected status code %d", resp.StatusCode)
		}
	})

	t.Run("When the signature is invalid the request is NOT passed to the handler", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Error("did not expect handler to be called")
		}

		server := httptest.NewServer(middleware(http.HandlerFunc(handler)))
		defer server.Close()

		r := makeValidSlackRequest(server.URL)
		r.Header.Set("X-Slack-Signature", "foobar")
		resp := mustMakeRequest(http.DefaultClient.Do(r))

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("unexpected status code %d", resp.StatusCode)
		}
	})

	t.Run("It verifies the timestamp is recent", func(t *testing.T) {
		// Move the clock into the future to simulate a replay
		// attack using a valid request, some time after the
		// request was originally made
		getNowTime = func() time.Time {
			return fixedTimeNow().Add(time.Minute * 30)
		}
		defer func() { getNowTime = fixedTimeNow }()

		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Error("did not expect handler to be called")
		}

		server := httptest.NewServer(middleware(http.HandlerFunc(handler)))
		defer server.Close()

		r := makeValidSlackRequest(server.URL)
		resp := mustMakeRequest(http.DefaultClient.Do(r))

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("unexpected status code %d", resp.StatusCode)
		}
	})

	t.Run("It ensures r.Body is reset so that subsequent handlers can read the request body", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Error(err)
			}

			if len(b) <= 0 {
				t.Error("expected non-empty body in handler")
			}

			if string(b) != SlackTutorialBody {
				t.Errorf("unexpected request body %q", b)
			}
		}

		server := httptest.NewServer(middleware(http.HandlerFunc(handler)))
		defer server.Close()

		r := makeValidSlackRequest(server.URL)
		resp := mustMakeRequest(http.DefaultClient.Do(r))

		if resp.StatusCode != http.StatusOK {
			t.Errorf("unexpected status code %d", resp.StatusCode)
		}
	})

}
