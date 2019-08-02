package material

import (
	"testing"
	"github.com/hersle/gl3d/utils"
)

func TestBumpToNormal(t *testing.T) {
	//in := "../assets/objects/sponza/textures/lion2_bump.png"
	in := "../assets/objects/sponza/textures/spnza_bricks_a_bump.png"
	out := "normal.png"
	bump, err := utils.ReadImage(in)
	if err != nil {
		panic(err)
	}
	normal := bumpMapToNormalMap(bump)
	err = utils.WriteImagePNG(normal, out)
	if err != nil {
		panic(err)
	}
	err = utils.WriteImagePNG(bump, "bump.png")
	if err != nil {
		panic(err)
	}
}
