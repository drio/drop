package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type TemplateOpts struct {
	AddError string
	AddOk    bool
}

type ServerOpts struct {
	model   Model
	logger  Logger
	appName string
	inFly   bool
	port    string
	domain  string
}

type Server struct {
	model  Model
	logger Logger

	mux       *chi.Mux
	templates *template.Template

	appName string
	inFly   bool
	port    string
	domain  string
}

type Logger interface {
	Printf(format string, v ...interface{})
}

type Model interface {
	Add(string) (string, error)
	Get(string) (string, error)
	Delete(string) error
}

func NewServer(opts ServerOpts) (*Server, error) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	s := &Server{
		model:   opts.model,
		logger:  opts.logger,
		mux:     r,
		appName: opts.appName,
		inFly:   opts.inFly,
		port:    opts.port,
		domain:  opts.domain,
	}

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "assets"))
	FileServer(r, "/assets", filesDir)

	s.addRoutes()
	s.addTemplates()
	return s, nil
}

func (s *Server) addRoutes() {
	s.mux.Get("/", s.index)
	s.mux.Post("/d/ui", s.dropUI)
	s.mux.Post("/d", s.drop)
	s.mux.Get("/g/{uuid}", s.get)
}

func (s *Server) dropUI(w http.ResponseWriter, r *http.Request) {
	data := strings.TrimSpace(r.FormValue("data"))
	if data == "" {
		s.badRequestError(w, "no data", nil)
		return
	}

	s.doDrop(data, w)
}

func (s *Server) drop(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil || len(b) == 0 {
		http.Redirect(w, r, "/", http.StatusBadRequest)
		return
	}

	s.doDrop(string(b), w)
}

func (s *Server) doDrop(msg string, w http.ResponseWriter) {
	uuid, err := s.model.Add(msg)
	if err != nil {
		s.internalError(w, "error adding drop:", err)
		return
	}

	url := fmt.Sprintf("%s/g/%s", s.url(), uuid)
	_, err = w.Write([]byte(fmt.Sprintf("%s", url)))
	if err != nil {
		s.internalError(w, "error returning url:", err)
	}
}

func (s *Server) url() string {
	url := fmt.Sprintf("http://localhost:%s", s.port)
	if s.inFly {
		url = fmt.Sprintf("https://%s.%s", s.appName, s.domain)
	}
	return url
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		s.badRequestError(w, "uuid not provided", nil)
		return
	}

	content, err := s.model.Get(uuid)
	if err != nil {
		s.internalError(w, "error getting content:", err)
		return
	}

	if content == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = s.model.Delete(uuid)
	if err != nil {
		s.logger.Printf("error deleting after get drop:", err)
	}

	_, err = w.Write([]byte(fmt.Sprintf("%s", content)))
	if err != nil {
		s.internalError(w, "error returning content:", err)
	}
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	err := s.templates.ExecuteTemplate(w, "index", TemplateOpts{})
	if err != nil {
		s.logger.Printf("error executing index template: %s", err)
		s.templates.ExecuteTemplate(w, "error", nil)
	}
}

func (s *Server) addTemplates() {
	s.templates = template.Must(template.ParseFS(templateFS, "templates/*.gohtml"))
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	w.Header().Set("Cache-Control", "no-cache")
	s.mux.ServeHTTP(w, r)
	s.logger.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(startTime))
}

func (s *Server) internalError(w http.ResponseWriter, msg string, err error) {
	s.logger.Printf("error %s: %v", msg, err)
	http.Error(w, "error "+msg, http.StatusInternalServerError)
}

func (s *Server) badRequestError(w http.ResponseWriter, msg string, err error) {
	s.logger.Printf("error %s: %v", msg, err)
	http.Error(w, "error "+msg, http.StatusBadRequest)
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
