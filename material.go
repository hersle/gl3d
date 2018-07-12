package main

import (
	"os"
	"bufio"
	"strings"
	"fmt"
	"image"
	"image/color"
)

type Material struct {
	name string
	ambient Vec3
	diffuse Vec3
	specular Vec3
	shine float32
	ambientMapFilename string
	ambientMapTexture *Texture2D
	diffuseMapFilename string
	diffuseMapTexture *Texture2D
}

// spec: http://paulbourke.net/dataformats/mtl/

var defaultTexture *Texture2D = nil

func NewDefaultMaterial(name string) *Material {
	var mtl Material
	mtl.name = name
	mtl.ambient = NewVec3(0, 0, 0)
	mtl.diffuse = NewVec3(0, 0, 0)
	mtl.specular = NewVec3(0, 0, 0)
	mtl.shine = 0
	return &mtl
}

func (mtl *Material) Finish() {
	if mtl.ambientMapFilename == "" {
		// use default 1x1 blank texture
		if defaultTexture == nil {
			defaultTexture = NewTexture2D()
			img := image.NewRGBA(image.Rect(0, 0, 1, 1))
			img.Set(0, 0, color.RGBA{0xff, 0xff, 0xff, 0})
			defaultTexture.SetImage(img)
		}
		mtl.ambientMapTexture = defaultTexture
	} else {
		mtl.ambientMapTexture = NewTexture2D()
		err := mtl.ambientMapTexture.ReadImage(mtl.ambientMapFilename)
		if err != nil {
			panic(err)
		}
	}
	if mtl.diffuseMapFilename == "" {
		mtl.diffuseMapTexture = defaultTexture
	} else {
		mtl.diffuseMapTexture = NewTexture2D()
		err := mtl.diffuseMapTexture.ReadImage(mtl.diffuseMapFilename)
		if err != nil {
			panic(err)
		}
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
				mtl.ambient = NewVec3(tmp[0], tmp[1], tmp[2])
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
				mtl.diffuse = NewVec3(tmp[0], tmp[1], tmp[2])
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
				mtl.specular = NewVec3(tmp[0], tmp[1], tmp[2])
			case "Ns":
				if len(fields[1:]) != 1 {
					panic("shine error")
				}
				_, err := fmt.Sscan(fields[1], &mtl.shine)
				if err != nil {
					panic("shine error")
				}
			case "map_Ka":
				if len(fields[1:]) != 1 {
					panic("ambient map error")
				}
				mtl.ambientMapFilename = fields[1]
			case "map_Kd":
				if len(fields[1:]) != 1 {
					panic("diffuse map error")
				}
				mtl.diffuseMapFilename = fields[1]
			}
		}

		if mtl != nil {
			mtl.Finish()
			mtls = append(mtls, mtl)
		}
	}

	return mtls
}
