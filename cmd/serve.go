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

	"github.com/lavagetto/memeoid/api"
	"github.com/spf13/cobra"
)

var gifDir string
var memeDir string
var port int

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "An http server to generate memes on request.",
	Long:  `At the moment memeoid only works with a local filesystem.`,
	Run: func(cmd *cobra.Command, args []string) {
		memes := http.FileServer(http.Dir(memeDir))
		memeURL := "meme"
		memeHandler := &api.MemeHandler{
			ImgPath:    gifDir,
			OutputPath: memeDir,
			MemeURL:    memeURL,
			FontName:   fontName,
		}
		// Memes should be read from disk
		http.Handle(fmt.Sprintf("/%s/", memeURL), memes)
		// I "heart" the action api
		http.HandleFunc("/w/api.php", memeHandler.MemeFromRequest)
		http.HandleFunc("/gifs", memeHandler.ListGifs)
		portStr := fmt.Sprintf(":%d", port)
		fmt.Printf("Listening on %s\n", portStr)
		http.ListenAndServe(portStr, nil)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// flags and configuration settings.
	serveCmd.Flags().StringVarP(&gifDir, "image-dir", "i", "./fixtures", "The directory where base gifs are stored")
	serveCmd.Flags().StringVarP(&memeDir, "meme-dir", "m", "./memes", "The directory where memes are stored")
	serveCmd.Flags().IntVarP(&port, "port", "p", 3000, "The port to listen on")
}
