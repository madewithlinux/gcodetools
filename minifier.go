package gcodetools

import (
	"fmt"
	"strconv"
	"strings"
)

type GcodeMinifierConfig struct {
	RemoveComments bool
	////

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
