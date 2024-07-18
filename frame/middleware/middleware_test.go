package middleware

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func logger(logger *log.Logger) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Println("Begin Request:", r.URL.Path)
			defer logger.Println("End Request:", r.URL.Path)

			// setting a value on the request context
			ctx := context.WithValue(r.Context(), "logger", logger)

			// next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func timer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// accessing data from the request context
		logger := r.Context().Value("logger").(*log.Logger)

		timeStart := time.Now()
		logger.Println("time begin:", timeStart)

		// next handler
		next.ServeHTTP(w, r)

		timeElapsed := time.Since(timeStart)
		logger.Println("time elapsed:", timeElapsed)
	})
}

func TestRouter(t *testing.T) {
	r := NewRouter()
	r.Use(logger(log.New(os.Stdout, "[server] ", log.Lshortfile)))
	r.Use(timer)

	r.Add("/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))

		logger := r.Context().Value("logger").(*log.Logger)
		logger.Println("hello")
	}))

	req, err := http.NewRequest("GET", "/hello", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.mux["/hello"].ServeHTTP(rr, req)

	expected := "hello"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
