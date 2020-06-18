package server

import (
	"expertisetest/config"
	"expertisetest/server/middlewares"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
)

// Server ...
type Server struct {
	router chi.Router
	config *config.Config
}

// New returns a new Server.
func New() *Server {
	s := chi.NewRouter()

	mws := []func(http.Handler) http.Handler{
		middleware.RequestID,
		middleware.RealIP,
		middleware.Recoverer,
		middlewares.LoggerMiddleware,
	}

	for _, mw := range mws {
		s.Use(mw)
	}

	s.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	return &Server{
		router: s,
	}
}

// Serve serves the server. :P
func (s *Server) Serve() {

	errChan := make(chan error, 1)
	defer close(errChan)

	// Check for errors.
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	// If errors log and exit
	go func() {
		logrus.WithFields(logrus.Fields{"transport": "http", "state": "listening"}).Info("http init")
		errChan <- http.ListenAndServe(":80", s.router)
	}()

	logrus.WithFields(logrus.Fields{"transport": "http", "state": "terminated"}).Error(<-errChan)
	os.Exit(1)
}
