package gcodetools

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

const CommentChar = ';'
const G = byte('G')
const M = byte('M')

type GcodeLine struct {
	CmdLetter     uint8 // e.g. G0/G1 or M83 or whatever
	CmdNumber     uint16
	X             float64
	Y             float64
	Z             float64
	E             float64
	Xvalid        bool
	Yvalid        bool
	Zvalid        bool
	Evalid        bool
	Feedrate      float64 // Feedrate == 0 is obviously invalid
	NumericParams map[uint8]float64
	StringParams  map[uint8]string
	Comment       *string
}

func (g GcodeLine) String() string {
	var buf bytes.Buffer
	//if g.Comment != nil {
	//	_, _ = fmt.Fprintf(&buf, "comment%d := %q\n", g.Comment, *g.Comment)
	//}
	buf.WriteString("GcodeLine{")
	if g.CmdLetter != 0 { // if g.CmdLetter == 0, this line is probably a comment (or some other thing we don't know how to parse...)
		_, _ = fmt.Fprintf(&buf, "CmdLetter: '%c', CmdNumber: %v,", g.CmdLetter, g.CmdNumber)
	}
	if g.Xvalid {
		_, _ = fmt.Fprintf(&buf, "X: %v, Xvalid: true,", g.X)
	}
	if g.Yvalid {
		_, _ = fmt.Fprintf(&buf, "Y: %v, Yvalid: true,", g.Y)
	}
	if g.Zvalid {
		_, _ = fmt.Fprintf(&buf, "Z: %v, Zvalid: true,", g.Z)
	}
	if g.Evalid {
		_, _ = fmt.Fprintf(&buf, "E: %v, Evalid: true,", g.E)
	}
	if g.Feedrate != 0 {
		_, _ = fmt.Fprintf(&buf, "Feedrate: %v,", g.Feedrate)
	}
	if g.NumericParams != nil {
		_, _ = fmt.Fprint(&buf, "NumericParams: map[uint8]float64{")
		for u, f := range g.NumericParams {
			_, _ = fmt.Fprintf(&buf, "'%c': %v,", u, f)
		}
		_, _ = fmt.Fprint(&buf, "},")
	}
	if g.StringParams != nil {
		_, _ = fmt.Fprint(&buf, "StringParams: map[uint8]string{")
		for u, s := range g.StringParams {
			_, _ = fmt.Fprintf(&buf, "'%c': %#q,", u, s)
		}
		_, _ = fmt.Fprint(&buf, "},")
	}
	if g.Comment != nil {
		_, _ = fmt.Fprintf(&buf, "Comment: &comment%d,", g.Comment)
	}
	buf.WriteString("}")
	return buf.String()
}

func (g *GcodeLine) Empty() bool {
	return g.CmdLetter == 0 &&
		g.CmdNumber == 0 &&
		!g.Xvalid &&
		!g.Yvalid &&
		!g.Zvalid &&
		!g.Evalid &&
		g.Feedrate == 0 &&
		(g.NumericParams == nil || len(g.NumericParams) == 0) &&
		(g.StringParams == nil || len(g.StringParams) == 0) &&
		(g.Comment == nil || len(*g.Comment) == 0)
}

//var gcodeLineRegexp = regexp.MustCompile(`([GgMm]\d+)(?:\s+([A-Za-z]\S*))*(;.*)?`)

func ParseLine(str string) (line GcodeLine, err error) {
	i := 0
	for isSpace(str[i]) {
		i++
	}
	if str[i] == CommentChar {
		comment := str[i:]
		line.Comment = &comment
		return
	}

	commentStartChar := i
	for ; commentStartChar < len(str); commentStartChar++ {
		if str[commentStartChar] == CommentChar {
			break
		}
	}
	if commentStartChar < len(str) {
		comment := str[commentStartChar:]
		line.Comment = &comment
	}

	// TODO: support tabs in gcode
	// TODO: support gcodes with quoted string parameters that have spaces in them https://duet3d.dozuki.com/Wiki/Gcode#Section_Quoted_strings
	splits := strings.Split(str[i:commentStartChar], " ")
	for _, split := range splits {
		match := []byte(split)
		if len(match) == 0 {
			continue
		}
		switch match[0] {
		case 'G', 'g', 'M', 'm':
			if match[0] == 'G' || match[0] == 'g' {
				line.CmdLetter = G
			} else {
				line.CmdLetter = M
			}
			var u64 uint64
			u64, err = strconv.ParseUint(string(match[1:]), 10, 16)
			line.CmdNumber = uint16(u64)
		case 'X', 'x':
			line.X, err = strconv.ParseFloat(string(match[1:]), 64)
			line.Xvalid = true
		case 'Y', 'y':
			line.Y, err = strconv.ParseFloat(string(match[1:]), 64)
			line.Yvalid = true
		case 'Z', 'z':
			line.Z, err = strconv.ParseFloat(string(match[1:]), 64)
			line.Zvalid = true
		case 'E', 'e':
			line.E, err = strconv.ParseFloat(string(match[1:]), 64)
			line.Evalid = true
		case 'F', 'f':
			line.Feedrate, err = strconv.ParseFloat(string(match[1:]), 64)
		default:
			f, parseFloatErr := strconv.ParseFloat(string(match[1:]), 64)
			if parseFloatErr != nil {
				if line.StringParams == nil {
					line.StringParams = map[uint8]string{}
				}
				line.StringParams[match[0]] = string(match[1:])
			} else {
				if line.NumericParams == nil {
					line.NumericParams = map[uint8]float64{}
				}
				line.NumericParams[match[0]] = f
			}
		}
		if err != nil {
			return
		}
	}
	return
}

func isSpace(c uint8) bool {
	return (c == ' ') || (c == '\t') || (c == '\n')
}
