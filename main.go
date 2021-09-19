package main

import (
	"github.com/josexy/gochip8/chip8"
	"os"
)

func main() {
	if len(os.Args) == 1 {
		os.Exit(0)
	}
	emulator := chip8.NewEmulator(os.Args[1])
	emulator.Start()
}
