package main

import (
	"bufio"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/distatus/battery"
)

//go:embed assets
var embeddedFiles embed.FS

//go:embed index.tmpl
var index string

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
	Port     string
	Services []Service
}

type Service struct {
	Name string
	Uri  string
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
	if parsed := net.ParseIP(r.RemoteAddr); parsed != nil {
		return parsed.IsPrivate()
	}
	return false
}

func ReadConfig(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	config := Config{
		Port:     ":80",
		Services: []Service{},
	}
	if err != nil {
		return &config, errors.New("failed to open config.txt")
	}
	reader := bufio.NewScanner(file)
	for reader.Scan() {
		line := reader.Text()
		lastSpace := strings.LastIndex(line, " ")
		before, after := line[:lastSpace], line[lastSpace+1:]

		// First the port
		if before == "port" {
			config.Port = ":" + after
			continue
		}

		// Now the services,
		// sanity check to ensure the second field is a URL
		if _, err := url.Parse(after); err != nil {
			return nil, err
		}

		config.Services = append(config.Services, Service{
			Name: before,
			Uri:  after,
		})
	}
	if err := reader.Err(); err != nil {
		return nil, fmt.Errorf("while reading config: %v", err)
	}
	return &config, nil
}

func main() {
	assets, err := fs.Sub(embeddedFiles, "assets")

	configPath := "config.txt"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	config, err := ReadConfig(configPath)
	if err != nil {
		log.Printf("warn: %q", err)
	}

	templ := template.Must(template.New("").Parse(index))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templ.Execute(w, map[string]any{
			"battery":  batteryCheck(),
			"services": config.Services,
		})

	})
	http.HandleFunc("/api/v1/battery", func(w http.ResponseWriter, r *http.Request) {
		marshal, err := json.Marshal(batteryCheck())
		if err != nil {
			log.Fatal(err)
		}
		w.Write(marshal)
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

	log.Fatal(http.ListenAndServe(config.Port, nil))
}
