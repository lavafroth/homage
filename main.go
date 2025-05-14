package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/distatus/battery"
)

//go:embed assets templates
var embeddedFiles embed.FS

const serverAddr string = "http://192.168.12.1"

func mustGetBinPath(name string) string {
	cmd, err := exec.LookPath(name)
	if err != nil {
		log.Fatal(err)
	}
	return cmd
}

func batteryCheck() string {
	batteries, err := battery.GetAll()
	if err != nil {
		return string("Battery data unavailable")
	}

	output := ""
	for i, battery := range batteries {
		output += fmt.Sprintf("Battery %d %.02f%% %s at %.02f mW\n", i, battery.Current/battery.Full*100, battery.State.String(), battery.ChargeRate)
	}
	return output
}

func IsLocalIP(r *http.Request) bool {
	return strings.HasPrefix(r.RemoteAddr, "192.168.")
}

func main() {
	assets, err := fs.Sub(embeddedFiles, "assets")
	if err != nil {
		log.Fatal(err)
	}

	templ := template.Must(template.New("").ParseFS(embeddedFiles, "templates/index.tmpl"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templ.ExecuteTemplate(w, "index.tmpl", map[string]any{
			"battery":    batteryCheck(),
			"serverAddr": serverAddr,
		})

	})
	http.HandleFunc("/poweroff", func(w http.ResponseWriter, r *http.Request) {
		if !IsLocalIP(r) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if err := exec.Command(mustGetBinPath("poweroff")).Run(); err != nil {
			log.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/reboot", func(w http.ResponseWriter, r *http.Request) {
		if !IsLocalIP(r) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if err := exec.Command(mustGetBinPath("reboot")).Run(); err != nil {
			log.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
	})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(assets))))

	log.Fatal(http.ListenAndServe(":7047", nil))
}
