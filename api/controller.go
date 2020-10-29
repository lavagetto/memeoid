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

func (r *Controller) routeFor(path string, f func(w http.ResponseWriter, r *http.Request), requireFrom bool, methods ...string) {
	subrouter := r.Router.Path(path).Methods(methods...).Subrouter()
	if requireFrom {
		subrouter.Queries("from", "{from}").HandlerFunc(f)
		subrouter.Queries().HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "The 'from' parameter is required", http.StatusBadRequest)
		})
	} else {
		subrouter.HandleFunc("/", f)
	}
}

// Load sets up all dynamic routes.
func (r *Controller) Load(tplPath string) {
	r.Handler.LoadTemplates(tplPath)
	// Homepage
	r.routeFor("/", r.Handler.ListGifs, false, "GET", "HEAD")
	// Form
	r.routeFor("/generate", r.Handler.Form, true, "GET", "HEAD")
	// I "heart" the action api
	r.routeFor("/w/api.php", r.Handler.MemeFromRequest, true, "GET")
	// Thumbnails
	r.Router.Path("/thumb/{width:[0-9]+}x{height:[0-9]+}/{from}").Methods("GET", "HEAD").HandlerFunc(r.Handler.Preview)
}

// StaticRoute sets up a static route
func (r *Controller) StaticRoute(uriPrefix string, docRoot string) {
	dir := http.FileServer(http.Dir(docRoot))
	r.Router.PathPrefix(uriPrefix).Handler(http.StripPrefix(uriPrefix, dir))
}
