package main

import (
	"log"
	"net/http"
	"os"

	"github.com/geckoboard/slash-infra/slackutil"
	"github.com/joho/godotenv"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)

	err := godotenv.Load()
	if err != nil {
		log.Println("could not load .env file", err)
	}

	server := makeHttpHandler()

	handler := slackutil.VerifyRequestSignature(os.Getenv("SLACK_SIGNING_SECRET"))(server)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	log.Fatal(http.ListenAndServe(":"+port, handler))
}
