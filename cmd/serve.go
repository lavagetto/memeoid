package cmd

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
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/lavagetto/memeoid/api"
	"github.com/spf13/cobra"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var gifDir string
var memeDir string
var port int
var tplPath string
var certPath string

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "An http server to generate memes on request.",
	Long:  `At the moment memeoid only works with a local filesystem.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctl := api.Controller{
			Handler: &api.MemeHandler{
				ImgPath:    gifDir,
				OutputPath: memeDir,
				FontName:   fontName,
				MemeURL:    "meme",
			},
			Router: mux.NewRouter(),
		}

		ctl.StaticRoute("/gifs/", gifDir)
		ctl.StaticRoute("/meme/", memeDir)
		ctl.Load(tplPath)
		// Add prometheus metrics
		ctl.Router.Use(telemetryMiddleware)
		ctl.Router.Path("/metrics").Handler(promhttp.Handler())

		// Handle the root of the site with the controller
		http.Handle("/", ctl.Router)
		portStr := fmt.Sprintf(":%d", port)
		// Setup logging
		customLog := handlers.CombinedLoggingHandler(os.Stdout, http.DefaultServeMux)
		if certPath != "" {
			key := path.Join(certPath, "privkey.pem")
			fullchain := path.Join(certPath, "fullchain.pem")
			// Add an HSTS header
			ctl.Router.Use(hstsMiddleware)
			http.ListenAndServeTLS(portStr, fullchain, key, customLog)
		} else {
			http.ListenAndServe(portStr, customLog)
		}
	},
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

func init() {
	rootCmd.AddCommand(serveCmd)

	// flags and configuration settings.
	serveCmd.Flags().StringVarP(&gifDir, "image-dir", "i", "./fixtures", "The directory where base gifs are stored")
	serveCmd.Flags().StringVarP(&memeDir, "meme-dir", "m", "./memes", "The directory where memes are stored")
	serveCmd.Flags().IntVarP(&port, "port", "p", 3000, "The port to listen on")
	serveCmd.Flags().StringVar(&tplPath, "templates", "./templates", "Path to the teplate directory")
	serveCmd.Flags().StringVar(&certPath, "certpath", "", "Set this to your letsencrypt directory if you want TLS to work")
}
