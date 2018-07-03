// Copyright 2018 Łukasz Pankowski <lukpank at o2 dot pl>. All rights
// reserved.  This source code is licensed under the terms of the MIT
// license. See LICENSE file for details.

/*

# Add the following to your ~/.bashrc:

lcd() {
    declare paths
    paths=$("$HOME/go/bin/lcd" "$@")
    if [ $(echo "${paths}" | wc -l) -eq 1 ]; then
	cd "${paths}"
    else
	echo "${paths}"
    fi
}

*/

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) == 1 {
		return
	}
	f, err := os.Open(filepath.Join(os.Getenv("HOME"), ".lcd", "cache"))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	suffix := []byte("/" + strings.TrimSuffix(os.Args[1], "/"))
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Bytes()
		if bytes.HasSuffix(line, suffix) {
			s := string(line)
			st, err := os.Stat(s)
			if err != nil {
				continue
			}
			if st.IsDir() {
				fmt.Println(s)
			}
		}
	}
	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}
}
