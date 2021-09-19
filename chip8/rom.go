package chip8

import (
	"io/ioutil"
	"os"
)

type Rom struct {
	data []byte
}

func NewRom(filename string) (rom *Rom, err error) {
	rom = &Rom{}
	err = rom.load(filename)
	return
}

func (r *Rom) Len() int {
	return len(r.data)
}

func (r *Rom) load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	r.data, err = ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	return nil
}
