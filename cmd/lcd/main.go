// Copyright 2018 Łukasz Pankowski <lukpank at o2 dot pl>. All rights
// reserved.  This source code is licensed under the terms of the MIT
// license. See LICENSE file for details.

/*

# Add the following to your ~/.bashrc:

lcd() {
    declare paths
    paths=$("$HOME/go/bin/lcd" -menu -output_fd 3 -- "$@" 3>&1 >/dev/tty)
    if [ $? -eq 0 ]; then
	cd "${paths}"
    elif [ ! -z "${paths}" ]; then
	echo "${paths}"
    fi
}

_lcd() {
    declare cur="${COMP_WORDS[COMP_CWORD]}"
    COMPREPLY=( $("$HOME/go/bin/lcd" -complete "${cur}") )
}

complete -F _lcd lcd

*/

package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
)

var errSilentExit1 = errors.New("silent exit(1)")

func main() {
	if err := run(); err != nil {
		if err != errSilentExit1 {
			fmt.Fprintf(os.Stderr, "lcd: %v\n", err)
		}
		os.Exit(1)
	}
}

func run() error {
	compl := flag.String("complete", "", "list completions")
	menu := flag.Bool("menu", false, "show menu if more than one directory")
	outputFd := flag.Int("output_fd", 0, "menu output mode")
	flag.Parse()
	output := os.Stdout
	if *outputFd != 0 {
		output = os.NewFile(uintptr(*outputFd), fmt.Sprintf("/dev/fd/%d", *outputFd))
	}
	if flag.NArg() < 1 && *compl == "" && !*menu {
		return nil
	}
	f, err := os.Open(filepath.Join(os.Getenv("HOME"), ".lcd", "cache"))
	if err != nil {
		return err
	}
	defer f.Close()

	if *compl != "" {
		return complete(*compl, output, f)
	}
	nArg := flag.NArg()
	if nArg > 1 {
		n, err := strconv.Atoi(flag.Arg(1))
		if err == nil {
			return matchingN(flag.Arg(0), n, output, f)
		}
	}
	if *menu {
		return matchingWithMenu(flag.Arg(0), output, f)
	} else if nArg > 0 {
		return matching(flag.Arg(0), output, f)
	}
	return nil
}

const pathSep = string(os.PathSeparator)

var pathSepB = []byte{os.PathSeparator}

func matching(word string, w io.Writer, r io.Reader) error {
	return matchingF(word, "", r, func(path string) bool {
		fmt.Fprintln(w, path)
		return true
	})
}

func matchingN(word string, idx int, w io.Writer, r io.Reader) error {
	i := 0
	return matchingF(word, " "+strconv.Itoa(idx), r, func(path string) bool {
		i++
		if i == idx {
			fmt.Fprintln(w, path)
			return true
		}
		return false
	})
}

func matchingPaths(word string, r io.Reader) ([]string, error) {
	paths := []string{}
	if word == "" {
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			path := strings.TrimSuffix(sc.Text(), pathSep)
			if st, err := os.Stat(path); err == nil && st.IsDir() {
				paths = append(paths, path)
			}
		}
		return paths, sc.Err()
	}
	err := matchingF(word, "", r, func(path string) bool {
		paths = append(paths, path)
		return true
	})
	return paths, err
}

func matchingF(word, msgSuffix string, r io.Reader, fn func(string) bool) error {
	suffix := []byte(pathSep + strings.TrimSuffix(word, pathSep))
	found := false
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := bytes.TrimSuffix(sc.Bytes(), pathSepB)
		if !bytes.HasSuffix(line, suffix) {
			continue
		}
		s := string(line)
		if st, err := os.Stat(s); err == nil && st.IsDir() && fn(s) {
			found = true
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("%q%s: directory not found", word, msgSuffix)
	}
	return nil
}

func complete(prefix string, w io.Writer, r io.Reader) error {
	seen := make(map[string]struct{})
	prefixB := []byte(pathSep + prefix)
	n := strings.Count(prefix, pathSep)
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := bytes.TrimSuffix(sc.Bytes(), pathSepB)
		i := bytes.LastIndex(line, prefixB)
		if i == -1 || bytes.Count(line[i+1:], pathSepB) > n {
			continue
		}
		completion := string(line[i+1:])
		if completion == "" {
			continue
		}
		s := string(line)
		st, err := os.Stat(s)
		if err != nil || !st.IsDir() {
			continue
		}
		if _, present := seen[completion]; !present {
			fmt.Fprintln(w, completion)
			seen[completion] = struct{}{}
		}
	}
	return sc.Err()
}

func matchingWithMenu(word string, w io.Writer, r io.Reader) error {
	paths, err := matchingPaths(word, r)
	if err != nil {
		return err
	}
	switch {
	case len(paths) == 0:
		return fmt.Errorf("%q: directory not found", word)
	case len(paths) == 1:
		fmt.Fprintln(w, paths[0])
		return nil
	case os.Getenv("TERM") == "dumb":
		fmt.Fprintln(w, strings.Join(paths, "\n"))
		return errSilentExit1
	default:
		prompt := promptui.Select{
			Label: "Change directory",
			Items: paths,
			Size:  10,
			Searcher: func(input string, index int) bool {
				for _, s := range strings.Fields(input) {
					if !strings.Contains(paths[index], s) {
						return false
					}
				}
				return true
			},
		}
		_, result, err := prompt.Run()
		if err != nil {
			return err
		}
		fmt.Fprintln(w, result)
		return nil
	}
}
