package scene

import (
	"github.com/hersle/gl3d/light"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/utils"
	"image"
	"strings"
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

type Node struct {
	mesh *object.Mesh
	ambientLight *light.AmbientLight
	pointLight *light.PointLight
	spotLight *light.SpotLight
	dirLight *light.DirectionalLight
	children []*Node
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

func NewScene() *Node {
	var n Node
	return &n
}

func (n *Node) Traverse(f func(n *Node, depth int)) {
	// depth first search
	var nodeStack []*Node = []*Node{n}
	var depthStack []int = []int{0};

	for len(nodeStack) > 0 {
		// pop
		n1 := nodeStack[len(nodeStack)-1]
		depth := depthStack[len(depthStack)-1]
		nodeStack = nodeStack[:len(nodeStack)-1]
		depthStack = depthStack[:len(depthStack)-1]

		f(n1, depth)

		for _, n2 := range n1.children {
			nodeStack = append(nodeStack, n2)
			depthStack = append(depthStack, depth + 1)
		}
	}
}

func (n *Node) addNode(n2 *Node) {
	n.children = append(n.children, n2)
}

func (n *Node) AddMesh(mesh *object.Mesh) {
	var n2 Node
	n2.mesh = mesh
	n.addNode(&n2)
}

func (n *Node) AddAmbientLight(l *light.AmbientLight) {
	var n2 Node
	n2.ambientLight = l
	n.addNode(&n2)
}

func (n *Node) AddPointLight(l *light.PointLight) {
	var n2 Node
	n2.pointLight = l
	n.addNode(&n2)
}

func (n *Node) AddSpotLight(l *light.SpotLight) {
	var n2 Node
	n2.spotLight = l
	n.addNode(&n2)
}

func (n *Node) AddDirectionalLight(l *light.DirectionalLight) {
	var n2 Node
	n2.dirLight = l
	n.addNode(&n2)
}

func (n *Node) String() string {
	str := ""

	f := func(n *Node, depth int) {
		str += strings.Repeat("  ", depth) + "+ "
		if n.mesh != nil {
			str += "mesh"
		} else if n.ambientLight != nil {
			str += "ambient"
		} else if n.pointLight != nil {
			str += "point"
		} else if n.spotLight != nil {
			str += "spot"
		} else if n.dirLight != nil {
			str += "dir"
		} else {
			str += "node"
		}
		str += "\n"
	}

	n.Traverse(f)

	return str
}
