package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/openwms/go-garden/types"
	rpio "github.com/stianeikeland/go-rpio"
)

const (
	apiURL              = "https://api.thingspeak.com/update.json"
	thingsSpeakInterval = 40
)

var (
	apiKey = os.Getenv("THINGS_SPEAK_API_KEY")

	trace   *log.Logger
	info    *log.Logger
	warning *log.Logger
	error   *log.Logger

	// Inputs
	fillLevel   = rpio.Pin(12)
	temperature = rpio.Pin(14)
	brightness  = rpio.Pin(40)
	wetness     = rpio.Pin(42)
	flowRate    = rpio.Pin(45)

	// Outputs
	mainValve    = rpio.Pin(9)
	sprinkler    = rpio.Pin(10)
	fillFountain = rpio.Pin(11)
	pump         = rpio.Pin(13)
	ledLights    = rpio.Pin(15)

	// Virtual Outputs
	errorFillingFountain = false
	errorSprinkling      = false
)

func init() {
	fmt.Printf("Initializing...\n")
	initLoggers(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
}

func initLoggers(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

// Send the data c to the ThingSpeak API
func sendData(capture types.Capture) {
	info.Println(">> Send Data", capture)

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	data := url.Values{}
	data.Set("api_key", apiKey)
	data.Add("field1", strconv.Itoa(capture.Input.Temperature))
	data.Add("field2", strconv.Itoa(capture.Input.Moisture))
	data.Add("field3", strconv.Itoa(capture.Input.Sonic))
	data.Add("field4", strconv.Itoa(capture.Input.Brightness))
	req, _ := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	req = req.WithContext(ctx)
	resp, _ := http.DefaultClient.Do(req)
	trace.Println(resp)
}

func readInputs() (d types.Inputs) {
	trace.Println("> Read Inputs")
	res := types.Inputs{Temperature: 1, Brightness: 2, Moisture: 3, FlowRate: 4, Sonic: 5}
	return res
}

func process(inputs types.Inputs) (outputs types.Outputs) {
	trace.Println("  Process Inputs")

	return types.Outputs{false, false, false, false}
}

func writeOutput(output types.Outputs) {
	trace.Println("< Write Outputs", output)
}

func main() {
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Unmap gpio memory when done
	defer rpio.Close()

	// Set pin to output mode
	sprinkler.Output()

	// Toggle pin 20 times
	for x := 0; x < 20; x++ {
		sprinkler.Toggle()
		time.Sleep(time.Second / 5)
	}

	cnt := 1
	for {
		inputs := readInputs()

		outputs := process(inputs)

		writeOutput(outputs)

		if cnt%thingsSpeakInterval == 0 {
			sendData(types.Capture{Input: inputs, Output: outputs})
		}
		cnt++
		time.Sleep(250 * time.Millisecond)
	}
}
