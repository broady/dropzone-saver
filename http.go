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
		defer p.Close()
		filename := filepath.Join(dir, filepath.FromSlash("/"+path.Clean(p.FileName())))
		if err := writeFile(filename, p); err != nil {
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
