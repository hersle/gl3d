package graphics

import (
	"github.com/hersle/gl3d/math"
)

var vShaderSrc string = `#version 450
in vec2 position;
out vec2 texCoord;
void main() {
	texCoord = 0.5 * (position - vec2(-1, -1));
	gl_Position = vec4(position, 0, 1);
}
`

var fShaderSrc string = `#version 450
in vec2 texCoord;
uniform sampler2D tex;
out vec4 fragColor;
void main() {
	fragColor = texture(tex, texCoord).rgba;
}
`

var quadProgram *Program
var opts *RenderOptions

func displayTexture(tex *Texture2D, blend Blending) {
	quadProgram.UniformByName("tex").Set(tex)
	opts.Blending = blend
	quadProgram.Render(6, opts)
}

func Clear(rgba math.Vec4) {
	defaultFramebuffer.clearColor(rgba)
}

func init() {
	bl := math.Vec2{-1, -1}
	br := math.Vec2{+1, -1}
	tr := math.Vec2{+1, +1}
	tl := math.Vec2{-1, +1}
	verts := []math.Vec2{bl, br, tr, tl}
	buf := NewVertexBuffer()
	buf.SetData(verts, 0)

	quadProgram = NewProgram(vShaderSrc, fShaderSrc, "")
	quadProgram.InputByName("position").SetSourceVertex(buf, 0)
	quadProgram.framebuffer = defaultFramebuffer

	opts = NewRenderOptions()
	opts.Primitive = TriangleFan
}
