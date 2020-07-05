package api

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/labstack/echo"
	"net/http"
	"path/filepath"
	"strings"
)

type Handler struct {
	Api *Api
	Echo *echo.Echo
	Chi *chi.Mux
}

func NewHandler(api *Api) *Handler {
	h := new(Handler)
	c := chi.NewRouter()

	h.Api = api

	c.Use(middleware.RequestID)
	c.Use(middleware.RealIP)
	c.Use(middleware.Logger)
	c.Use(middleware.Recoverer)

	c.Group(func(r chi.Router) {
		r.Route("/todo", func(r chi.Router) {
			r.Get("/", TodoController{Api: h.Api}.Index)
			r.Get("/{id}", TodoController{Api: h.Api}.Show)
			r.Post("/", TodoController{Api: h.Api}.Create)
			r.Put("/{id}", TodoController{Api: h.Api}.Update)
			r.Delete("/{id}", TodoController{Api: h.Api}.Delete)
			r.Route("/{todoID}", func(r chi.Router) {
				r.Get("/file", TodoFileController{Api: h.Api}.Index)
				r.Post("/file", TodoFileController{Api: h.Api}.Create)
				r.Delete("/file/{id}", TodoFileController{Api: h.Api}.Delete)
			})
		})
	})

	h.FileServer(c)

	h.Chi = c
	return h
}

func (h *Handler) FileServer(router *chi.Mux) {
	p := "/files"

	if strings.ContainsAny(p, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if p != "/" && p[len(p)-1] != '/' {
		router.Get(p, http.RedirectHandler(p+"/", 301).ServeHTTP)
		p += "/"
	}
	p += "*"

	router.Get(p, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(http.Dir(filepath.Join(h.Api.Path, "files"))))
		fs.ServeHTTP(w, r)
	})
}

func (h *Handler) renderJSON(w http.ResponseWriter, statusCode int, response interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}





