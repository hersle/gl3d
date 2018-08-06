package material

import (
	"bufio"
	"fmt"
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/math"
	"image"
	"image/color"
	"os"
	"path"
	"strings"
)

type Material struct {
	Name                string
	Ambient             math.Vec3
	Diffuse             math.Vec3
	Specular            math.Vec3
	Shine               float32
	Alpha               float32
	ambientMapFilename  string
	AmbientMap          *graphics.Texture2D
	diffuseMapFilename  string
	DiffuseMap          *graphics.Texture2D
	specularMapFilename string
	SpecularMap         *graphics.Texture2D
	bumpMapFilename     string
	BumpMap             *graphics.Texture2D
	alphaMapFilename    string
	AlphaMap            *graphics.Texture2D
}

// spec: http://paulbourke.net/dataformats/mtl/

var defaultTexture *graphics.Texture2D = nil

func NewDefaultMaterial(name string) *Material {
	var mtl Material
	mtl.Name = name
	mtl.Ambient = math.NewVec3(0.2, 0.2, 0.2)
	mtl.Diffuse = math.NewVec3(0.8, 0.8, 0.8)
	mtl.Specular = math.NewVec3(0, 0, 0)
	mtl.Shine = 1
	mtl.Alpha = 1
	return &mtl
}

func initDefaultTexture() {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{0xff, 0xff, 0xff, 0})
	defaultTexture = graphics.NewTexture2DFromImage(gl.NEAREST, gl.REPEAT, gl.RGBA8, img)
}

func (mtl *Material) HasBumpMap() bool {
	return mtl.BumpMap != nil
}

func (mtl *Material) HasAlphaMap() bool {
	return mtl.AlphaMap != nil
}

func (mtl *Material) Finish() {
	var err error

	if defaultTexture == nil {
		initDefaultTexture()
	}

	if mtl.ambientMapFilename != "" {
		mtl.AmbientMap, err = graphics.ReadTexture2D(gl.LINEAR, gl.REPEAT, gl.RGBA8, mtl.ambientMapFilename)
		if err != nil {
			mtl.AmbientMap = defaultTexture
		}
	} else {
		mtl.AmbientMap = defaultTexture
	}

	if mtl.diffuseMapFilename != "" {
		mtl.DiffuseMap, err = graphics.ReadTexture2D(gl.LINEAR, gl.REPEAT, gl.RGBA8, mtl.diffuseMapFilename)
		if err != nil {
			mtl.DiffuseMap = defaultTexture
		}
	} else {
		mtl.DiffuseMap = defaultTexture
	}

	if mtl.specularMapFilename != "" {
		mtl.SpecularMap, err = graphics.ReadTexture2D(gl.LINEAR, gl.REPEAT, gl.RGBA8, mtl.specularMapFilename)
		if err != nil {
			mtl.SpecularMap = defaultTexture
		}
	} else {
		mtl.SpecularMap = defaultTexture
	}

	if mtl.bumpMapFilename != "" {
		mtl.BumpMap, err = graphics.ReadTexture2D(gl.LINEAR, gl.REPEAT, gl.RGB8, mtl.bumpMapFilename)
	}

	if mtl.alphaMapFilename != "" {
		mtl.AlphaMap, err = graphics.ReadTexture2D(gl.LINEAR, gl.REPEAT, gl.R8, mtl.alphaMapFilename)
	}
}

func ReadMaterials(filenames []string) []*Material {
	var mtls []*Material
	var mtl *Material

	var tmp [3]float32

	for _, filename := range filenames {
		mtl = nil

		file, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		s := bufio.NewScanner(file)
		for s.Scan() {
			line := s.Text()
			fields := strings.Fields(line)

			if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
				continue // comment
			}
			if mtl == nil && fields[0] != "newmtl" {
				panic("newmtl was not first statement in .mtl file")
			}

			switch fields[0] {
			case "newmtl":
				if len(fields[1:]) != 1 {
					panic("newmtl error")
				}
				if mtl != nil {
					mtl.Finish()
					mtls = append(mtls, mtl)
				}
				name := fields[1]
				mtl = NewDefaultMaterial(name)
			case "Ka":
				if len(fields[1:]) < 3 {
					panic("Ka error")
				}
				for i := 0; i < 3; i++ {
					_, err := fmt.Sscan(fields[1+i], &tmp[i])
					if err != nil {
						panic("Ka error")
					}
				}
				mtl.Ambient = math.NewVec3(tmp[0], tmp[1], tmp[2])
			case "Kd":
				if len(fields[1:]) < 3 {
					panic("Kd error")
				}
				for i := 0; i < 3; i++ {
					_, err := fmt.Sscan(fields[1+i], &tmp[i])
					if err != nil {
						panic("Kd error")
					}
				}
				mtl.Diffuse = math.NewVec3(tmp[0], tmp[1], tmp[2])
			case "Ks":
				if len(fields[1:]) < 3 {
					panic("Ks error")
				}
				for i := 0; i < 3; i++ {
					_, err := fmt.Sscan(fields[1+i], &tmp[i])
					if err != nil {
						panic("Ks error")
					}
				}
				mtl.Specular = math.NewVec3(tmp[0], tmp[1], tmp[2])
			case "Ns":
				if len(fields[1:]) != 1 {
					panic("shine error")
				}
				_, err := fmt.Sscan(fields[1], &mtl.Shine)
				if err != nil {
					panic("shine error")
				}
			case "map_Ka":
				if len(fields[1:]) != 1 {
					panic("ambient map error")
				}
				if path.IsAbs(fields[1]) {
					mtl.ambientMapFilename = fields[1]
				} else {
					mtl.ambientMapFilename = path.Join(path.Dir(filename), fields[1])
				}
			case "map_Kd":
				if len(fields[1:]) != 1 {
					panic("diffuse map error")
				}
				if path.IsAbs(fields[1]) {
					mtl.diffuseMapFilename = fields[1]
				} else {
					mtl.diffuseMapFilename = path.Join(path.Dir(filename), fields[1])
				}
			case "map_Ks":
				if len(fields[1:]) != 1 {
					panic("specular map error")
				}
				if path.IsAbs(fields[1]) {
					mtl.specularMapFilename = fields[1]
				} else {
					mtl.specularMapFilename = path.Join(path.Dir(filename), fields[1])
				}
			case "bump":
				if len(fields[1:]) != 1 {
					panic("bump map error")
				}
				if path.IsAbs(fields[1]) {
					mtl.bumpMapFilename = fields[1]
				} else {
					mtl.bumpMapFilename = path.Join(path.Dir(filename), fields[1])
				}
			case "map_d":
				if len(fields[1:]) != 1 {
					panic("alpha map error")
				}
				if path.IsAbs(fields[1]) {
					mtl.alphaMapFilename = fields[1]
				} else {
					mtl.alphaMapFilename = path.Join(path.Dir(filename), fields[1])
				}
			case "d":
				if len(fields[1:]) != 1 {
					panic("dissolve error")
				}
				_, err = fmt.Sscan(fields[1], &mtl.Alpha)
				if err != nil {
					panic(err)
				}
			default:
				println("ignored material file prefix", fields[0])
			}
		}

		if mtl != nil {
			mtl.Finish()
			mtls = append(mtls, mtl)
		}
	}

	return mtls
}