package gcodetools

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGcodeBuilder(t *testing.T) {
	builder := GcodeBuilder{
		NozzleSize:          0.4,
		LayerHeight:         0.2,
		ExtrusionWidth:      0.4,
		ExtrusionMultiplier: 1,
		FilamentDiameter:    1.75,
		TravelFeedrate:      60 * 100,
		PrintFeedrate:       60 * 40,
	}

	builder.Comment("; unit test gcode")
	builder.Home()
	builder.RelativeExtrusion()
	builder.TravelToXY(100, 100)
	builder.PrintToXY(100, 200)
	builder.PrintToXY(200, 200)
	builder.PrintToXY(200, 100)
	builder.PrintToXY(100, 100)
	builder.TravelTo(100, 100, 20)
	builder.Home()

	outputGcodeStr := builder.ToString()

	expected := `; unit test gcode
G28
M83
G0 X100 Y100 F6000
G1 Y200 E3.3260135 F2400
G1 X200 E3.3260135
G1 Y100 E3.3260135
G1 X100 E3.3260135
G0 Z20 F6000
G28
`

	assert.Equal(t, expected, outputGcodeStr)
}
