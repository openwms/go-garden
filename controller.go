package main

import (
	"context"
	"encoding/json"
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
	apiReadURL          = "https://api.thingspeak.com/channels/507346/feeds.json"
	thingsSpeakInterval = 40
)

var (
	apiKey = os.Getenv("THINGS_SPEAK_API_KEY")

	trace   *log.Logger
	info    *log.Logger
	warning *log.Logger
	error   *log.Logger

	// Inputs
	fillLevel   = rpio.Pin(12) //pwd
	temperature = rpio.Pin(13) //pwd
	brightness  = rpio.Pin(40) //pwd
	wetness     = rpio.Pin(41) //pwd
	flowRate    = rpio.Pin(45) //pwd

	// Virtual Inputs
	pumpOn      = false
	mainValveOn = false
	sprinklerOn = false

	// Outputs
	mainValve    = rpio.Pin(4)
	sprinkler    = rpio.Pin(5)
	fillFountain = rpio.Pin(6)
	pump         = rpio.Pin(7)
	ledLights    = rpio.Pin(8)

	// Virtual Outputs
	errorFillingFountain = false
	errorSprinkling      = false
)

func init() {
	fmt.Printf("Initializing...\n")
	initLoggers(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	initGpio()
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

func initGpio() {
	fillLevel.Mode(rpio.Pwm)
	fillLevel.Freq(64000)
	fillLevel.DutyCycle(0, 32)

	temperature.Mode(rpio.Pwm)
	temperature.Freq(64000)
	temperature.DutyCycle(0, 32)

	brightness.Mode(rpio.Pwm)
	brightness.Freq(64000)
	brightness.DutyCycle(0, 32)

	wetness.Mode(rpio.Pwm)
	wetness.Freq(64000)
	wetness.DutyCycle(0, 32)

	flowRate.Mode(rpio.Pwm)
	flowRate.Freq(64000)
	flowRate.DutyCycle(0, 32)
}

// Send the data c to the ThingSpeak API
func sendData(capture types.Capture) {
	info.Println(">> Send Data", capture)

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	data := url.Values{}
	data.Set("api_key", apiKey)
	data.Add("field1", strconv.Itoa(capture.Input.Temperature))
	data.Add("field2", strconv.Itoa(capture.Input.Wetness))
	data.Add("field3", strconv.Itoa(capture.Input.FillLevel))
	data.Add("field4", strconv.Itoa(capture.Input.Brightness))
	req, _ := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	req = req.WithContext(ctx)
	resp, _ := http.DefaultClient.Do(req)
	trace.Println(resp)
}

func readVirtualInputs() (pumpOn bool, sprinklerOn bool) {
	info.Println(">> Read Virtual Inputs")

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	req, _ := http.NewRequest("GET", apiReadURL, nil)
	req = req.WithContext(ctx)
	q := req.URL.Query()
	q.Add("api_key", apiKey)
	q.Add("results", "1")
	req.URL.RawQuery = q.Encode()

	resp, _ := http.DefaultClient.Do(req)
	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	ts := types.ThingSpeakQuery{}
	jsonErr := json.Unmarshal(body, &ts)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	trace.Println(ts)
	f1 := false
	f2 := false
	f1, _ = strconv.ParseBool(ts.Feeds[0].Field1)
	f2, _ = strconv.ParseBool(ts.Feeds[0].Field2)
	return f1, f2
}

func readInputs() (d types.Inputs) {
	trace.Println("> Read Inputs")
	pumpOn, sprinklerOn = readVirtualInputs()
	res := types.Inputs{Temperature: int(temperature.Read()), Brightness: int(brightness), Wetness: int(wetness), FlowRate: int(flowRate), FillLevel: int(fillLevel), PumpOn: pumpOn, SprinklerOn: sprinklerOn}
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

	cnt := 1
	for {
		inputs := readInputs()

		outputs := process(inputs)

		writeOutput(outputs)

		sprinkler.Toggle()

		if cnt%thingsSpeakInterval == 0 {
			sendData(types.Capture{Input: inputs, Output: outputs})
		}
		cnt++
		time.Sleep(250 * time.Millisecond)
	}
}
