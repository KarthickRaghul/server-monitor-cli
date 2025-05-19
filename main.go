package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

// Default log paths for some common servers (Linux + Windows)
var defaultLogPaths = map[string]map[string]string{
	"nginx": {
		"linux":   "/var/log/nginx/access.log",
		"windows": `C:\nginx\logs\access.log`,
	},
	"apache": {
		"linux":   "/var/log/apache2/access.log",
		"windows": `C:\Apache24\logs\access.log`,
	},
}

func main() {
	fmt.Println("Welcome to Server Monitor CLI")
	fmt.Println("-----------------------------")

	// Ask user what they want
	choice := askChoice("Choose option (1- Logs, 2- Real-time Metrics): ", []string{"1", "2"})

	if choice == "1" {
		handleLogs()
	} else {
		handleMetrics()
	}
}

func askChoice(prompt string, valid []string) string {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(prompt)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		for _, v := range valid {
			if input == v {
				return input
			}
		}
		fmt.Println("Invalid choice, please try again.")
	}
}

func handleLogs() {
	server := askServerType()
	logPath := getLogPath(server)

	if logPath == "" {
		logPath = askCustomLogPath()
	}

	fmt.Printf("Tailing logs from: %s\n\n", logPath)
	err := tailLogFile(logPath)
	if err != nil {
		fmt.Println("Error tailing log:", err)
	}
}

func askServerType() string {
	fmt.Println("Select server type or enter 'custom':")
	keys := []string{}
	for k := range defaultLogPaths {
		keys = append(keys, k)
	}
	fmt.Println(strings.Join(keys, ", "), ", custom")

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Server type: ")
		input, _ := reader.ReadString('\n')
		input = strings.ToLower(strings.TrimSpace(input))
		if input == "custom" || defaultLogPaths[input] != nil {
			return input
		}
		fmt.Println("Invalid server type, try again.")
	}
}

func getLogPath(server string) string {
	if server == "custom" {
		return ""
	}
	osType := getOsType()
	if paths, ok := defaultLogPaths[server]; ok {
		if p, ok2 := paths[osType]; ok2 {
			return p
		}
	}
	return ""
}

func getOsType() string {
	if runtime.GOOS == "windows" {
		return "windows"
	}
	return "linux"
}

func askCustomLogPath() string {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter full log file path: ")
		path, _ := reader.ReadString('\n')
		path = strings.TrimSpace(path)
		if _, err := os.Stat(path); err == nil {
			return path
		} else {
			fmt.Println("File does not exist or is not accessible, try again.")
		}
	}
}

func tailLogFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Seek to end of file to tail new logs
	fi, err := file.Stat()
	if err != nil {
		return err
	}
	start := fi.Size()
	_, err = file.Seek(start, 0)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(file)
	fmt.Println("Starting log tail... Press Ctrl+C to stop.")

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			// No new data, wait a bit
			time.Sleep(500 * time.Millisecond)
			continue
		}
		fmt.Print(line)
	}
}

func handleMetrics() {
	fmt.Println("Starting real-time metrics (CPU & RAM). Press Ctrl+C to stop.")
	for {
		printMetrics()
		time.Sleep(2 * time.Second)
	}
}

func printMetrics() {
	// CPU percentage
	percent, err := cpu.Percent(0, false)
	if err != nil {
		fmt.Println("Error getting CPU usage:", err)
		return
	}

	// Memory usage
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println("Error getting memory info:", err)
		return
	}

	fmt.Printf("CPU Usage: %.2f%% | RAM Usage: %.2f%% (%.2f GB free)\n",
		percent[0], vmStat.UsedPercent, float64(vmStat.Free)/1024/1024/1024)
}
