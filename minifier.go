package gcodetools

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

var defaultThreshold = 0.001

type GcodeMinifierConfig struct {
	RemoveComments                   bool
	Threshold                        float64
	thresholdSqr                     float64
	XYDecimals, ZDecimals, EDecimals int
	AllowUnknownGcode                bool
	////
}

var DefaultGcodeMinifierConfig = GcodeMinifierConfig{
	RemoveComments: true,
	Threshold:      defaultThreshold,
	thresholdSqr:   defaultThreshold * defaultThreshold,
	XYDecimals:     4,
	ZDecimals:      5,
	EDecimals:      8,
}

func (cfg *GcodeMinifierConfig) Init() *GcodeMinifierConfig {
	if cfg.Threshold == 0 {
		cfg.Threshold = DefaultGcodeMinifierConfig.Threshold
	}
	if cfg.thresholdSqr == 0 {
		cfg.thresholdSqr = cfg.Threshold * cfg.Threshold
	}
	if cfg.XYDecimals == 0 {
		cfg.XYDecimals = DefaultGcodeMinifierConfig.XYDecimals
	}
	if cfg.ZDecimals == 0 {
		cfg.ZDecimals = DefaultGcodeMinifierConfig.ZDecimals
	}
	if cfg.EDecimals == 0 {
		cfg.EDecimals = DefaultGcodeMinifierConfig.EDecimals
	}
	return cfg
}

func (cfg *GcodeMinifierConfig) float64ApproxEq(a, b float64) bool {
	return math.Abs(a-b) < cfg.Threshold
}

func (cfg *GcodeMinifierConfig) formatGcode(g *GcodeLine) string {
	return formatGcode(g, cfg.XYDecimals, cfg.ZDecimals, cfg.EDecimals)
}

func (cfg *GcodeMinifierConfig) MinifyGcodeStr(initialState MachineState, gcodeStr string) (output string, state MachineState) {
	state = initialState
	strLines := strings.Split(gcodeStr, "\n")
	gcodeLines := make([]GcodeLine, 0, len(strLines))
	for _, line := range strLines {
		g, err := ParseLine(line)
		if err != nil {
			panic(err) // TODO: better error handling
		}
		cfg.MinifyGcodeLineInPlace(&state, &g)
		if !g.Empty() {
			gcodeLines = append(gcodeLines, g)
		}
	}

	// len+1 so that there's a trailing newline
	outputLines := make([]string, len(gcodeLines)+1)
	for i, gcodeLine := range gcodeLines {
		outputLines[i] = cfg.formatGcode(&gcodeLine)
	}
	output = strings.Join(outputLines, "\n")
	return
}

func (cfg *GcodeMinifierConfig) MinifyGcodeLineInPlace(state *MachineState, line *GcodeLine) {
	if cfg.RemoveComments {
		line.Comment = nil
	}
	if line.NumericParams != nil && len(line.NumericParams) == 0 {
		line.NumericParams = nil
	}
	if line.StringParams != nil && len(line.StringParams) == 0 {
		line.StringParams = nil
	}

	if line.CommentOnly() || line.Empty() {
		return
	}

	if line.IsG(28) {
		state.X = 0
		state.Y = 0
		state.Z = 0
		state.IsHomed = true
		return
	}
	if line.IsG(0) || line.IsG(1) {
		cfg.minifyAbsoluteG0G1Move(state, line)
		return
	}

	if line.IsM(83) {
		state.RelativeExtrusion = true
		state.E = 0
		return
	}
	if line.IsM(82) {
		state.RelativeExtrusion = false
		state.E = state.EAbsolute
		return
	}

	if line.IsG(90) { // G90 - Absolute Positioning
		// TODO
	}
	if line.IsG(91) { // G91 - Relative Positioning
		// TODO
	}

	if !cfg.AllowUnknownGcode {
		panic("unimplemented: " + line.String())
	}
}

// move must be G0 or G1
func (cfg *GcodeMinifierConfig) minifyAbsoluteG0G1Move(state *MachineState, line *GcodeLine) {
	if !state.IsHomed || state.RelativeCoordinates {
		panic("error: relative moves are unimplemented")
	}

	if line.NumericParams != nil || line.StringParams != nil {
		panic("error: extra G0/G1 parameters are unimplemented") // TODO handle extra parameters on G0/G1 (or maybe just pass them through unchanged?)
	}
	if line.Xvalid && !cfg.float64ApproxEq(line.X, state.X) {
		state.X = line.X
	} else {
		line.Xvalid = false
		line.X = 0
	}
	if line.Yvalid && !cfg.float64ApproxEq(line.Y, state.Y) {
		state.Y = line.Y
	} else {
		line.Yvalid = false
		line.Y = 0
	}
	if line.Zvalid && !cfg.float64ApproxEq(line.Z, state.Z) {
		state.Z = line.Z
	} else {
		line.Zvalid = false
		line.Z = 0
	}
	if !state.RelativeExtrusion {
		if line.Evalid && !cfg.float64ApproxEq(line.E, state.E) {
			state.E = line.E
			state.EAbsolute = line.E
		} else {
			line.Evalid = false
			line.E = 0
		}
	} else if line.Evalid {
		if cfg.float64ApproxEq(line.E, 0) {
			line.Evalid = false
		}
		// keep track of E in relative extrusion, anyway
		state.EAbsolute += line.E
	}
	if line.Feedrate != 0 && !cfg.float64ApproxEq(line.Feedrate, state.Feedrate) {
		state.Feedrate = line.Feedrate
	} else if line.Feedrate != 0 && cfg.float64ApproxEq(line.Feedrate, state.Feedrate) {
		line.Feedrate = 0
	}

	// if we're left with a do-nothing move, just empty it
	if !(line.Xvalid || line.Yvalid || line.Zvalid || line.Evalid || line.Feedrate != 0 ||
		line.NumericParams != nil || line.StringParams != nil) {
		*line = GcodeLine{}
	}
}

type MachineState struct {
	X                   float64
	Y                   float64
	Z                   float64
	E                   float64
	EAbsolute           float64
	Feedrate            float64
	RelativeExtrusion   bool
	RelativeCoordinates bool
	IsHomed             bool
	// TODO: temperature
}

func formatGcode(g *GcodeLine, xyDecimals, zDecimals, eDecimals int) string {
	if g.Empty() {
		return ""
	}
	parts := []string{}
	//var buf bytes.Buffer
	if g.CmdLetter != 0 {
		parts = append(parts, fmt.Sprintf("%c%d", g.CmdLetter, g.CmdNumber))
	}
	if g.Xvalid {
		parts = append(parts, "X"+FloatToSmallestString(g.X, xyDecimals))
	}
	if g.Yvalid {
		parts = append(parts, "Y"+FloatToSmallestString(g.Y, xyDecimals))
	}
	if g.Zvalid {
		parts = append(parts, "Z"+FloatToSmallestString(g.Z, zDecimals))
	}
	if g.Evalid {
		parts = append(parts, "E"+FloatToSmallestString(g.E, eDecimals))
	}
	if g.Feedrate != 0 {
		parts = append(parts, "F"+FloatToSmallestString(g.Feedrate, 0))
	}
	if g.NumericParams != nil {
		for u, f := range g.NumericParams {
			parts = append(parts, fmt.Sprintf("%c%v", u, FloatToSmallestString(f, xyDecimals)))
		}
	}
	if g.StringParams != nil {
		for u, s := range g.StringParams {
			parts = append(parts, fmt.Sprintf("%c%s", u, s))
		}
	}
	if g.Comment != nil {
		parts = append(parts, *g.Comment)
	}

	return strings.Join(parts, " ")
}

func FloatToSmallestString(f float64, decimals int) string {
	s := strconv.FormatFloat(f, 'f', decimals, 64)
	if !strings.Contains(s, ".") {
		return s
	}
	// this is faster than using strings.TrimRight()

	// 3d printers should be able to interpret less-than-1 numbers without leading zeros (like ".01" instead of "0.01")
	// RepRapFirmware should be fine (ref. https://github.com/Duet3D/RRFLibraries/blob/master/src/General/SafeStrtod.cpp)
	// Marlin should be fine, too: https://github.com/MarlinFirmware/Marlin/blob/2.0.x/Marlin/src/gcode/parser.h#L248
	if s[0] == '0' {
		s = s[1:]
	}
	for s[len(s)-1] == '0' {
		s = s[:len(s)-1]
	}
	if s[len(s)-1] == '.' {
		s = s[:len(s)-1]
	}
	if len(s) == 0 {
		// special case, to make sure we still have a number after all this
		return "0"
	}
	return s
}
