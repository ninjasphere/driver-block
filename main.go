package main

import "github.com/ninjasphere/go-ninja/support"

func main() {

	if _, err := NewDriver(); err != nil {
		log.Errorf("Failed to create driver: %s", err)
		return
	}

	support.WaitUntilSignal()
}
