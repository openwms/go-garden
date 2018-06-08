package types

// Capture defines the data set send to ThingsSpeak
type Capture struct {
	input  Inputs
	output Outputs
}

// Inputs defines all possible input signals
type Inputs struct {
	temperature int
	brightness  int
	moisture    int
	flowRate    int
	sonic       int
}

// Outputs defines all possible output signals
type Outputs struct {
	mainValve     bool
	fontaineValve bool
	spinklerValve bool
	fontaine      bool
}
