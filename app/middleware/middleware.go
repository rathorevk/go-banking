package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/rathorevk/GoBanking/app/helpers"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		// Log the request details
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}

// Create panic handler
func PanicHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			error := recover()
			if error != nil {
				log.Println(error)

				resp := helpers.ErrorResponse{Error: "Internal server error"}
				json.NewEncoder(w).Encode(resp)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// SourceHeaderMatcher checks if the "Source" header is present in the request
func SourceHeaderMatcher(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sourceHeader := r.Header.Get("Source-Type")
		if !helpers.IsValidSource(sourceHeader) {
			http.Error(w, "Source type is invalid", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
