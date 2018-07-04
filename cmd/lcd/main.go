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
	"log"
	"os"
	"path/filepath"
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
		log.Fatal(err)
	}
	defer f.Close()

	if *compl != "" {
		if err := complete(*compl, os.Stdout, f); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := matching(flag.Arg(0), os.Stdout, f); err != nil {
			log.Fatal(err)
		}
	}
}

const pathSep = string(os.PathSeparator)

var pathSepB = []byte{os.PathSeparator}

func matching(word string, w io.Writer, r io.Reader) error {
	suffix := []byte(pathSep + strings.TrimSuffix(word, pathSep))
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
			fmt.Fprintln(w, s)
		}
	}
	return sc.Err()
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
