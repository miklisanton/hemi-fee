package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

func main() {
    if len(os.Args) != 2 {
        log.Fatalf("Usage: %s <path-to-service-config>", os.Args[0])
    }
    path := os.Args[1]
    log.Printf("Path to service config: %s", path)
    splitted := strings.Split(path, "/")
    name := splitted[len(splitted)-1]
    log.Printf("Service name: %s", name)


    for {
        // Get current fee
        fee := fetchFee()
        log.Printf("Current fee: %s", fee)
        // Replace fee in config file
        replaceEnvVar("POPM_STATIC_FEE", fee, path)        
        // Restart daemon
        _, err := exec.Command("systemctl", "daemon-reload").Output()
        if err != nil {
            log.Fatalf("Error restarting daemon: %s", err)
        }
        // Restart service
        _, err = exec.Command("systemctl", "restart", name).Output()
        if err != nil {
            log.Fatalf("Error restarting service: %s", err)
        }
        time.Sleep(30 * time.Minute)
    }

}


func replaceEnvVar(name string, value string, path string) {
    // Read file    
    file, err := os.ReadFile(path)
    if err != nil {
        log.Fatalf("Error reading file: %s", err)
    }
    content := string(file)
    // Replace env var
    envVarPattern := fmt.Sprintf(`Environment="%s=.*"`, name)
    newEnvVar := fmt.Sprintf(`Environment="%s=%s"`, name, value)
    re := regexp.MustCompile(envVarPattern)
    if re.MatchString(content) {
        newContent := re.ReplaceAllString(content, newEnvVar)
        // Write file
        err = os.WriteFile(path, []byte(newContent), 0644)
        if err != nil {
            log.Fatalf("Error writing file: %s", err)
        }

        log.Printf("Replaced env var %s with value %s", name, value)
    } else {
        log.Fatalf("Env var %s not found in file", name)
    }

}

func fetchFee() string {
    resp, err := http.Get("https://blockstream.info/testnet/api/fee-estimates")
    if err != nil {
        log.Fatalf("Error fetching fee: %s", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        log.Fatalf("Error fetching fee: %s", resp.Status)
    }
    
    var stats map[string]float64

    err = json.NewDecoder(resp.Body).Decode(&stats)
    if err != nil {
        log.Fatalf("Error decoding response: %s", err)
    }

    avg := int((stats["1"] + stats["25"]) / 2)
    return fmt.Sprintf("%d", avg)
}




