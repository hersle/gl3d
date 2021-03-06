package material

import (
	"bufio"
	"fmt"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/utils"
	"image"
	"image/color"
	"os"
	"path"
	"strings"
	"log"
)

type Material struct {
	Name        string
	Ambient     math.Vec3
	Diffuse     math.Vec3
	Specular    math.Vec3
	Shine       float32
	Alpha       float32
	AmbientMap  image.Image
	DiffuseMap  image.Image
	SpecularMap image.Image
	BumpMap     image.Image
	AlphaMap    image.Image
}

// spec: http://paulbourke.net/dataformats/mtl/

var whiteTransparentTexture image.Image
var defaultNormalTexture image.Image

func NewDefaultMaterial(name string) *Material {
	var mtl Material
	mtl.Name = name
	mtl.Ambient = math.Vec3{0.2, 0.2, 0.2}
	mtl.Diffuse = math.Vec3{0.8, 0.8, 0.8}
	mtl.Specular = math.Vec3{0, 0, 0}
	mtl.Shine = 1
	mtl.Alpha = 1
	mtl.AmbientMap = whiteTransparentTexture
	mtl.DiffuseMap = whiteTransparentTexture
	mtl.SpecularMap = whiteTransparentTexture
	mtl.AlphaMap = whiteTransparentTexture
	mtl.BumpMap = defaultNormalTexture
	return &mtl
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
				mtl.Ambient = math.Vec3{tmp[0], tmp[1], tmp[2]}
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
				mtl.Diffuse = math.Vec3{tmp[0], tmp[1], tmp[2]}
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
				mtl.Specular = math.Vec3{tmp[0], tmp[1], tmp[2]}
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
				ambientMapFilename := fields[1]
				if !path.IsAbs(ambientMapFilename) {
					ambientMapFilename = path.Join(path.Dir(filename), ambientMapFilename)
				}
				ambientMap, err := utils.ReadImage(ambientMapFilename)
				if err == nil {
					mtl.AmbientMap = ambientMap
				}
			case "map_Kd":
				if len(fields[1:]) != 1 {
					panic("diffuse map error")
				}
				diffuseMapFilename := fields[1]
				if !path.IsAbs(diffuseMapFilename) {
					diffuseMapFilename = path.Join(path.Dir(filename), diffuseMapFilename)
				}
				diffuseMap, err := utils.ReadImage(diffuseMapFilename)
				if err == nil {
					mtl.DiffuseMap = diffuseMap
				}
			case "map_Ks":
				if len(fields[1:]) != 1 {
					panic("specular map error")
				}
				specularMapFilename := fields[1]
				if !path.IsAbs(specularMapFilename) {
					specularMapFilename = path.Join(path.Dir(filename), specularMapFilename)
				}
				specularMap, err := utils.ReadImage(specularMapFilename)
				if err == nil {
					mtl.SpecularMap = specularMap
				}
			case "bump", "map_bump":
				if len(fields[1:]) != 1 {
					panic("bump map error")
				}
				bumpMapFilename := fields[1]
				if !path.IsAbs(bumpMapFilename) {
					bumpMapFilename = path.Join(path.Dir(filename), bumpMapFilename)
				}
				bumpMap, err := utils.ReadImage(bumpMapFilename)
				if err == nil {
					mtl.BumpMap = bumpMap
				}
			case "map_d":
				if len(fields[1:]) != 1 {
					panic("alpha map error")
				}
				alphaMapFilename := fields[1]
				if !path.IsAbs(alphaMapFilename) {
					alphaMapFilename = path.Join(path.Dir(filename), alphaMapFilename)
				}
				alphaMap, err := utils.ReadImage(alphaMapFilename)
				if err == nil {
					mtl.AlphaMap = alphaMap
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
				log.Print("ignored material line prefix ", fields[0])
			}
		}

		if mtl != nil {
			mtls = append(mtls, mtl)
		}
	}

	return mtls
}

func bumpMapToNormalMap(bump image.Image) image.Image {
	normal := image.NewRGBA(bump.Bounds().Inset(1))
	for y0 := normal.Bounds().Min.Y; y0 < normal.Bounds().Max.Y - 1; y0++ {
		for x0 := normal.Bounds().Min.X; x0 < normal.Bounds().Max.X - 1; x0++ {
			z1, _, _, _ := bump.At(x0-1, y0).RGBA()
			z2, _, _, _ := bump.At(x0+1, y0).RGBA()
			dzdx := float32(z2-z1) * 0.000001
			z1, _, _, _ = bump.At(x0, y0-1).RGBA()
			z2, _, _, _ = bump.At(x0, y0+1).RGBA()
			dzdy := float32(z2-z1) * 0.000001
			n := math.Vec3{dzdx, dzdy, 2}.Norm()
			n = n.Sub(math.Vec3{-1, -1, -1}).Scale(0.5)
			r := uint8(float32(0xff) * n.X())
			g := uint8(float32(0xff) * n.Y())
			b := uint8(float32(0xff) * n.Z())
			a := uint8(0xff)
			rgba := color.RGBA{r, g, b, a}
			normal.SetRGBA(x0, y0, rgba)
		}
	}
	return normal
}

func init() {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.SetRGBA(0, 0, color.RGBA{255, 255, 255, 0})
	whiteTransparentTexture = img

	img = image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.SetRGBA(0, 0, color.RGBA{0x80, 0x80, 0xff, 0})
	defaultNormalTexture = img
}
