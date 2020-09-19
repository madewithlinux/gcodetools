package gcodetools

import "github.com/go-gl/mathgl/mgl64"

type vec3 mgl64.Vec3

type GcodeVirtualMachine struct {
	Position vec3
	Feedrate float64
}
