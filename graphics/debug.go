package graphics

import (
	_ "github.com/hersle/gl3d/window" // initialize graphics
	"github.com/go-gl/gl/v4.5-core/gl"
	"unsafe"
)

func debugcb(source uint32, gltype uint32, id uint32, severity uint32, length int32, message string, userParam unsafe.Pointer) {
	var severitystr string
	switch (severity) {
	case gl.DEBUG_SEVERITY_HIGH:
		severitystr = "HIGH"
	case gl.DEBUG_SEVERITY_MEDIUM:
		severitystr = "MEDIUM"
	case gl.DEBUG_SEVERITY_LOW:
		severitystr = "LOW"
	case gl.DEBUG_SEVERITY_NOTIFICATION:
		severitystr = "NOTIFY"
	}
	print("OPENGL DEBUG MESSAGE (")
	print(severitystr)
	print("): ")
	println(message)
}

func init() {
	gl.Enable(gl.DEBUG_OUTPUT)
	gl.Enable(gl.DEBUG_OUTPUT_SYNCHRONOUS)
	gl.DebugMessageControl(gl.DONT_CARE, gl.DONT_CARE, gl.DONT_CARE, 0, nil, true)
	gl.DebugMessageCallback(debugcb, nil)
}
