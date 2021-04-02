package server

/*
Copyright © 2020 Giuseppe Lavagetto

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/dyson/certman"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/lavagetto/memeoid/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Serve(ctl *api.Controller, certPath string, addr string, logger *log.Logger) error {

	// Add middlewares, logging
	addMiddleware(ctl.Router, true)
	// Handle the root of the site with the controller
	http.Handle("/", ctl.Router)
	// Add the static routes
	ctl.StaticRoute("/gifs/", ctl.Handler.ImgPath)
	ctl.StaticRoute("/meme/", ctl.Handler.OutputPath)
	// setup certmanager if the service is live.
	cm, err := setupCertManager(certPath, logger)
	if err != nil {
		return err
	}
	customLog := handlers.CombinedLoggingHandler(os.Stdout, http.DefaultServeMux)
	server := &http.Server{Addr: addr, Handler: customLog}
	logger.Printf("Listening on %s\n", addr)
	if cm != nil {
		server.TLSConfig = &tls.Config{GetCertificate: cm.GetCertificate}
		err = server.ListenAndServeTLS("", "")
	} else {
		err = server.ListenAndServe()
	}
	return err
}

func setupCertManager(certPath string, logger *log.Logger) (*certman.CertMan, error) {
	var cm *certman.CertMan
	var err error
	if certPath != "" {
		key := path.Join(certPath, "privkey.pem")
		fullchain := path.Join(certPath, "fullchain.pem")
		cm, err = certman.New(fullchain, key)
		if err == nil {
			cm.Logger(logger)
			err = cm.Watch()
		}
	}
	return cm, err
}

func addMiddleware(r *mux.Router, hasTLS bool) {
	// Add prometheus metrics
	r.Use(telemetryMiddleware)
	r.Path("/metrics").Handler(promhttp.Handler())
	if hasTLS {
		r.Use(hstsMiddleware)
	}
}

var httpDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "memeoid_http_duration_seconds",
		Help: "Duration of http requests",
	},
	[]string{"path", "gif"},
)

func telemetryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			route := mux.CurrentRoute(r)
			path, _ := route.GetPathTemplate()
			v := mux.Vars(r)
			gif := "-"
			if g, ok := v["from"]; ok {
				gif = g
			}
			timer := prometheus.NewTimer(httpDuration.WithLabelValues(path, gif))
			next.ServeHTTP(w, r)
			timer.ObserveDuration()
		},
	)
}

func hstsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Strict-Transport-Security", "max-age=864000")
			next.ServeHTTP(w, r)
		},
	)
}
