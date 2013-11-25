/*
Copyright 2013 Google Inc.

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

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %v <host:port>\n", os.Args[0])
		os.Exit(2)
	}
	addr := os.Args[1]

	http.HandleFunc("/save", saveFunc)
	http.Handle("/", http.FileServer(http.Dir("static")))
	log.Fatal(http.ListenAndServe(addr, nil))
}

func saveFunc(w http.ResponseWriter, r *http.Request) {
	mr, err := r.MultipartReader()
	if err != nil {
		logError(w, "error parsing form: %v", err)
		return
	}
	dir, err := mkdir()
	if err != nil {
		logError(w, "error creating directory: %v", err)
		return
	}
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			logError(w, "error reading: %v", err)
			return
		}
		filename := filepath.Join(dir, filepath.FromSlash("/"+path.Clean(p.FileName())))
		err = writeFile(filename, p)
		p.Close()
		if err != nil {
			logError(w, "couldn't write %v: %v", filename, err)
			return
		}
	}
	fmt.Fprintf(w, "all done!")
}

func mkdir() (string, error) {
	dir := time.Now().Format("2006-01-02-15-04")
	if fi, err := os.Stat(dir); err == nil {
		if fi.IsDir() {
			return dir, nil
		}
		return "", fmt.Errorf("path %v exists and is not a directory", dir)
	} else if !os.IsNotExist(err) {
		return "", err
	}
	if err := os.Mkdir(dir, 0777); err != nil {
		return "", err
	}
	return dir, nil
}

func writeFile(name string, r io.Reader) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func logError(w http.ResponseWriter, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Print(msg)
	http.Error(w, msg, http.StatusInternalServerError)
}
