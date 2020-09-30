package gcodetools

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
)

type GcodeBuilder struct {
	NozzleSize          float64
	LayerHeight         float64
	ExtrusionWidth      float64
	ExtrusionMultiplier float64
	FilamentDiameter    float64
	PrintFeedrate       float64
	TravelFeedrate      float64
	//
	buf          []GcodeLine
	machineState MachineState
	minifier     *GcodeMinifierConfig
}

func (b *GcodeBuilder) extrusionPerLinearMm() float64 {
	return b.LayerHeight * b.ExtrusionWidth / (math.Pi * math.Pow(b.FilamentDiameter/2, 2))
}

// TODO: gcode commands to heat extruder?
// TODO: method to add a multi-line string of gcode (such as for start/end gcode)

func (b *GcodeBuilder) extrusionLengthForPrintMove(x, y, z float64) float64 {
	if !b.machineState.RelativeExtrusion {
		// TODO: support absolute extrusion
		panic("relative extrusion is required")
	}
	return math.Sqrt(math.Pow(x-b.machineState.X, 2)+
		math.Pow(y-b.machineState.Y, 2)+
		math.Pow(z-b.machineState.Z, 2)) *
		b.extrusionPerLinearMm()
}

func (b *GcodeBuilder) RelativeExtrusion() {
	b.AddGcodeLine(GcodeLine{CmdLetter: M, CmdNumber: 83})
}

func (b *GcodeBuilder) Home() {
	b.AddGcodeLine(GcodeLine{CmdLetter: G, CmdNumber: 28})
}

func (b *GcodeBuilder) AddGcodeLine(line GcodeLine) {
	if b.minifier == nil {
		b.minifier = &GcodeMinifierConfig{
			RemoveComments:    false,
			AllowUnknownGcode: true,
		}
		b.minifier.Init()
	}
	b.minifier.MinifyGcodeLineInPlace(&b.machineState, &line)
	b.buf = append(b.buf, line)
}

func (b *GcodeBuilder) TravelTo(x, y, z float64) {
	b.AddGcodeLine(GcodeLine{
		CmdLetter: G, CmdNumber: 0,
		Xvalid: true, X: x,
		Yvalid: true, Y: y,
		Zvalid: true, Z: z,
		Feedrate: b.TravelFeedrate,
	})
}

func (b *GcodeBuilder) TravelToXY(x, y float64) {
	b.TravelTo(x, y, b.machineState.Z)
}

func (b *GcodeBuilder) PrintTo(x, y, z float64) {
	b.AddGcodeLine(GcodeLine{
		CmdLetter: G, CmdNumber: 1,
		Xvalid: true, X: x,
		Yvalid: true, Y: y,
		Zvalid: true, Z: z,
		Evalid: true, E: b.extrusionLengthForPrintMove(x, y, z),
		Feedrate: b.PrintFeedrate,
	})
}

func (b *GcodeBuilder) PrintToXY(x, y float64) {
	b.PrintTo(x, y, b.machineState.Z)
}

func (b *GcodeBuilder) TravelToF(x, y, z, feedrate float64) {
	b.AddGcodeLine(GcodeLine{
		CmdLetter: G, CmdNumber: 0,
		Xvalid: true, X: x,
		Yvalid: true, Y: y,
		Zvalid: true, Z: z,
		Feedrate: feedrate,
	})
}

func (b *GcodeBuilder) TravelToXYF(x, y, feedrate float64) {
	b.TravelToF(x, y, b.machineState.Z, feedrate)
}

func (b *GcodeBuilder) PrintToF(x, y, z, feedrate float64) {
	b.AddGcodeLine(GcodeLine{
		CmdLetter: G, CmdNumber: 1,
		Xvalid: true, X: x,
		Yvalid: true, Y: y,
		Zvalid: true, Z: z,
		Feedrate: feedrate,
		Evalid:   true, E: b.extrusionLengthForPrintMove(x, y, z),
	})
}

func (b *GcodeBuilder) PrintToXYF(x, y, feedrate float64) {
	b.PrintToF(x, y, b.machineState.Z, feedrate)
}

func (b *GcodeBuilder) Comment(comment string) {
	b.AddGcodeLine(GcodeLine{Comment: &comment})
}

func (b *GcodeBuilder) ToWriter(writer io.Writer) error {
	minifier := b.minifier
	if minifier == nil {
		minifier = &GcodeMinifierConfig{
			RemoveComments:    false,
			AllowUnknownGcode: true,
		}
		minifier.Init()
	}
	for _, line := range b.buf {
		lineStr := minifier.formatGcode(&line)
		_, err := fmt.Fprintln(writer, lineStr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *GcodeBuilder) ToString() string {
	var buf bytes.Buffer
	_ = b.ToWriter(&buf)
	return buf.String()
}

func (b *GcodeBuilder) ToFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	err = b.ToWriter(file)
	if err != nil {
		return err
	}
	return nil
}
