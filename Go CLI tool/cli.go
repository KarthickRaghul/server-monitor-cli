// SNSMS CLI Tool - Cross-Platform First-Time Setup + Daemon Monitor + Live TCP Server with Preloaded Log Paths
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	OS         string `json:"os"`
	ServerName string `json:"server_name"`
	LogPath    string `json:"log_path"`
}

var configFile = "snsms_config.json"

var linuxLogs = map[string]string{
	"Apache":     "/var/log/apache2/access.log",
	"Nginx":      "/var/log/nginx/access.log",
	"MySQL":      "/var/log/mysql/error.log",
	"PostgreSQL": "/var/log/postgresql/postgresql.log",
	"Docker":     "/var/log/docker.log",
	"SSH":        "/var/log/auth.log",
	"Syslog":     "/var/log/syslog",
	"Kubernetes": "/var/log/kubelet.log",
	"Cron":       "/var/log/cron.log",
	"Systemd":    "/var/log/journal/syslog",
}

var windowsLogs = map[string]string{
	"IIS":            "C:\\inetpub\\logs\\LogFiles\\W3SVC1\\u_exYYMMDD.log",
	"System":         "C:\\Windows\\System32\\winevt\\Logs\\System.evtx",
	"Application":    "C:\\Windows\\System32\\winevt\\Logs\\Application.evtx",
	"Security":       "C:\\Windows\\System32\\winevt\\Logs\\Security.evtx",
	"HTTPERR":        "C:\\Windows\\System32\\LogFiles\\HTTPERR\\httperr1.log",
	"DNS":            "C:\\Windows\\System32\\Dns\\dns.log",
	"DHCP":           "C:\\Windows\\System32\\Dhcp\\DhcpSrvLog.txt",
	"Print":          "C:\\Windows\\System32\\spool\\PRINTERS",
	"Firewall":       "C:\\Windows\\System32\\LogFiles\\Firewall\\pfirewall.log",
	"WindowsUpdate":  "C:\\Windows\\WindowsUpdate.log",
}

func firstTimeSetup() Config {
	reader := bufio.NewReader(os.Stdin)
	osType := runtime.GOOS
	fmt.Println("🔧 First-time setup:")
	fmt.Println("Detected OS:", osType)

	fmt.Print("Enter Server Name: ")
	serverName, _ := reader.ReadString('\n')
	serverName = strings.TrimSpace(serverName)

	// Show predefined log path options
	fmt.Println("Available Log File Paths:")
	var options []string
	var selectedPath string
	index := 1
	if osType == "windows" {
		for name, path := range windowsLogs {
			fmt.Printf("%d) %s - %s\n", index, name, path)
			options = append(options, path)
			index++
		}
	} else {
		for name, path := range linuxLogs {
			fmt.Printf("%d) %s - %s\n", index, name, path)
			options = append(options, path)
			index++
		}
	}

	fmt.Print("Choose a number for a predefined log path, or press Enter to enter a custom path: ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if choice == "" {
		fmt.Print("Enter custom Log File Path: ")
		selectedPath, _ = reader.ReadString('\n')
		selectedPath = strings.TrimSpace(selectedPath)
	} else {
		numChoice, err := strconv.Atoi(choice)
		if err == nil && numChoice >= 1 && numChoice <= len(options) {
			selectedPath = options[numChoice-1]
		} else {
			fmt.Println("❌ Invalid choice. Using default.")
			if osType == "windows" {
				selectedPath = windowsLogs["HTTPERR"]
			} else {
				selectedPath = linuxLogs["Syslog"]
			}
		}
	}

	config := Config{OS: osType, ServerName: serverName, LogPath: selectedPath}
	data, _ := json.MarshalIndent(config, "", "  ")
	ioutil.WriteFile(configFile, data, 0644)
	fmt.Println("✅ Configuration saved to", configFile)
	return config
}

func loadConfig() (Config, bool) {
	var config Config
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return config, false
	}
	json.Unmarshal(data, &config)
	return config, true
}

func monitor(config Config) {
	fmt.Println("📡 Starting monitoring for:", config.ServerName)
	fmt.Println("📝 Log path:", config.LogPath)

	for {
		fmt.Println("\n---- System Status @", time.Now().Format(time.RFC1123), "----")

		if config.OS == "windows" {
			execCommand("wmic cpu get loadpercentage")
			execCommand("wmic os get FreePhysicalMemory,TotalVisibleMemorySize")
			execCommand("wmic logicaldisk get size,freespace,caption")
		} else {
			execCommand("top -b -n1 | head -n 5")
			execCommand("free -h")
			execCommand("df -h")
		}

		fmt.Println("\n🔍 Last 10 log entries:")
		if config.OS == "windows" {
			execCommand("powershell -Command \"Get-Content '" + config.LogPath + "' -Tail 10\"")
		} else {
			execCommand("tail -n 10 " + config.LogPath)
		}

		time.Sleep(10 * time.Second)
	}
}

func execCommand(command string) {
	cmd := exec.Command("bash", "-c", command)
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println(string(output))
}

func runAndCapture(command string) string {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "Error: " + err.Error() + "\n"
	}
	return string(output) + "\n"
}

func startLiveServer(config Config) {
	ln, err := net.Listen("tcp", ":9090")
	if err != nil {
		fmt.Println("Failed to start server:", err)
		return
	}
	fmt.Println("🌐 Listening for live requests on port 9090")

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn, config)
	}
}

func handleConnection(conn net.Conn, config Config) {
	defer conn.Close()
	var result strings.Builder

	result.WriteString("---- LIVE SYSTEM STATS ----\n")
	if config.OS == "windows" {
		result.WriteString(runAndCapture("wmic cpu get loadpercentage"))
		result.WriteString(runAndCapture("wmic os get FreePhysicalMemory,TotalVisibleMemorySize"))
		result.WriteString(runAndCapture("wmic logicaldisk get size,freespace,caption"))
		result.WriteString(runAndCapture("powershell -Command \"Get-Content '" + config.LogPath + "' -Tail 10\""))
	} else {
		result.WriteString(runAndCapture("top -b -n1 | head -n 5"))
		result.WriteString(runAndCapture("free -h"))
		result.WriteString(runAndCapture("df -h"))
		result.WriteString(runAndCapture("tail -n 10 " + config.LogPath))
	}

	conn.Write([]byte(result.String()))
}

func main() {
	var config Config
	existsConfig, ok := loadConfig()
	if !ok {
		config = firstTimeSetup()
	} else {
		config = existsConfig
	}

	go monitor(config)
	startLiveServer(config)
}