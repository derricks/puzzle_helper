/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "puzzle_helper",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.puzzle_helper.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

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

		// Search config in home directory with name ".puzzle_helper" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".puzzle_helper")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// feedDictionaryPaths takes a set of file paths (or - for stdin) and reads through
// each one, feeding it to the channel. Many of the puzzle types this helps with need
// to read from a dictionary file, so this creates a simple reusable pattern that
// any solving functionality can use.
func feedDictionaryPaths(feed chan string, files ...string) {
	readers := make([]*bufio.Reader, 0, len(files))
	for _, file := range files {
		if file == "-" {
			readers = append(readers, bufio.NewReader(os.Stdin))
		} else {
			file, err := os.Open(file)
			if err != nil {
				fmt.Printf("Could not access file: %v\n", err)
				os.Exit(1)
			}
			defer file.Close()
			readers = append(readers, bufio.NewReader(file))
		}
	}
	feedDictionaryReaders(feed, readers...)
}

// feedDictionaryReaders reads from readers and pushes strings to the feed,
// closing it when it's done. This is separated out from above largely to
// facilitate testing.
func feedDictionaryReaders(feed chan string, readers ...*bufio.Reader) {
	for _, reader := range readers {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			feed <- strings.ToUpper(scanner.Text())
		}
	}
	close(feed)
}
