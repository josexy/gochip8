package chip8

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
)

type Emulator struct {
	Rom      *Rom
	CPU      *CPU
	Keyboard *Keyboard
	Screen   *Screen
	running  bool
}

func NewEmulator(filename string) *Emulator {
	e := &Emulator{
		CPU:      NewCPU(),
		Screen:   NewScreen(),
		Keyboard: NewKeyboard(),
		running:  true,
	}
	var err error
	e.Rom, err = NewRom(filename)
	if err != nil {
		panic(fmt.Errorf("create chip8 emulator failed: %s", err.Error()))
	}
	e.CPU.InitMemory(e.Rom.data)
	e.CPU.ConnectKeyboard(e.Keyboard)
	e.CPU.ConnectScreen(e.Screen)
	return e
}

func (e *Emulator) Start() {
	e.start()
}

func (e *Emulator) start() {
	defer e.Screen.destroy()

	for e.running {

		e.clock()

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				e.running = false
				break
			case *sdl.KeyboardEvent:
				ke := event.(*sdl.KeyboardEvent)
				if ke.Keysym.Sym == sdl.K_ESCAPE {
					e.running = false
					break
				} else if code, ok := e.Keyboard.GetKey(ke.Keysym.Sym); ok {
					if ke.State == sdl.PRESSED {
						e.Keyboard.PressKey(code)
					} else if ke.State == sdl.RELEASED {
						e.Keyboard.ReleaseKey(code)
					}
				}
			}
		}
		//
		sdl.Delay(1000 / 60)
	}
}

func (e *Emulator) clock() {
	e.CPU.Step()

	// 重绘
	if e.CPU.drawAction {
		e.Screen.update()
		e.CPU.drawAction = false
	}
}
