package main

import (
  "net/http"
"time"
"path/filepath"
  "fmt"
  "io"
  "os"
  "log"
)

func main() {
  http.HandleFunc("/save", saveFunc)
  http.Handle("/", http.FileServer(http.Dir("static")))
  log.Fatal(http.ListenAndServe(os.Args[1], nil))
}

func saveFunc(w http.ResponseWriter, r *http.Request) {
  mr, err := r.MultipartReader() // 20 MB.
  if err != nil {
    http.Error(w, "error parsing multipart form", 500)
  }
  dir := fmt.Sprintf("%d", time.Now().Unix())
  if err := os.Mkdir(dir, 0777); err != nil {
    http.Error(w, fmt.Sprintf("couldn't write directory %s", dir), 500)
  }
  for {
    p, err := mr.NextPart()
    if err == io.EOF {
      break
    }
    if err != nil {
      http.Error(w, fmt.Sprintf("couldn't read file: %v", err), 500)
    }
    defer p.Close()
    f, err := os.Create(filepath.Join(dir, p.FileName()))
    if err != nil {
      http.Error(w, fmt.Sprintf("couldn't create file: %v", err), 500)
    }
    if _, err := io.Copy(f, p); err != nil {
      http.Error(w, fmt.Sprintf("couldn't write file: %v", err), 500)
    }
  }
  fmt.Fprintf(w, "all done!")
}