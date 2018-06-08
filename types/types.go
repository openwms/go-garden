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

// Inputs defines all possible input signals
type Inputs struct {
	Temperature int
	Brightness  int
	Moisture    int
	FlowRate    int
	Sonic       int
}

// Outputs defines all possible output signals
type Outputs struct {
	MainValve     bool
	FontaineValve bool
	SpinklerValve bool
	Fontaine      bool
}
