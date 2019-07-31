package graphics

import (
	_ "github.com/hersle/gl3d/window" // initialize graphics
	"github.com/go-gl/gl/v4.5-core/gl"
	"unsafe"
	"fmt"
	"os"
)

func debugcb(_, _, _, severity uint32, _ int32, message string, _ unsafe.Pointer) {
	var severitystr string
	switch (severity) {
	case gl.DEBUG_SEVERITY_HIGH:
		severitystr = "high severity"
	case gl.DEBUG_SEVERITY_MEDIUM:
		severitystr = "medium severity"
	case gl.DEBUG_SEVERITY_LOW:
		severitystr = "low severity"
	case gl.DEBUG_SEVERITY_NOTIFICATION:
		severitystr = "notification"
	}

	fmt.Fprintf(os.Stderr, "OpenGL debug message (%s): %s\n", severitystr, message)
}

func init() {
	gl.Enable(gl.DEBUG_OUTPUT)
	gl.Enable(gl.DEBUG_OUTPUT_SYNCHRONOUS)

	// disable notification debug messages
	gl.DebugMessageControl(gl.DONT_CARE, gl.DONT_CARE, gl.DEBUG_SEVERITY_NOTIFICATION, 0, nil, false)

	gl.DebugMessageCallback(debugcb, nil)
}
