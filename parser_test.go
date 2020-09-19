package gcodetools_test

import (
	"fmt"
	. "github.com/madewithlinux/gcodetools"
	"github.com/stretchr/testify/assert"
	"testing"
)

type toStringTestPair struct {
	gcode GcodeLine
	str   string
}

func TestGcodeLine_String(t *testing.T) {
	cases := []toStringTestPair{
		{GcodeLine{}, `GcodeLine{}`},
		{GcodeLine{CmdLetter: 'G', CmdNumber: 0, X: 1, Xvalid: true, Y: 2, Yvalid: true, Z: 3, Zvalid: true, E: 4, Evalid: true}, `GcodeLine{CmdLetter: 'G', CmdNumber: 0,X: 1, Xvalid: true,Y: 2, Yvalid: true,Z: 3, Zvalid: true,E: 4, Evalid: true,}`},
		{GcodeLine{CmdLetter: 'G', CmdNumber: 0, X: 0.21, Xvalid: true, Y: 20, Yvalid: true, Z: 3, Zvalid: true}, `GcodeLine{CmdLetter: 'G', CmdNumber: 0,X: 0.21, Xvalid: true,Y: 20, Yvalid: true,Z: 3, Zvalid: true,}`},
		{GcodeLine{CmdLetter: 'G', CmdNumber: 1, X: 1, Xvalid: true, Y: 2, Yvalid: true, Z: 0.31, Zvalid: true}, `GcodeLine{CmdLetter: 'G', CmdNumber: 1,X: 1, Xvalid: true,Y: 2, Yvalid: true,Z: 0.31, Zvalid: true,}`},
		{GcodeLine{CmdLetter: 'G', CmdNumber: 0, X: 100, Xvalid: true, Feedrate: 1234}, `GcodeLine{CmdLetter: 'G', CmdNumber: 0,X: 100, Xvalid: true,Feedrate: 1234,}`},
		{GcodeLine{CmdLetter: 'M', CmdNumber: 83}, `GcodeLine{CmdLetter: 'M', CmdNumber: 83,}`},
		{GcodeLine{NumericParams: map[uint8]float64{'T': 0}}, `GcodeLine{NumericParams: map[uint8]float64{'T': 0,},}`},
		{GcodeLine{CmdLetter: 'M', CmdNumber: 118, StringParams: map[uint8]string{'S': `"Hello_Duet"`}}, "GcodeLine{CmdLetter: 'M', CmdNumber: 118,StringParams: map[uint8]string{'S': `\"Hello_Duet\"`,},}"},
		{GcodeLine{CmdLetter: 'M', CmdNumber: 118, StringParams: map[uint8]string{'S': `Hello_Duet`}}, "GcodeLine{CmdLetter: 'M', CmdNumber: 118,StringParams: map[uint8]string{'S': `Hello_Duet`,},}"},
		// this is not very testable because the order of strings in a map is undefined/inconsistent
		//{GcodeLine{CmdLetter: 'M', CmdNumber: 587,StringParams: map[uint8]string{'P': `"Network_Password"`,'S': `"Network_SSID"`,},}, "GcodeLine{CmdLetter: 'M', CmdNumber: 587,StringParams: map[uint8]string{'S': `\"Network_SSID\"`,'P': `\"Network_Password\"`,},}"},
	}

	for _, pair := range cases {
		gcode := pair.gcode
		//fmt.Printf("{%v, %#q},\n", gcode, gcode.String())
		assert.Equal(t, pair.str, gcode.String())
	}
	//fmt.Println()
}

func TestParseLine(t *testing.T) {
	testParsesAs(t, `G0 X1 Y2 Z3 E4`, GcodeLine{CmdLetter: 'G', X: 1, Y: 2, Z: 3, E: 4, Xvalid: true, Yvalid: true, Zvalid: true, Evalid: true})
	testParsesAs(t, ` G0 X.21 Y20 Z3`, GcodeLine{CmdLetter: 'G', CmdNumber: 0, X: 0.21, Xvalid: true, Y: 20, Yvalid: true, Z: 3, Zvalid: true})
	testParsesAs(t, `	G1 Z0.31 X1 Y2`, GcodeLine{CmdLetter: 'G', CmdNumber: 1, X: 1, Xvalid: true, Y: 2, Yvalid: true, Z: 0.31, Zvalid: true})
	testParsesAs(t, `G0 X100 F1234`, GcodeLine{CmdLetter: 'G', CmdNumber: 0, X: 100, Xvalid: true, Feedrate: 1234})
	testParsesAs(t, `M83`, GcodeLine{CmdLetter: 'M', CmdNumber: 83})

	comment824634126112 := `; comment text`
	testParsesAs(t, `; comment text`, GcodeLine{Comment: &comment824634126112})
	comment824634126176 := `; move Z axis up`
	testParsesAs(t, `G1 Z20 F200 ; move Z axis up`, GcodeLine{CmdLetter: 'G', CmdNumber: 1, Z: 20, Zvalid: true, Feedrate: 200, Comment: &comment824634126176})

	testParsesAs(t, `T0`, GcodeLine{NumericParams: map[uint8]float64{'T': 0}})

	testParsesAs(t, `M587 S"Network_SSID" P"Network_Password"`, GcodeLine{CmdLetter: 'M', CmdNumber: 587, StringParams: map[uint8]string{'S': `"Network_SSID"`, 'P': `"Network_Password"`}})
	// TODO: support gcodes with quoted string parameters that have spaces in them https://duet3d.dozuki.com/Wiki/Gcode#Section_Quoted_strings
	//testParsesAs(t, `M587 S"Network SSID" P"Network Password"`, GcodeLine{CmdLetter: 'M', CmdNumber: 587, StringParams: map[uint8]string{'S': `"Network_SSID"`, 'P': `"Network_Password"`}})

	// https://duet3d.dozuki.com/Wiki/Gcode#Section_M117_Display_Message
	//testParsesAs(t, `M117 Hello World`, GcodeLine{})
	testParsesAs(t, `M118 S"Hello_Duet"`, GcodeLine{CmdLetter: 'M', CmdNumber: 118, StringParams: map[uint8]string{'S': `"Hello_Duet"`}})
	testParsesAs(t, `M118 SHello_Duet`, GcodeLine{CmdLetter: 'M', CmdNumber: 118, StringParams: map[uint8]string{'S': `Hello_Duet`}})

	// TODO: test parse failures
}

func testParsesAs(t *testing.T, str string, expected GcodeLine) {
	actual, err := ParseLine(str)
	if err != nil {
		t.Fatal(err)
	}
	if !assert.Equal(t, expected, actual) {
		fmt.Println("passing test (use when updating tests):")
		if actual.Comment != nil {
			fmt.Printf("\tcomment%d := %q\n", actual.Comment, *actual.Comment)
		}
		fmt.Printf("\ttestParsesAs(t, %#q, %v)\n\n", str, actual)
	}
}
