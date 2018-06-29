package types

// FontaineFull is the threshold in millimeter from sonic sensor to water surface. Lower values indicate the fontain holds a higher level of water
const FontaineFull = 300

// DryGround is the treshold that indicates a dry ground. Lower values are more dry than higher values
const DryGround = 1000

const StartWateringHours1 = 4
const StartWateringMinutes1 = 0
const MaxDurationWateringPeriod1 = 300

const StartWateringHours2 = 23
const StartWateringMinutes2 = 0
const MaxDurationWateringPeriod2 = 300

// Capture defines the data set send to ThingsSpeak
type Capture struct {
	Input  Inputs
	Output Outputs
}

// TS is a simple request format to POST ThingSpeak
type TS struct {
	Field1  string
	Field2  string
	Api_key string
}

// Feed is a single entry in ThingSpeak API
type Feed struct {
	Field1 string `json:"field1"`
	Field2 string `json:"field2"`
	Field3 string `json:"field3"`
	Field4 string `json:"field4"`
	Field5 string `json:"field5"`
	Field6 string `json:"field6"`
	Field7 string `json:"field7"`
	Field8 string `json:"field8"`
}

// ThingSpeakQuery is a query to ThingSpeak API
type ThingSpeakQuery struct {
	Feeds []Feed `json:"feeds"`
}

// Inputs defines all possible input signals
type Inputs struct {
	Temperature       int
	Brightness        int
	Wetness           int
	FlowRate          int
	FillLevel         int
	PumpOn            bool
	SprinklerOn       bool
	FillFontaineValve bool
}

// Outputs defines all possible output signals
type Outputs struct {
	MainValve      bool
	FontaineValve  bool
	SprinklerValve bool
	Fontaine       bool
}
