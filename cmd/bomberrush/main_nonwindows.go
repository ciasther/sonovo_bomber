//go:build !windows

package main

import "fmt"

func main() {
	fmt.Println("BOMBER RUSH: aplikacja kiosku jest przeznaczona dla Windows. Zbuduj GOOS=windows GOARCH=amd64.")
}
