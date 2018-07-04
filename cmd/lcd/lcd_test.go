// Copyright 2018 Łukasz Pankowski <lukpank at o2 dot pl>. All rights
// reserved.  This source code is licensed under the terms of the MIT
// license. See LICENSE file for details.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const cache = `
/home/user/go/src/github.com/lukpank/
/home/user/go/src/github.com/lukpank/go-glpk
/home/user/go/src/github.com/lukpank/go-glpk/examples
/home/user/go/src/github.com/lukpank/go-glpk/glpk
/NOT_EXISTS/glpk
/home/user/go/src/github.com/lukpank/jsonlexer
/home/user/go/src/github.com/lukpank/jsondoc
/home/user/go/src/github.com/lukpank/jsondoc/cmd
/home/user/go/src/github.com/lukpank/jsondoc/cmd/jsondoc
/home/user/go/src/github.com/lukpank/jsondoc/example
/home/user/go/src/github.com/lukpank/jsondoc/example/another
/NOT_EXISTS/jsontest
/home/user/go/src/github.com/lukpank/lcd
/home/user/go/src/github.com/lukpank/lcd/cmd
/home/user/go/src/github.com/lukpank/lcd/cmd/lcd
`

func TestMatching(t *testing.T) {
	d := newTestData(t, cache)
	defer d.Close()
	cases := []struct {
		word     string
		expected string
	}{
		{
			"lukpank",
			d.ConvertPath("/home/user/go/src/github.com/lukpank") + "\n",
		},
		{
			"glpk",
			d.ConvertPath("/home/user/go/src/github.com/lukpank/go-glpk/glpk") + "\n",
		},
		{
			"jsonlexer",
			d.ConvertPath("/home/user/go/src/github.com/lukpank/jsonlexer") + "\n",
		},
		{
			"lcd",
			d.ConvertPath("/home/user/go/src/github.com/lukpank/lcd") + "\n" + d.ConvertPath("/home/user/go/src/github.com/lukpank/lcd/cmd/lcd") + "\n",
		},
	}

	for _, c := range cases {
		t.Run(c.word, func(t *testing.T) {
			var b bytes.Buffer
			err := matching(c.word, &b, strings.NewReader(d.Cache))
			if err != nil {
				t.Fatal(err)
			}
			if got := b.String(); c.expected != got {
				t.Errorf("expected %q but got %q", c.expected, got)
			}
		})
	}
}

func TestComplete(t *testing.T) {
	d := newTestData(t, cache)
	defer d.Close()
	cases := []struct {
		word     string
		expected string
	}{
		{
			"lukpank",
			"lukpank\n",
		},
		{
			"l",
			"lukpank\nlcd\n",
		},
		{
			"gl",
			"glpk\n",
		},
		{
			"g",
			"go-glpk\nglpk\n",
		},
		{
			"jso",
			"jsonlexer\njsondoc\n",
		},
		{
			"lc",
			"lcd\n",
		},
	}

	for _, c := range cases {
		t.Run(c.word, func(t *testing.T) {
			var b bytes.Buffer
			err := complete(c.word, &b, strings.NewReader(d.Cache))
			if err != nil {
				log.Fatal(err)
			}
			if got := b.String(); c.expected != got {
				t.Errorf("expected %q but got %q", c.expected, got)
			}
		})
	}
}

type testData struct {
	TempBaseDir string
	Cache       string
}

func newTestData(t *testing.T, cache string) testData {
	tempDir, err := ioutil.TempDir("", "lcd-test")
	if err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	d := testData{TempBaseDir: tempDir}
	for _, path := range strings.Split(cache, "\n") {
		if path == "" {
			b.WriteByte('\n')
			continue
		}
		notExists := strings.HasPrefix(path, "/NOT_EXISTS/")
		path = d.ConvertPath(path)
		fmt.Fprintln(&b, path)
		if notExists {
			continue
		}
		if err := os.MkdirAll(path, 0755); err != nil {
			d.Close()
			t.Fatal(err)
		}
	}
	d.Cache = b.String()
	return d
}

func (d testData) ConvertPath(path string) string {
	return filepath.Join(d.TempBaseDir, strings.Replace(path, "/", pathSep, -1))
}

func (d testData) Close() error {
	return os.RemoveAll(d.TempBaseDir)
}
