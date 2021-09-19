package chip8

import (
	"fmt"
	"math/rand"
	"time"
)

type Address uint16
type Opcode uint16

type external struct {
	screen   *Screen   // 屏幕
	keyboard *Keyboard // 键盘
	mem      *Memory   // 内存
}

type CPU struct {
	external
	PC         Address   // 程序计数器
	I          Address   // 地址寄存器
	SP         byte      // 栈指针寄存器
	STACK      []Address // 用于存储 subroutine 返回地址
	V          []byte    // 通用寄存器，其中 V[0xF]是进位标志
	delayTimer int
	soundTimer int
	halted     bool
	haltedReg  int  // 指定哪一个键需要等待按下
	drawAction bool // 指示 Screen 是否需要重绘
}

func NewCPU() *CPU {
	rand.Seed(time.Now().UnixNano())
	cpu := &CPU{}
	cpu.reset()
	cpu.mem = NewMemory()
	return cpu
}

func (c *CPU) ConnectScreen(screen *Screen) {
	c.screen = screen
}

func (c *CPU) ConnectKeyboard(keyboard *Keyboard) {
	c.keyboard = keyboard
}

func (c *CPU) reset() {
	c.PC = BaseAddress
	c.I = 0
	c.SP = 0
	c.STACK = make([]Address, 16)
	c.V = make([]byte, 16)
	c.delayTimer = 0
	c.soundTimer = 0
	c.halted = false
	c.haltedReg = 0
	c.drawAction = true
}

func (c *CPU) pushPC() {
	c.STACK[c.SP] = c.PC
	c.SP++
}

func (c *CPU) popPC() {
	c.SP--
	c.PC = c.STACK[c.SP]
}

func (c *CPU) InitMemory(data []byte) {
	// 首先复制 font set 到内存 0x000-0x04F
	copy(c.mem.mem, chip8Fontset[:])
	// 然后加载ROM文件到 0x200 基地址处
	c.mem.LoadFrom(data)
}

func (c *CPU) DumpMemoryArea() {
	var i int
	for i < MemorySize {
		fmt.Printf("=> [0x%03X]: ", i)
		for j := 0; j < 16; j++ {
			fmt.Printf("%02X ", c.mem.mem[i])
			i++
		}
		fmt.Println()
	}
	fmt.Println()
}

func (c *CPU) _NNN(opc Opcode) Address {
	return Address(opc) & 0x0FFF
}

func (c *CPU) _NN(opc Opcode) byte {
	return byte(opc & 0x00FF)
}

func (c *CPU) _N(opc Opcode) byte {
	return byte(opc & 0x000F)
}

func (c *CPU) _X(opc Opcode) int {
	return int(opc&0x0F00) >> 8
}

func (c *CPU) _Y(opc Opcode) int {
	return int(opc&0x00F0) >> 4
}

func (c *CPU) next() {
	c.PC += 2
}

/*
Draws a sprite at coordinate (VX, VY) that has a width of 8 pixels and a height of N pixels.
Each row of 8 pixels is read as bit-coded starting from memory location I;
I value does not change after the execution of this instruction.

As described above,
VF is set to 1 if any screen pixels are flipped from set to unset when the sprite is drawn,
and to 0 if that does not happen
*/

/*
// [vx, vy, vx+8, vy+n]
// in memory "0" meaning: 0xF0 0x90 0x90 0x90 0xF0
"0"	  Binary  Hex
**** 11110000 0xF0
*  * 10010000 0x90
*  * 10010000 0x90
*  * 10010000 0x90
**** 11110000 0xF0
*/
func (c *CPU) draw(vx, vy, n int) {
	fmt.Printf("-> vx:%d, vy:%d, n:%d\n", vx, vy, n)
	c.V[0xF] = 0
	for y := 0; y < n; y++ {
		// 从内存中取出一个像素pixel
		pixel := c.mem.Read(c.I + Address(y))
		// 并计算该像素的8位中哪一位需要翻转
		for x := 0; x < 8; x++ {
			// 该位是否需要翻转，翻转的目的是产生其他的sprite
			if pixel&(1<<(7-x)) > 0 {
				nx := (vx + x) % WIDTH
				ny := (vy + y) % HEIGHT
				if c.screen.pixel(nx, ny) {
					c.V[0xF] = 1
				}
				c.screen.flip(nx, ny, true)
			}
		}
	}
}

func (c *CPU) Step() {
	if c.halted {
		var isPressed bool
		var keycode int
		// 检测是否有按键按下
		for i := 0; i < 16; i++ {
			if c.keyboard.IsPressed(byte(i)) {
				isPressed = true
				keycode = i
			}
		}
		if isPressed {
			// 将该按键的keycode 存储到 V 寄存器中
			c.halted = false
			c.V[c.haltedReg] = byte(keycode)
		}
		// 阻塞所有指令执行操作，直到下一个键盘事件发生
		return
	}

	// 读取操作码
	opcode := c.mem.ReadOpCode(c.PC)
	fmt.Printf("-> opcode:[0x%04X] PC:[0x%03X], SP:[0x%02X], I:[0x%03X]\n", opcode, c.PC, c.SP, c.I)

	// 移动PC
	c.next()
	switch opcode & 0xF000 {
	case 0x0000:
		switch opcode & 0x00FF {
		case 0x00E0: // Clears the screen
			c.screen.clear()
			c.drawAction = true
		case 0x00EE: // Returns from a subroutine
			c.popPC()
		}
	case 0x1000: // Jumps to address NNN
		c.PC = c._NNN(opcode)
	case 0x2000: // Calls subroutine at NNN
		c.pushPC()
		c.PC = c._NNN(opcode)
	case 0x3000: // Skips the next instruction if VX equals NN
		if c.V[c._X(opcode)] == c._NN(opcode) {
			c.next()
		}
	case 0x4000: // Skips the next instruction if VX does not equal NN
		if c.V[c._X(opcode)] != c._NN(opcode) {
			c.next()
		}
	case 0x5000: // Skips the next instruction if VX equals VY
		if c.V[c._X(opcode)] == c.V[c._Y(opcode)] {
			c.next()
		}
	case 0x6000: // Sets VX to NN
		c.V[c._X(opcode)] = c._NN(opcode)
	case 0x7000: // Adds NN to VX. (Carry flag is not changed)
		c.V[c._X(opcode)] += c._NN(opcode)
	case 0x8000:
		switch opcode & 0x000F {
		case 0x0000: // Sets VX to the value of VY
			c.V[c._X(opcode)] = c.V[c._Y(opcode)]
		case 0x0001: // Sets VX to VX or VY
			c.V[c._X(opcode)] |= c.V[c._Y(opcode)]
		case 0x0002: // Sets VX to VX and VY
			c.V[c._X(opcode)] &= c.V[c._Y(opcode)]
		case 0x0003: // Sets VX to VX xor VY
			c.V[c._X(opcode)] ^= c.V[c._Y(opcode)]
		case 0x0004: // Adds VY to VX. VF is set to 1 when there's a carry, and to 0 when there is not
			x := c._X(opcode)
			y := c._Y(opcode)
			r := uint16(c.V[x]) + uint16(c.V[y])
			// 结果是否产生进位
			if r&0xFF00 > 0 {
				c.V[0xF] = 1
			} else {
				c.V[0xF] = 0
			}
			c.V[x] += c.V[y]
		case 0x0005: // VY is subtracted from VX. VF is set to 0 when there's a borrow, and 1 when there is not
			x := c._X(opcode)
			y := c._Y(opcode)
			if c.V[y] > c.V[x] {
				c.V[0xF] = 0
			} else {
				c.V[0xF] = 1
			}
			c.V[x] -= c.V[y]
		case 0x0006: // Stores the least significant bit of VX in VF and then shifts VX to the right by 1
			x := c._X(opcode)
			c.V[0xF] = c.V[x] & 0x1
			c.V[x] >>= 1
		case 0x0007: // Sets VX to VY minus VX. VF is set to 0 when there's a borrow, and 1 when there is not
			x := c._X(opcode)
			y := c._Y(opcode)
			// 是否借位
			if c.V[x] > c.V[y] {
				c.V[0xF] = 0
			} else {
				c.V[0xF] = 1
			}
			c.V[x] = c.V[y] - c.V[x]
		case 0x000E: // Stores the most significant bit of VX in VF and then shifts VX to the left by 1
			x := c._X(opcode)
			c.V[0xF] = c.V[x] >> 7
			c.V[x] <<= 1
		}
	case 0x9000: // Skips the next instruction if VX does not equal VY
		if c.V[c._X(opcode)] != c.V[c._Y(opcode)] {
			c.next()
		}
	case 0xA000: // Sets I to the address NNN
		c.I = c._NNN(opcode)
	case 0xB000: // Jumps to the address NNN plus V0
		c.PC = Address(c.V[0]) + c._NNN(opcode)
	case 0xC000: // Sets VX to the result of a bitwise and operation on a random number (Typically: 0 to 255) and NN
		c.V[c._X(opcode)] = byte(rand.Int()%256) & c._NN(opcode)
	case 0xD000:
		c.draw(int(c.V[c._X(opcode)]), int(c.V[c._Y(opcode)]), int(c._N(opcode)))
		c.drawAction = true
	case 0xE000:
		switch opcode & 0x00FF {
		case 0x009E: // Skips the next instruction if the key stored in VX is pressed
			if c.keyboard.IsPressed(c.V[c._X(opcode)]) {
				c.next()
			}
		case 0x00A1: // Skips the next instruction if the key stored in VX is not pressed
			if c.keyboard.IsReleased(c.V[c._X(opcode)]) {
				c.next()
			}
		}
	case 0xF000:
		switch opcode & 0x00FF {
		case 0x0007: // Sets VX to the value of the delay timer
			c.V[c._X(opcode)] = byte(c.delayTimer)
		case 0x000A: // A key press is awaited, and then stored in VX
			c.halted = true
			// haltedReg 保存了需要等待的那个键
			c.haltedReg = int(c.V[c._X(opcode)])
		case 0x0015: // Sets the delay timer to VX
			c.delayTimer = int(c.V[c._X(opcode)])
		case 0x0018: // Sets the sound timer to VX
			c.soundTimer = int(c.V[c._X(opcode)])
		case 0x001E: // Adds VX to I. VF is not affected
			c.I += Address(c.V[c._X(opcode)])
		case 0x0029: // Sets I to the location of the sprite for the character in VX.
			// Characters 0-F (in hexadecimal) are represented by a 4x5 font
			// 每个字体占用5个字节且位于内存区域 0x00-0x4F 中
			// 这里 VX 保存的是字体索引(0-F)
			c.I = Address(c.V[c._X(opcode)]) * 0x5
		case 0x0033: // BCD
			vx := c.V[c._X(opcode)]
			c.mem.Write(c.I+0, vx/100)
			c.mem.Write(c.I+1, byte(vx/10)%10)
			c.mem.Write(c.I+2, vx%10)
		case 0x0055: // Stores V0 to VX (including VX) in memory starting at address I
			vx := c._X(opcode)
			for i := 0; i <= vx; i++ {
				c.mem.Write(c.I+Address(i), c.V[i])
			}
		case 0x0065: // Fills V0 to VX (including VX) with values from memory starting at address I
			vx := c._X(opcode)
			for i := 0; i <= vx; i++ {
				c.V[i] = c.mem.Read(c.I + Address(i))
			}
		}
	}
	if c.delayTimer > 0 {
		c.delayTimer--
	}
	if c.soundTimer > 0 {
		if c.soundTimer == 1 {
			// beep
		}
		c.soundTimer--
	}
}
