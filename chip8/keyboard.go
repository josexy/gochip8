package chip8

import "github.com/veandco/go-sdl2/sdl"

// The computers which originally used the Chip-8 Language had a 16-key
// hexadecimal keypad with the following layout:
//
// 1	2	3	C
// 4	5	6	D
// 7	8	9	E
// A	0	B	F
//
// This layout must be mapped into various other configurations to fit the
// keyboards of today's platforms.

const (
	Key0 = iota
	Key1
	Key2
	Key3
	Key4
	Key5
	Key6
	Key7
	Key8
	Key9
	KeyA
	KeyB
	KeyC
	KeyD
	KeyE
	KeyF
)

type KeyMap map[sdl.Keycode]byte

var DefaultKeyMap = KeyMap{
	sdl.K_1: Key1,
	sdl.K_2: Key2,
	sdl.K_3: Key3,
	sdl.K_4: KeyC,
	sdl.K_q: Key4,
	sdl.K_w: Key5,
	sdl.K_e: Key6,
	sdl.K_r: KeyD,
	sdl.K_a: Key7,
	sdl.K_s: Key8,
	sdl.K_d: Key9,
	sdl.K_f: KeyE,
	sdl.K_z: KeyA,
	sdl.K_x: Key0,
	sdl.K_c: KeyB,
	sdl.K_v: KeyF,
}

type Keyboard struct {
	kb []bool
	km KeyMap
}

func NewKeyboard() *Keyboard {
	return &Keyboard{
		kb: make([]bool, 16),
		km: DefaultKeyMap,
	}
}

func (k *Keyboard) GetKey(key sdl.Keycode) (code byte, ok bool) {
	code, ok = k.km[key]
	return
}

func (k *Keyboard) IsPressed(kt byte) bool {
	return k.kb[kt]
}

func (k *Keyboard) IsReleased(kt byte) bool {
	return !k.kb[kt]
}

func (k *Keyboard) PressKey(kt byte) {
	k.kb[kt] = true
}

func (k *Keyboard) ReleaseKey(kt byte) {
	k.kb[kt] = false
}
