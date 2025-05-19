package utils

import (
	"fmt"
	"os/exec"
)

func ShowRAMUsage() {
	cmd := exec.Command("free", "-m")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error fetching RAM:", err)
		return
	}
	fmt.Println(string(output))
}
