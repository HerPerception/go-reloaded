package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"go-reloaded/processor"
)

type PageData struct {
	InputText  string
	OutputText string
	FileName   string
	IsFile     bool
}

func main() {
	http.HandleFunc("/", handleDashboard)
	http.HandleFunc("/download", handleDownload)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Production Processing Server running on port %s", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	tmplPath := filepath.Join("templates", "index.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Internal Server Error: Template missing.", http.StatusInternalServerError)
		return
	}

	data := PageData{}

	if r.Method == http.MethodPost {
		// 1. Check for physical file streams
		file, header, err := r.FormFile("textFile")
		if err == nil {
			defer file.Close()
			fileBytes, err := io.ReadAll(file)
			if err == nil {
				data.InputText = string(fileBytes)
				data.FileName = header.Filename
				data.IsFile = true
			}
		} else {
			// 2. Fall back to raw text field copy-paste inputs
			data.InputText = r.FormValue("rawText")
			data.FileName = "processed_result.txt"
			data.IsFile = false
		}

		// Execute memory-optimized parsing pipeline
		data.OutputText = processor.ProcessText(data.InputText)
	}

	tmpl.Execute(w, data)
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	content := r.FormValue("downloadContent")
	filename := r.FormValue("downloadName")
	if filename == "" {
		filename = "result.txt"
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(content))
}