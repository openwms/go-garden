package main

import (
	"fmt"
	"os"
	"time"

	"github.com/openwms/go-garden/types"
)

const (
	API_URL = ""
)

var (
	API_KEY = os.Getenv("THING_SPEAK_API_KEY")
)

// Send the data c to the ThingSpeak API
func sendData(c types.Capture) {

}

func output(d types.Outputs) {

}

func main() {
	fmt.Printf("hello, worldx\n")

	for {
		//		embd.LEDToggle("LED0")
		time.Sleep(250 * time.Millisecond)
		//		sendData(temperature, brightness, moisture, flowRate, sonic)
	}
}
