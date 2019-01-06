package slackutil

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	bugsnag "github.com/bugsnag/bugsnag-go"
)

const (
	SlackRequestTimestampHeader = "X-Slack-Request-Timestamp"
	SlackSignatureHeader        = "X-Slack-Signature"
	SlackWebhookAllowedDelay    = time.Minute * 10
)

// Allows overriding now() in tests for the handler
var getNowTime = time.Now

func VerifyRequestSignature(secret string) func(next http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			bodyBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				bugsnag.Notify(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body.Close()
			r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

			signature := r.Header.Get(SlackSignatureHeader)
			timestamp := r.Header.Get(SlackRequestTimestampHeader)

			if !isRequestTimestampRecent(timestamp) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("request too old"))
				return
			}

			if !isRequestSignatureValid(secret, signature, timestamp, string(bodyBytes)) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("signature invalid"))
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

func isRequestTimestampRecent(timestamp string) bool {
	unixTimestamp, _ := strconv.Atoi(timestamp)
	t := time.Unix(int64(unixTimestamp), 0)

	if !getNowTime().Add(-SlackWebhookAllowedDelay).Before(t) {
		return false
	}

	return true
}

func isRequestSignatureValid(secret, signature, timestamp, body string) bool {
	computed := computeSlackSignature(secret, timestamp, body)

	return hmac.Equal([]byte(signature), []byte(computed))
}

func computeSlackSignature(secret, timestamp, body string) string {
	combined := fmt.Sprintf("v0:%s:%s", timestamp, body)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(combined))
	return fmt.Sprintf("v0=%s", hex.EncodeToString(mac.Sum(nil)))
}
