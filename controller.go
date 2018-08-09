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
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/openwms/go-garden/types"
	rpio "github.com/stianeikeland/go-rpio"
)

const (
	apiURL                  = "https://api.thingspeak.com/update.json"
	apiReadURL              = "https://api.thingspeak.com/channels/507346/feeds.json"
	thingsSpeakInterval     = 120
	maximumWateringInterval = 300
)

var (
	apiKey = os.Getenv("THINGS_SPEAK_API_KEY")

	trace   *log.Logger
	info    *log.Logger
	warning *log.Logger
	error   *log.Logger

	// Inputs
	fillLevel     = rpio.Pin(18) //pwd
	fillLevelEcho = rpio.Pin(25)
	temperature   = rpio.Pin(4)  //pwd
	brightness    = rpio.Pin(40) //pwd
	wetness       = rpio.Pin(10) //pwd
	flowRate      = rpio.Pin(45) //pwd

	// Virtual Inputs
	pumpOn            = false
	mainValveOn       = false
	sprinklerOn       = false
	fillFontaineValve = false

	// Outputs
	mainValve    = rpio.Pin(17) //IN1 (Relais 1)
	sprinkler    = rpio.Pin(27) //IN2
	fillFountain = rpio.Pin(22) //IN3
	pump         = rpio.Pin(23) //IN4
	ledLights    = rpio.Pin(24) //IN5

	// Virtual Outputs
	errorFillingFountain = false
	errorSprinkling      = false
)

func init() {
	fmt.Printf("Initializing...\n")
	//initLoggers(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
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

func boolToa(in bool) string {
	if in {
		return "1"
	}
	return "0"
}

func boolToStr(in bool) string {
	if in {
		return "ON"
	}
	return "OFF"
}

// Send the data c to the ThingSpeak API
func sendData(capture types.Capture) {
	trace.Println(">> Send Data", capture)

	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	data := url.Values{}
	data.Set("api_key", apiKey)
	data.Add("field1", strconv.FormatFloat(capture.Input.Temperature, 'f', 4, 64))
	data.Add("field2", strconv.Itoa(capture.Input.Wetness))
	data.Add("field3", strconv.Itoa(capture.Input.FillLevel))
	data.Add("field4", strconv.Itoa(capture.Input.Brightness))
	data.Add("field6", boolToa(capture.Input.FillFontaineValve))
	data.Add("field7", boolToa(capture.Input.PumpOn))
	data.Add("field8", boolToa(capture.Input.SprinklerOn))
	req, _ := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	req = req.WithContext(ctx)
	trace.Println(">> Send Request", req)
	resp, _ := http.DefaultClient.Do(req)
	trace.Println(resp)
}

func readVirtualInputs(currentOutput types.Outputs) (pumpOn bool, sprinklerOn bool, fillFontaine bool) {
	trace.Println(">> Read Virtual Inputs")
	f1 := currentOutput.Fontaine
	f2 := false
	f3 := false

	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	req, _ := http.NewRequest("GET", apiReadURL, nil)
	req = req.WithContext(ctx)
	q := req.URL.Query()
	q.Add("api_key", apiKey)
	q.Add("results", "1")
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		error.Println("Error performing GET: ", err)
		return f1, f2, f3
	}
	defer resp.Body.Close()
	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		error.Println("Error ready response body GET: ", readErr)
		return f1, f2, f3
	}

	var ts = new(types.ThingSpeakQuery)
	jsonErr := json.Unmarshal(body, &ts)
	if jsonErr != nil {
		error.Println("Error unmarshalling json response in GET: ", jsonErr)
		return f1, f2, f3
	}

	if ts.Feeds != nil && len(ts.Feeds) > 0 {
		f1, _ = strconv.ParseBool(ts.Feeds[0].Field7)
		f2, _ = strconv.ParseBool(ts.Feeds[0].Field8)
		f3, _ = strconv.ParseBool(ts.Feeds[0].Field6)
		trace.Println("Virtual inputs from API: ", f1, f2, f3)
	}
	return f1, f2, f3
}

func readTemperature() (temp float64) {
	dat, e := ioutil.ReadFile("/sys/bus/w1/devices/28-0417501596ff/w1_slave")
	if e != nil {
		panic(e)
	}
	str := string(dat)
	tempStr := str[len(str)-6 : len(str)-3]
	trace.Println("tempStr: ", tempStr)
	i, _ := strconv.ParseInt(tempStr, 10, 32)
	return float64(i) / float64(10)
}

// 51cm: Distance Ground to sensor
// 35cm: Maximum possible fill level
func readDistance() int {
	var i = 0
	var begin time.Time
	var end time.Time
	var status rpio.State
	info.Println(fillLevelEcho.Read())
	fillLevel.Low()
	info.Println(fillLevelEcho.Read())
	time.Sleep(time.Second * 2)
	info.Println(fillLevelEcho.Read())
	fillLevel.High()
	info.Println(fillLevelEcho.Read())
	time.Sleep(time.Microsecond * 10)
	fillLevel.Low()
	for {
		status = fillLevelEcho.Read()
		if status == rpio.High {
			break
		}
		//		begin = time.Now()
		//		i++
	}
	begin = time.Now()
	info.Println("i = ", i)
	i = 0
	end = time.Now()
	for {
		status = fillLevelEcho.Read()
		if status == rpio.Low {
			break
		}
		end = time.Now()
		i++
	}
	info.Println("i2 = ", i)
	diff := end.Sub(begin)
	info.Println("diff = ", diff)
	timeDiff := float64(diff.Nanoseconds()) / 1000000000.0
	info.Println("timeDiff = ", timeDiff, " cm ", 52-int(timeDiff*34300.0/2))
	// https://www.modmypi.com/blog/hc-sr04-ultrasonic-range-sensor-on-the-raspberry-pi
	return 52 - int(timeDiff*34300.0/2)
}

func switchOffSprinkler() {
	info.Println("Finally switching Sprinkler OFF")
}

func readInputs(currentOutput types.Outputs) (d types.Inputs) {
	trace.Println("> Read Inputs")
	pumpOn, sprinklerOn, fillFontaineValve = readVirtualInputs(currentOutput)
	res := types.Inputs{
		Temperature:       readTemperature(),
		Brightness:        int(brightness.Read()),
		Wetness:           int(wetness.Read()),
		FlowRate:          int(flowRate.Read()),
		FillLevel:         readDistance(),
		PumpOn:            pumpOn,
		SprinklerOn:       sprinklerOn,
		FillFontaineValve: fillFontaineValve}
	trace.Println("Working with inputs: ", res)
	return res
}

func enoughWaterInFontaine(fillLevel int) bool {
	return fillLevel < types.FontaineFull
}

func timeForWatering() bool {
	hour, min, _ := time.Now().Clock()
	minutes1 := int(types.MaxDurationWateringPeriod1 / 60)
	minutes2 := int(types.MaxDurationWateringPeriod2 / 60)
	res :=
		(hour == types.StartWateringHours1 &&
			min >= types.StartWateringMinutes1 &&
			min < types.StartWateringMinutes1+minutes1) ||
			(hour == types.StartWateringHours2 &&
				min >= types.StartWateringMinutes2 &&
				min < types.StartWateringMinutes2+minutes2)
	if res {
		info.Println("Time for watering ", hour, ":", min)
	}
	return res
}

func isDaylight() bool {
	hour := time.Now().Hour()
	return hour >= types.FountainStartHour && hour <= types.FountainStopHour
}

func dryGround(wetness int) bool {
	return wetness < types.DryGround
}

func process(inputs types.Inputs, currentOutput types.Outputs) (outputs types.Outputs) {
	trace.Println("  Process Inputs")

	var output = types.Outputs{}
	var closeMainValve bool
	// Fontaine
	var fontaine = inputs.PumpOn && enoughWaterInFontaine(inputs.FillLevel) && isDaylight()
	if currentOutput.Fontaine != fontaine {
		info.Println("Switching Fontaine: ", boolToStr(fontaine), ", is it day? ", isDaylight())
	}
	output.Fontaine = fontaine

	// sprinkler
	var delaySprinklerValve time.Time
	var sprinklerValve = inputs.SprinklerOn ||
		(timeForWatering() && dryGround(inputs.Wetness))
	if currentOutput.SprinklerValve != sprinklerValve &&
		delaySprinklerValve.Before(time.Now().Add(-time.Second*10)) {
		info.Println("Switching SprinklerValve: ", boolToStr(sprinklerValve))
		if !sprinklerValve {
			// Delay switching off to empty the tube
			info.Println("Delaying Switch OFF")
			time.AfterFunc(time.Second*10, switchOffSprinkler)
		}
	}
	if delaySprinklerValve.Before(time.Now().Add(-time.Second * 10)) {
		output.SprinklerValve = sprinklerValve
		closeMainValve = false
	}

	// fill fontaine
	var fontaineValve = !enoughWaterInFontaine(inputs.FillLevel) || inputs.FillFontaineValve
	if currentOutput.FontaineValve != fontaineValve {
		info.Println("Switching FontaineValve: ", boolToStr(fontaineValve))
	}
	output.FontaineValve = fontaineValve

	// water on the system
	var mainValve = (output.FontaineValve || sprinklerValve) && !closeMainValve
	if currentOutput.MainValve != mainValve {
		info.Println("Switching MainValve: ", boolToStr(mainValve))
	}
	output.MainValve = mainValve

	trace.Println("Calculated output: ", output)
	return output
}

func writeOutput(output types.Outputs) {
	trace.Println("< Write Outputs")

	// Relais 1 // IN 1 // Main Valve
	mainValve.High()
	if output.MainValve {
		trace.Println("Main valve ON")
	} else {
		//mainValve.Low()
		trace.Println("Main valve OFF")
	}
	// Relais 2 // IN 2 // Sprinkler
	if output.SprinklerValve {
		sprinkler.High()
		trace.Println("Sprinkler ON")
	} else {
		sprinkler.Low()
		trace.Println("Sprinkler OFF")
	}
	// Relais 3 // IN 3 // Fill Fontaine
	if output.FontaineValve {
		fillFountain.High()
		trace.Println("FillFontaine ON")
	} else {
		fillFountain.Low()
		trace.Println("FillFontaine OFF")
	}
	// Relais 3 // IN 3 // Pump
	if output.Fontaine {
		pump.Low()
		trace.Println("Fontaine ON")
	} else {
		pump.High()
		trace.Println("Fontaine OFF")
	}
	trace.Println("Write Output: ", output)
}

func initializePins() {
	wetness.Input()
	mainValve.Output()
	fillFountain.Output()
	pump.Output()
	ledLights.Output()
	sprinkler.Output()
	fillLevel.Output()
	fillLevelEcho.Input()
}

func main() {
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Unmap gpio memory when done
	defer rpio.Close()

	initializePins()

	cnt := 1
	var outputs = types.Outputs{}
	for {
		inputs := readInputs(outputs)

		var newOutput = process(inputs, outputs)

		writeOutput(newOutput)

		if cnt%thingsSpeakInterval == 0 || !reflect.DeepEqual(newOutput, outputs) {
			cnt = 0
			outputs = newOutput
			sendData(types.Capture{Input: inputs, Output: newOutput})
		}
		cnt++
		time.Sleep(1000 * time.Millisecond)
	}
}
