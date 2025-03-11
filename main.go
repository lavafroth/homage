package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed assets templates
var embeddedFiles embed.FS

const serverAddr string = "http://192.168.12.1"

func batteryCheck() string {
	cmds := []string{"/run/current-system/sw/bin/acpi", "acpi"}
	for _, cmd := range cmds {
		acpi := exec.Command(cmd)
		acpiOutput, err := acpi.CombinedOutput()
		if err == nil {
			return string(acpiOutput)
		}
	}
	return string("Battery data unavailable")
}

func IsLocalIP(c *gin.Context) bool {
	return strings.HasPrefix(c.RemoteIP(), "192.168.12.")
}

func main() {
	router := gin.Default()
	templ := template.Must(template.New("").ParseFS(embeddedFiles, "templates/*"))
	router.SetHTMLTemplate(templ)

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"battery":    batteryCheck(),
			"serverAddr": serverAddr,
		})
	})

	router.StaticFS("/public", http.FS(embeddedFiles))
	router.GET("/poweroff", func(c *gin.Context) {
		if !IsLocalIP(c) {
			c.Status(http.StatusForbidden)
			return
		}
		cmd := exec.Command("/run/current-system/sw/sbin/poweroff")
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
		c.Status(http.StatusOK)
	})

	router.GET("/reboot", func(c *gin.Context) {
		if !IsLocalIP(c) {
			c.Status(http.StatusForbidden)
			return
		}
		cmd := exec.Command("/run/current-system/sw/sbin/reboot")
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
		c.Status(http.StatusOK)
	})

	router.Run(":7047")
}
