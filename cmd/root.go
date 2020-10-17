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
	"math/rand"
	"os"
	"regexp"
	"runtime/pprof"
	"strings"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var profile bool
var cpuFilePath = "cpu.prof"
var memFilePath = "mem.prof"

var cpuFile *os.File
var memFile *os.File

// enough of these commands use a dictionary file that we can declare it at the top level
var dictionaryFile string

var lettersRegex = regexp.MustCompile("^[A-Za-z]+$")

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "puzzle_helper",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if profile {
			cpuFile, err := os.Create(cpuFilePath)
			if err != nil {
				fmt.Printf("Could not open %s: %v\n", cpuFilePath, err)
				os.Exit(1)
			}

			memFile, err = os.Create(memFilePath)
			if err != nil {
				fmt.Printf("Could not open %s: %v\n", memFilePath, err)
				os.Exit(1)
			}

			pprof.StartCPUProfile(cpuFile)
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if profile {
			cpuFile.Close()
			memFile.Close()
			pprof.StopCPUProfile()
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

	rand.Seed(time.Now().UnixNano())

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.puzzle_helper.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&profile, "profile", "", false, "turn on profiling for this run")

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

// dictionaryChanToTrie will read the dictionary channel populated by feedDictionaryReaders
// and will add the items to a Trie structure that it will return
func readDictionaryToTrie(dictionary chan string) *trieNode {
	now := time.Now().UnixNano()
	newTrie := newTrie()
	for entry := range dictionary {
		// something needs to be the value or else nodes will get ignored in walks
		err := newTrie.addValueForString(entry, nil)
		if err != nil {
			fmt.Printf("Could not add %s to trie %v\n", entry, err)
		}
	}
	if profile {
		fmt.Printf("Reading into trie took: %.8fms\n", float64(time.Now().UnixNano()-now)/float64(1000000))
	}
	return newTrie
}
