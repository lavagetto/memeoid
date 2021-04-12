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
	"log"
	"os"

	"github.com/lavagetto/memeoid/api"
	"github.com/lavagetto/memeoid/server"
	"github.com/spf13/cobra"

	"github.com/gorilla/mux"
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
		// Error logging.
		// TODO: use it everywhere.
		logger := log.New(os.Stderr, "[memeoid] ", log.LstdFlags)
		ctl := api.Controller{
			Handler: &api.MemeHandler{
				ImgPath:    gifDir,
				OutputPath: memeDir,
				FontName:   fontName,
				MemeURL:    "meme",
				Logger:     logger,
			},
			Router: mux.NewRouter(),
		}
		ctl.Load(tplPath)

		portStr := fmt.Sprintf(":%d", port)
		err := server.Serve(&ctl, certPath, portStr, logger)
		if err != nil {
			logger.Fatalf("%v\n", err)
		}
	},
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
