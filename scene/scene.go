package scene

import (
	"github.com/hersle/gl3d/light"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/math"
	"image"
	"path"
	"os"
)

type CubeMap struct {
	Posx image.Image
	Negx image.Image
	Posy image.Image
	Negy image.Image
	Posz image.Image
	Negz image.Image
}

type Scene struct {
	Meshes            []*object.Mesh
	AmbientLight      *light.AmbientLight
	SpotLights        []*light.SpotLight
	PointLights       []*light.PointLight
	DirectionalLights []*light.DirectionalLight
	Skybox            *CubeMap
}

func readImage(filename string) (image.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func ReadCubeMapFromDir(dir string) *CubeMap {
	names := []string{"posx.jpg", "negx.jpg", "posy.jpg", "negy.jpg", "posz.jpg", "negz.jpg"}
	var filenames [6]string
	for i, name := range names {
		filenames[i] = path.Join(dir, name)
		println(filenames[i])
	}
	return ReadCubeMap(filenames[0], filenames[1], filenames[2], filenames[3], filenames[4], filenames[5])
}

func ReadCubeMap(filename1, filename2, filename3, filename4, filename5, filename6 string) *CubeMap {
	var imgs [6]image.Image
	var errs [6]error
	imgs[0], errs[0] = readImage(filename1)
	imgs[1], errs[1] = readImage(filename2)
	imgs[2], errs[2] = readImage(filename3)
	imgs[3], errs[3] = readImage(filename4)
	imgs[4], errs[4] = readImage(filename5)
	imgs[5], errs[5] = readImage(filename6)
	for _, err := range errs {
		if err != nil {
			panic(err)
		}
	}
	return NewCubeMap(imgs[0], imgs[1], imgs[2], imgs[3], imgs[4], imgs[5])
}

func NewCubeMap(posx, negx, posy, negy, posz, negz image.Image) *CubeMap {
	var cm CubeMap

	faces := [6]image.Image{posx, negx, posy, negy, posz, negz}
	for _, face := range faces[1:] {
		if !face.Bounds().Size().Eq(faces[0].Bounds().Size()) {
			panic("cube map faces of different size")
		}
	}

	cm.Posx = posx
	cm.Negx = negx
	cm.Posy = posy
	cm.Negy = negy
	cm.Posz = posz
	cm.Negz = negz

	return &cm
}

func NewScene() *Scene {
	var s Scene
	var err error
	if err != nil {
		panic(err)
	}
	s.AmbientLight = light.NewAmbientLight(math.NewVec3(0, 0, 0))
	return &s
}

func (s *Scene) AddMesh(m *object.Mesh) {
	s.Meshes = append(s.Meshes, m)
}

func (s *Scene) AddAmbientLight(l *light.AmbientLight) {
	s.AmbientLight.Color = s.AmbientLight.Color.Add(l.Color)
}

func (s *Scene) AddPointLight(l *light.PointLight) {
	s.PointLights = append(s.PointLights, l)
}

func (s *Scene) AddSpotLight(l *light.SpotLight) {
	s.SpotLights = append(s.SpotLights, l)
}

func (s *Scene) AddDirectionalLight(l *light.DirectionalLight) {
	s.DirectionalLights = append(s.DirectionalLights, l)
}

func (s *Scene) AddSkybox(skybox *CubeMap) {
	s.Skybox = skybox
}
