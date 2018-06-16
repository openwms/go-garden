package types

// Capture defines the data set send to ThingsSpeak
type Capture struct {
	Input  Inputs
	Output Outputs
}

type TS struct {
	Field1  string
	Field2  string
	Api_key string
}

// Feed is a single entry in ThingSpeak API
type Feed struct {
	Field1 string
	Field2 string
	Field3 string
	Field4 string
	Field5 string
	Field6 string
	Field7 string
	Field8 string
}

// ThingSpeakQuery is a query to ThingSpeak API
type ThingSpeakQuery struct {
	Feeds []Feed
}

// Inputs defines all possible input signals
type Inputs struct {
	Temperature int
	Brightness  int
	Wetness     int
	FlowRate    int
	FillLevel   int
	PumpOn      bool
	SprinklerOn bool
}

// Outputs defines all possible output signals
type Outputs struct {
	MainValve     bool
	FontaineValve bool
	SpinklerValve bool
	Fontaine      bool
}
