// Copyright 2018 Łukasz Pankowski <lukpank at o2 dot pl>. All rights
// reserved.  This source code is licensed under the terms of the MIT
// license. See LICENSE file for details.

/*

# Add the following to your ~/.bashrc:

lcd() {
    declare paths
    paths=$("$HOME/go/bin/lcd" -- "$@")
    if [ $(echo "${paths}" | wc -l) -eq 1 ]; then
	cd "${paths}"
    else
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
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	compl := flag.String("complete", "", "list completions")
	flag.Parse()
	if flag.NArg() < 1 && *compl == "" {
		return
	}
	f, err := os.Open(filepath.Join(os.Getenv("HOME"), ".lcd", "cache"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "lcd: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if *compl != "" {
		if err := complete(*compl, os.Stdout, f); err != nil {
			fmt.Fprintf(os.Stderr, "lcd: %v\n", err)
			os.Exit(1)
		}
	} else if nArg := flag.NArg(); nArg > 0 {
		if nArg > 1 {
			n, err := strconv.Atoi(flag.Arg(1))
			if err == nil {
				if err := matchingN(flag.Arg(0), n, os.Stdout, f); err != nil {
					fmt.Fprintf(os.Stderr, "lcd: %v\n", err)
					os.Exit(1)
				}
				return
			}
		}
		if err := matching(flag.Arg(0), os.Stdout, f); err != nil {
			fmt.Fprintf(os.Stderr, "lcd: %v\n", err)
			os.Exit(1)
		}
	}
}

const pathSep = string(os.PathSeparator)

var pathSepB = []byte{os.PathSeparator}

func matching(word string, w io.Writer, r io.Reader) error {
	suffix := []byte(pathSep + strings.TrimSuffix(word, pathSep))
	found := false
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := bytes.TrimSuffix(sc.Bytes(), pathSepB)
		if !bytes.HasSuffix(line, suffix) {
			continue
		}
		s := string(line)
		st, err := os.Stat(s)
		if err != nil {
			continue
		}
		if st.IsDir() {
			found = true
			fmt.Fprintln(w, s)
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("%q: directory not found", word)
	}
	return nil
}

func matchingN(word string, idx int, w io.Writer, r io.Reader) error {
	suffix := []byte(pathSep + strings.TrimSuffix(word, pathSep))
	found := false
	i := 0
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Bytes()
		if !bytes.HasSuffix(line, suffix) {
			continue
		}
		s := string(line)
		st, err := os.Stat(s)
		if err != nil || !st.IsDir() {
			continue
		}
		i++
		if i == idx {
			found = true
			fmt.Fprintln(w, s)
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("%q %d: directory not found", word, idx)
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
