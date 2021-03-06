package gcodetools

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestGcodeMinifierConfig_MinifyGcodeStr(t *testing.T) {
	cfg := (&GcodeMinifierConfig{
		RemoveComments: true,
	}).Init()

	state := MachineState{}
	gcodeStr := `
		G28 ; home the printer
		M83 ; set relative extrusion
		G0 X0.0 Y0.0
		
		; print a square
		G0 X0.0   Y0.0        F1000.0 ; bottom left
		G1 X0.0   Y100.0 E0.5 F1000.0 ; top left
		G1 X100.0 Y100.0 E0.5 F1000.0 ; top right 
		G1 X100.0 Y0.0   E0.5 F1000.0 ; bottom right
		G1 X0.0   Y0.0   E0.5 F1000.0 ; back to the start
		
		G1 X0.0 Y0.0 Z10
		G1 X0.0 Y0.0 Z10 ; repeated line for minifier to remove
		
		G1 X0.0 Y0.0 Z10 E-10
		G1 X0.0 Y0.0 Z10 E-10 ; this repeated line won't get removed because E is relative
		`
	outputGcodeStr, finalState := cfg.MinifyGcodeStr(state, gcodeStr)

	expected := strings.Join([]string{
		"G28",
		"M83",
		"G0 F1000",
		"G1 Y100 E.5",
		"G1 X100 E.5",
		"G1 Y0 E.5",
		"G1 X0 E.5",
		"G1 Z10",
		"G1 E-10",
		"G1 E-10",
		"", // make sure there's a trailing newline
	}, "\n")
	assert.Equal(t, expected, outputGcodeStr)
	assert.Equal(t, MachineState{X: 0, Y: 0, Z: 10, E: 0, EAbsolute: -18, Feedrate: 1000, RelativeExtrusion: true, IsHomed: true}, finalState)
}

func TestGcodeMinifierConfig_MinifyGcodeLineInPlace(t *testing.T) {
	cfg := (&GcodeMinifierConfig{
		RemoveComments: true,
	}).Init()

	state := MachineState{IsHomed: true}
	line := mustParseLine(`G0 X20 Y0 ; comment`)
	cfg.MinifyGcodeLineInPlace(&state, line)

	assert.Equal(t, 20.0, state.X)
	assert.True(t, line.Xvalid)
	assert.Equal(t, 20.0, line.X)
	assert.False(t, line.Yvalid)
	assert.Equal(t, 0.0, line.Y)
}

type gcodeFormatTestPair struct {
	gcode                            GcodeLine
	str                              string
	xyDecimals, zDecimals, eDecimals int
}

func TestFormatGcode(t *testing.T) {
	comment824634126112 := `; comment text`
	comment824634126176 := `; move Z axis up`
	cases := []gcodeFormatTestPair{
		{GcodeLine{}, ``, 4, 4, 8},
		{GcodeLine{CmdLetter: 'G', CmdNumber: 0, X: 1, Xvalid: true, Y: 2, Yvalid: true, Z: 3, Zvalid: true, E: 4, Evalid: true}, `G0 X1 Y2 Z3 E4`, 4, 4, 8},
		{GcodeLine{CmdLetter: 'G', CmdNumber: 0, X: 0.21, Xvalid: true, Y: 20, Yvalid: true, Z: 3, Zvalid: true}, `G0 X.21 Y20 Z3`, 4, 4, 8},
		{GcodeLine{CmdLetter: 'G', CmdNumber: 1, X: 1, Xvalid: true, Y: 2, Yvalid: true, Z: 0.31, Zvalid: true}, `G1 X1 Y2 Z.31`, 4, 4, 8},
		{GcodeLine{CmdLetter: 'G', CmdNumber: 0, X: 100, Xvalid: true, Feedrate: 1234}, `G0 X100 F1234`, 4, 4, 8},
		{GcodeLine{CmdLetter: 'M', CmdNumber: 83}, `M83`, 4, 4, 8},
		{GcodeLine{NumericParams: map[uint8]float64{'T': 0}}, `T0`, 4, 4, 8},
		{GcodeLine{CmdLetter: 'M', CmdNumber: 118, StringParams: map[uint8]string{'S': `"Hello_Duet"`}}, `M118 S"Hello_Duet"`, 4, 4, 8},
		{GcodeLine{CmdLetter: 'M', CmdNumber: 118, StringParams: map[uint8]string{'S': `Hello_Duet`}}, `M118 SHello_Duet`, 4, 4, 8},
		{GcodeLine{Comment: &comment824634126112}, comment824634126112, 4, 4, 8},
		{GcodeLine{CmdLetter: 'G', CmdNumber: 1, Z: 20, Zvalid: true, Feedrate: 200, Comment: &comment824634126176}, `G1 Z20 F200 ; move Z axis up`, 4, 4, 8},
		///
		{GcodeLine{CmdLetter: 'G', CmdNumber: 0, X: 1.2345, Xvalid: true}, `G0 X1.2345`, 4, 4, 8},
		{GcodeLine{CmdLetter: 'G', CmdNumber: 0, Y: 1.23456, Yvalid: true}, `G0 Y1.2346`, 4, 4, 8},
		{GcodeLine{CmdLetter: 'G', CmdNumber: 0, X: 1.23456, Xvalid: true}, `G0 X1.23456`, 5, 4, 8},
		{GcodeLine{CmdLetter: 'G', CmdNumber: 0, Z: 0.00004, Zvalid: true}, `G0 Z0`, 4, 4, 8},
		{GcodeLine{CmdLetter: 'G', CmdNumber: 0, Z: 0.00005, Zvalid: true}, `G0 Z.0001`, 4, 4, 8},
		{GcodeLine{CmdLetter: 'G', CmdNumber: 0, E: 1.00001, Evalid: true}, `G0 E1.00001`, 4, 4, 8},
		{GcodeLine{CmdLetter: 'G', CmdNumber: 0, E: 1.000000004, Evalid: true}, `G0 E1`, 4, 4, 8},
		{GcodeLine{CmdLetter: 'G', CmdNumber: 0, E: 1.000000004, Evalid: true}, `G0 E1.000000004`, 4, 4, 9},
		{GcodeLine{CmdLetter: 'G', CmdNumber: 0, E: 0.000000004, Evalid: true}, `G0 E.000000004`, 4, 4, 9},
	}

	for _, pair := range cases {
		gcode := pair.gcode
		str := formatGcode(&gcode, pair.xyDecimals, pair.zDecimals, pair.eDecimals)
		//fmt.Printf("{%v, %#q},\n", gcode, str)
		assert.Equal(t, pair.str, str)
	}

}

func TestFloatToSmallestString(t *testing.T) {
	assert.Equal(t, FloatToSmallestString(1024, 4), "1024")
	assert.Equal(t, FloatToSmallestString(300, 4), "300")

	assert.Equal(t, FloatToSmallestString(12.5, 4), "12.5")
	assert.Equal(t, FloatToSmallestString(12.111111111111, 4), "12.1111")
	assert.Equal(t, FloatToSmallestString(1.499999999999, 4), "1.5")
	assert.Equal(t, FloatToSmallestString(123.12345555555, 4), "123.1235")

	assert.Equal(t, FloatToSmallestString(0.01, 4), ".01")
	assert.Equal(t, FloatToSmallestString(0.001, 4), ".001")
	assert.Equal(t, FloatToSmallestString(0.0001, 4), ".0001")
	assert.Equal(t, FloatToSmallestString(0.00001, 4), "0")
	assert.Equal(t, FloatToSmallestString(0.00005, 4), ".0001")
	assert.Equal(t, FloatToSmallestString(0.00001, 5), ".00001")
}
