# puzzle_helper
A command-line tool for doing a wide range of puzzle-solving tasks

## Description
This tool serves two purposes: Aggregate a variety of scripts and programs I use
for designing and solving puzzles under one umbrella and give me exposure to the Cobra
tool for creating CLIs in golang.

## Usage:
Generate a frequency table of characters within the set of strings

    ./puzzle_helper cryptogram freq string1 [string2...]

Provide a REPL for interactively solving substitution-type cryptograms
Commands:
  - A=e -> make uppercase ciphertext A represent lowercase plaintext e
  - cipher2Plain -> (Default) Display the keys with ciphertext on top and in alphabetical order
  - plain2Cipher -> Display the keys with plaintext on top and in alphabetical order
  - clear -> clears out all cipher -> plain mappings


      ./puzzle_helper cryptogram substitution repl string1 [string2...]

Given a dictionary file, attempt to find a set of cribs that matches the ciphertext.

    ./puzzle_helper cryptogram substitution solve string1 [string2...] --dictionary path_to_dictionary_file

Given a set of strings, print out the caesar shifts of those strings

    ./puzzle_helper cryptogram caesar string1 [string2...]

The `solve` command will attempt to solve the set of strings concurrently. You can configure the number of goroutines that will get made for parallel solving with the --concurrency argument (default is 10):

    ./puzzle_helper cryptogram substitution solve string1 [string2...] --dictionary path_to_dictionary_file -concurrency 2
