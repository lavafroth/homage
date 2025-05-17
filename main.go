package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/distatus/battery"
)

//go:embed assets templates
var embeddedFiles embed.FS

func mustGetBinPath(name string) string {
	cmd, err := exec.LookPath(name)
	if err != nil {
		log.Fatal(err)
	}
	return cmd
}

type BatteryStats struct {
	Id         int
	Percentage float32
	State      string
	Rate       float32
}

type Config struct {
	Host string
	Port uint16
}

func batteryCheck() []BatteryStats {
	stats := []BatteryStats{}
	batteries, err := battery.GetAll()
	if err != nil {
		return stats
	}

	for i, battery := range batteries {
		stat := BatteryStats{
			Id:         i,
			Percentage: float32(battery.Current / battery.Full * 100),
			State:      battery.State.String(),
			Rate:       float32(battery.ChargeRate),
		}
		stats = append(stats, stat)
	}
	return stats
}

func IsLocalIP(r *http.Request) bool {
	return strings.HasPrefix(r.RemoteAddr, "192.168.")
}

func ReadConfig() (*Config, error) {
	configRaw, err := os.ReadFile("config.json")
	defaultConfig := Config{
		Host: "192.168.12.1",
		Port: 80,
	}
	if err != nil {
		return &defaultConfig, errors.New("failed to read config.json")
	}

	var config Config
	if err = json.Unmarshal(configRaw, &config); err != nil {
		return &defaultConfig, errors.New("failed to parse config.json")
	}
	return &config, nil
}

func main() {
	assets, err := fs.Sub(embeddedFiles, "assets")

	config, err := ReadConfig()
	if err != nil {
		log.Printf("warn: %q", err)
	}

	templ := template.Must(template.New("").ParseFS(embeddedFiles, "templates/index.tmpl"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templ.ExecuteTemplate(w, "index.tmpl", map[string]any{
			"battery":    batteryCheck(),
			"serverAddr": config.Host,
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

	portString := fmt.Sprintf(":%d", config.Port)
	log.Fatal(http.ListenAndServe(portString, nil))
}
