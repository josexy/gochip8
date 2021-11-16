package chip8

import (
	"github.com/veandco/go-sdl2/sdl"
)

const (
	WIDTH        = 64
	HEIGHT       = 32
	ScreenWidth  = 640
	ScreenHeight = 320
	ScreenTitle  = "Simple Chip8 Emulator"
)

var (
	// RGB
	BackgroundColor = [3]uint8{0, 0, 0}
	ForegroundColor = [3]uint8{0, 255, 255}
)

var chip8Fontset = [80]byte{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

type Screen struct {
	pixels   []byte
	window   *sdl.Window
	renderer *sdl.Renderer
}

func NewScreen(memory *Memory) *Screen {
	s := &Screen{
		// 显示屏数据存储在内存 [0xF00-0xFFF] 区域，即 256 字节
		pixels: memory.mem[0xF00:],
	}
	s.InitScreen()
	return s
}

func (s *Screen) InitScreen() {
	var err error
	if err = sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

	s.window, err = sdl.CreateWindow(ScreenTitle,
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		ScreenWidth, ScreenHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}

	s.renderer, err = sdl.CreateRenderer(s.window, -1, 0)
	if err != nil {
		panic(err)
	}
}

func (s *Screen) destroy() {
	_ = s.renderer.Destroy()
	_ = s.window.Destroy()
}

func (s *Screen) refresh() {
	_ = s.renderer.SetDrawColor(BackgroundColor[0], BackgroundColor[1], BackgroundColor[2], 255)
	_ = s.renderer.Clear()
}

func (s *Screen) clear() {
	s.pixels = make([]byte, HEIGHT*WIDTH)
	s.refresh()
}

func (s *Screen) update() {
	wSize := ScreenWidth / WIDTH
	hSize := ScreenHeight / HEIGHT
	s.refresh()

	for y := 0; y < HEIGHT; y++ {
		for x := 0; x < WIDTH; x++ {
			rect := sdl.Rect{
				X: int32(x * wSize),
				Y: int32(y * hSize),
				W: int32(wSize),
				H: int32(hSize),
			}
			if s.pixel(x, y) {
				// 前景色
				_ = s.renderer.SetDrawColor(ForegroundColor[0], ForegroundColor[1], ForegroundColor[2], 255)
			} else {
				// 背景色
				_ = s.renderer.SetDrawColor(BackgroundColor[0], BackgroundColor[1], BackgroundColor[2], 255)
			}
			_ = s.renderer.FillRect(&rect)
		}
	}

	// 重新更新、渲染
	s.renderer.Present()
}

func (s *Screen) pixel(x, y int) bool {
	addr := y*WIDTH + x
	// 判断一个字节中对应的某位是是1还是0
	return s.pixels[addr/8]&(1<<(addr%8)) != 0
}

func (s *Screen) flip(x, y int) {
	addr := y*WIDTH + x
	// 将某个字节中某位反转
	s.pixels[addr/8] ^= 1 << (addr % 8)
}
