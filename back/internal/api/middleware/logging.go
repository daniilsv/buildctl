package middleware

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func Logging(next http.Handler) http.Handler {
	return middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger:  log.New(os.Stdout, "chi", 0),
		NoColor: false,
	})(next)
}

func RequestID(next http.Handler) http.Handler {
	return middleware.RequestID(next)
}

func Recoverer(next http.Handler) http.Handler {
	return middleware.Recoverer(next)
}

func Timeout(duration time.Duration) func(http.Handler) http.Handler {
	return middleware.Timeout(duration)
}
