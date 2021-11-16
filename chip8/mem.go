package chip8

import "fmt"

const (
	MemorySize  = 2 << 11 // 0x1000 = 4096
	BaseAddress = 2 << 8  // 0x200 = 512
)

//+---------------+= 0xFFF (4095) End of Chip-8 RAM
//|               |
//|               |
//|               |
//|               |
//|               |
//| 0x200 to 0xFFF|
//|     Chip-8    |
//| Program / Data|
//|     Space     |
//|               |
//|               |
//|               |
//+- - - - - - - -+= 0x600 (1536) Start of ETI 660 Chip-8 programs
//|               |
//|               |
//|               |
//+---------------+= 0x200 (512) Start of most Chip-8 programs
//| 0x000 to 0x1FF|
//| Reserved for  |
//|  interpreter  |
//+---------------+= 0x000 (0) Start of Chip-8 RAM

type Memory struct {
	mem []byte
}

func NewMemory() *Memory {
	return &Memory{mem: make([]byte, MemorySize)}
}

func (m *Memory) LoadFrom(data []byte) {
	if len(data) > MemorySize {
		panic("data size more than memory max size")
	}
	// 复制ROM文件数据到 0x200 起始处
	copy(m.mem[BaseAddress:], data[:])
}

func (m *Memory) Read(addr Address) byte {
	m.checkAddress(addr)
	return m.mem[addr]
}

func (m *Memory) ReadOpCode(addr Address) Opcode {
	m.checkAddress(addr)
	high := Opcode(m.mem[addr])
	low := Opcode(m.mem[addr+1])
	return (high << 8) | low
}

func (m *Memory) Write(addr Address, data byte) {
	m.checkAddress(addr)
	m.mem[addr] = data
}

func (m *Memory) checkAddress(addr Address) {
	ok := addr <= 0xFFF
	if !ok {
		panic(fmt.Errorf("invalid memory address: %x", addr))
	}
}
