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
	"os"

	"image/gif"

	"github.com/lavagetto/memeoid/img"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var gifPath string
var topText string
var bottomText string
var outFile string
var fontName string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "memeoid",
	Short: "A non-cloud-native meme generator",
	Long: `Memeoid is a simple CLI or HTTP meme generator.
  	 Currently only CLI works, and it's extremely crude!`,
	Run: func(cmd *cobra.Command, args []string) {
		meme, err := img.MemeFromFile(
			gifPath,
			topText,
			bottomText,
			fontName,
		)
		if err != nil {
			panic(err)
		}
		// Uncomment for debugging
		/*
			meme.GifMetaData()
			return
		*/
		err = meme.Generate()
		if err != nil {
			panic(err)
		}
		out, err := os.Create(outFile)
		if err != nil {
			panic(err)
		}
		err = gif.EncodeAll(out, meme.Gif)
		if err != nil {
			panic(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.memeoid.yaml)")
	rootCmd.Flags().StringVar(&gifPath, "gif", "homer.gif", "The gif to use as a base for your meme")
	rootCmd.Flags().StringVarP(&topText, "top", "t", "", "The text to add at the top")
	rootCmd.Flags().StringVarP(&bottomText, "bottom", "b", "", "The text to insert at the bottom")
	rootCmd.Flags().StringVarP(&outFile, "out", "o", "meme.gif", "File to output to.")
	rootCmd.PersistentFlags().StringVarP(&fontName, "font", "f", "DejaVuSans", "Name of the ttf font on your system you want to use (default: impact).")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".memeoid" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".memeoid")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
