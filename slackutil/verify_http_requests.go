package slackutil

import (
	"bytes"
	"io/ioutil"
	"net/http"

	bugsnag "github.com/bugsnag/bugsnag-go"
)

func VerifyRequestSignature(secret string, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			bugsnag.Notify(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad request"))
			return
		}
		r.Body.Close()
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		// TODO: actually verify :facepalm:

		next.ServeHTTP(w, r)
	}
}
