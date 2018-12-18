package main

import (
	"net/http"

	"github.com/geckoboard/slash-infra/slackutil"
	"github.com/julienschmidt/httprouter"
)

func makeHttpHandler() *httprouter.Router {
	router := httprouter.New()

	s := httpServer{}

	router.POST("/slack/infra-what-is", s.whatIsHandler)

	return router

}

type httpServer struct {
}

func respondWithError(w http.ResponseWriter, statusCode int, msg string) {
	w.WriteHeader(statusCode)
	w.Write([]byte(msg))
}

func (h httpServer) whatIsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	_, err := slackutil.ParseSlashCommandRequest(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "could not parse payload")
	}

	resp := slackutil.Response{
		Text: "One second while we look that up...",
	}

	slackutil.RespondWith(w, resp)
}
