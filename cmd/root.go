package cmd

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"regexp"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	"encoding/json"

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

var outputFormat string

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
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "format", "F", "text", "output format (text or json)")

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

// FeedDictionaryPaths takes a set of file paths (or - for stdin) and reads through
// each one, feeding it to the channel. Many of the puzzle types this helps with need
// to read from a dictionary file, so this creates a simple reusable pattern that
// any solving functionality can use.
func FeedDictionaryPaths(feed chan string, files ...string) {
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
	FeedDictionaryReaders(feed, readers...)
}

// FeedDictionaryReaders reads from readers and pushes strings to the feed,
// closing it when it's done. This is separated out from above largely to
// facilitate testing.
func FeedDictionaryReaders(feed chan string, readers ...*bufio.Reader) {
	for _, reader := range readers {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			feed <- strings.ToUpper(scanner.Text())
		}
	}
	close(feed)
}

// ReadDictionaryToTrie will read the dictionary channel populated by FeedDictionaryReaders
// and will add the items to a Trie structure that it will return
func ReadDictionaryToTrie(dictionary chan string) *TrieNode {
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

// PopulateFrequencyMapFromReader reads from a reader and populates a frequency map.
func PopulateFrequencyMapFromReader(reader io.Reader) (map[string]float64, int) {
	result := make(map[string]float64)
	scanner := bufio.NewScanner(reader)
	var ngramSize int
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, "\t")
		if len(fields) != 2 {
			fmt.Printf("Invalid line in frequency file: %s\n", line)
			os.Exit(1)
		}
		frequency, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			fmt.Printf("Invalid float in frequency file: %s\n", fields[1])
			os.Exit(1)
		}
		result[fields[0]] = frequency
		if ngramSize == 0 { // Set ngramSize from the first entry
			ngramSize = len(fields[0])
		}
	}
	return result, ngramSize
}

func outputToStdout(output []interface{}) {
	for _, line := range output {
		fmt.Println(line)
	}
}

func outputToJSON(output []interface{}) {
	jsonOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling to JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonOutput))
}

func outputResponse(output []interface{}) {
	switch outputFormat {
	case "json":
		outputToJSON(output)
	case "text":
		outputToStdout(output)
	default:
		fmt.Printf("Unknown output format: %s. Please use 'text' or 'json'.\n", outputFormat)
		os.Exit(1)
	}
}
