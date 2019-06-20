package scene

import (
	"github.com/hersle/gl3d/light"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/utils"
	"image"
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

func ReadCubeMap(filename1, filename2, filename3, filename4, filename5, filename6 string) (*CubeMap, error) {
	imgs, err := utils.ReadImages(filename1, filename2, filename3, filename4, filename5, filename6)
	if err != nil {
		return nil, err
	}
	return NewCubeMap(imgs[0], imgs[1], imgs[2], imgs[3], imgs[4], imgs[5]), nil
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
	s.AmbientLight = light.NewAmbientLight(math.Vec3{0, 0, 0})
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
