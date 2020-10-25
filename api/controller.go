package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

var httpVerbs = []string{"CONNECT", "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT", "TRACE"}

func handleMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	http.Error(w, fmt.Sprintf("Method %s not allowed.", r.Method), http.StatusMethodNotAllowed)
}

type Controller struct {
	Handler *MemeHandler
	Router  *mux.Router
}

func (r *Controller) routeFor(path string, f func(w http.ResponseWriter, r *http.Request), methods ...string) {
	subrouter := r.Router.Path(path).Subrouter()
	subrouter.Methods(methods...).HandlerFunc(f)
}

// Load sets up all dynamic routes.
func (r *Controller) Load(tplPath string) {
	r.Handler.LoadTemplates(tplPath)
	// Homepage
	r.routeFor("/", r.Handler.ListGifs, "GET", "HEAD")
	// Form
	r.routeFor("/generate", r.Handler.Form, "GET", "HEAD")
	r.api()
}

func (r *Controller) api() {
	// I "heart" the action api
	subrouter := r.Router.Path("/w/api.php").Methods("GET").Subrouter()
	subrouter.Queries("from", "{from}").HandlerFunc(r.Handler.MemeFromRequest)
	subrouter.Queries().HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "The 'from' parameter is required", http.StatusBadRequest)
	})
}

// StaticRoute sets up a static route
func (r *Controller) StaticRoute(uriPrefix string, docRoot string) {
	dir := http.FileServer(http.Dir(docRoot))
	r.Router.PathPrefix(uriPrefix).Handler(http.StripPrefix(uriPrefix, dir))
}
