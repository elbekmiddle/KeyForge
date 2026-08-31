package main

import (
	"fmt"
	"log"

	"github.com/elbekmiddle/KeyForge/internal/linux"
)

func main() {
	fmt.Println("⚒️ KeyForge")
	fmt.Println("Scanning input devices...")
	fmt.Println()

	devices, err := linux.ListInputDevices()
	if err != nil {
		log.Fatal(err)
	}

	if len(devices) == 0 {
		fmt.Println("No input devices found.")
		return
	}

	for i, d := range devices {
		fmt.Printf("[%d] %s\n", i+1, d.Name)
		fmt.Printf("    Type: %s\n", d.Type)
		fmt.Printf("    Path: %s\n", d.Path)
		fmt.Println()
	}
}
