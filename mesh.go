package main

import (
	"os"
	"path"
	"bufio"
	"fmt"
	"errors"
	"strings"
)

type SubMesh struct {
	vbo *Buffer
	ibo *Buffer
	inds int
	mtl *Material
}

type Mesh struct {
	subMeshes []*SubMesh
	modelMat *Mat4
	tmpMat *Mat4
}

func NewSubMesh() *SubMesh {
	var sm SubMesh
	return &sm
}

func (sm *SubMesh) Finish(verts []Vertex, faces []int32) {
	sm.vbo = NewBuffer()
	sm.ibo = NewBuffer()
	sm.vbo.SetData(verts, 0)
	sm.ibo.SetData(faces, 0)
	sm.inds = len(faces)
	if sm.mtl == nil {
		sm.mtl = NewDefaultMaterial("")
	}
}

func ReadMesh(filename string) (*Mesh, error) {
	switch path.Ext(filename) {
	case ".obj":
		return ReadMeshObj(filename)
	default:
		return nil, errors.New(fmt.Sprintf("%s has unknown format", filename))
	}
}

/*
func readVec3(fields []string) (Vec3, error) {
	var xyz [3]float32
	if len(fields) >= 3 {
		for i := 0; i < 3; i++ {
			_, err := fmt.Sscan(fields[i], &xyz[i])
			if err != nil {
				return NewVec3(0, 0, 0), err
			}
		}
		return NewVec3(xyz[0], xyz[1], xyz[2]), nil
	} else {
		return NewVec3(0, 0, 0), errors.New("error reading vec3")
	}
}

func readVec2(fields []string) (Vec2, error) {
	var xy [2]float32
	if len(fields) >= 2 {
		for i := 0; i < 2; i++ {
			_, err := fmt.Sscan(fields[i], &xy[i])
			if err != nil {
				return NewVec2(0, 0), err
			}
		}
		return NewVec2(xy[0], xy[1]), nil
	} else {
		return NewVec2(0, 0), errors.New("error reading vec2")
	}
}

func ReadMeshObj(filename string) (*Mesh, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var m Mesh
	var subMesh *SubMesh = NewSubMesh()
	var verts []Vertex
	var faces []int32

	var positions []Vec3
	var texCoords []Vec2
	var normals []Vec3
	var mtls []*Material

	errMsg := ""
	lineNo := 0
	s := bufio.NewScanner(file)
	for s.Scan() {
		line := s.Text()
		lineNo++
		fields := strings.Fields(line)

		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}

		switch fields[0] {
		case "v":
			position, err := readVec3(fields[1:])
			if err != nil {
				errMsg = "vertex data error"
				break
			}
			positions = append(positions, position)
		case "vt":
			texCoord, err := readVec2(fields[1:])
			if err != nil {
				errMsg = "texture coordinate data error"
				break
			}
			texCoords = append(texCoords, texCoord)
		case "vn":
			normal, err := readVec3(fields[1:])
			if err != nil {
				errMsg = "vertex normal data error"
				break
			}
			normals = append(normals, normal)
		case "f":
			if len(fields[1:]) < 3 {
				errMsg = "face data error"
				break
			}
			i1 := len(verts)
			for _, field := range fields[1:] {
				var inds [3]int
				indStrs := strings.Split(field, "/")
				for i, indStr := range indStrs {
					_, err := fmt.Sscan(indStr, &inds[i])
					if err != nil {
						errMsg = "face data error"
						break
					}
				}
				var vert Vertex
				if inds[0] > 0 {
					vert.pos = positions[inds[0]-1]
				} else if inds[0] < 0 {
					vert.pos = positions[len(positions)+inds[0]]
				} else {
					errMsg = "face data error"
					break
				}
				if inds[1] > 0 {
					vert.texCoord = texCoords[inds[1]-1]
				} else if inds[1] < 0 {
					vert.texCoord = texCoords[len(texCoords)+inds[1]]
				} else {
					vert.texCoord = NewVec2(0, 0)
				}
				if inds[2] > 0 {
					vert.normal = normals[inds[2]-1]
				} else if inds[2] < 0 {
					vert.normal = normals[len(positions)+inds[2]]
				} else {
					vert.normal = NewVec3(0, 0, 0)
				}
				i2 := len(verts) - 1
				i3 := len(verts)
				faces = append(faces, int32(i1), int32(i2), int32(i3))
				verts = append(verts, vert)
			}
		case "mtllib":
			mtlFilenames := fields[1:]
			mtls = ReadMaterials(mtlFilenames)
		case "usemtl":
			if len(faces) > 0 {
				subMesh.Finish(verts, faces)
				m.subMeshes = append(m.subMeshes, subMesh)
			}
			subMesh = NewSubMesh()
			verts = verts[:0]
			faces = faces[:0]

			if len(fields) != 2 {
				errMsg = "material data error"
			} else {
				mtlName := fields[1]
				for _, mtl := range mtls {
					if mtl.name == mtlName {
						subMesh.mtl = mtl
						break
					}
				}
				if subMesh.mtl == nil {
					errMsg = "material not found"
				}
			}
		case "o":
			break // no reason to process
		case "g":
			break // no reason to process
		default:
			println("warning: ignoring line with unknown prefix", fields[0])
		}

		if errMsg != "" {
			err = errors.New(fmt.Sprintf("%s:%d: %s", filename, lineNo, errMsg))
			return nil, err
		}
	}

	if len(faces) > 0 {
		subMesh.Finish(verts, faces)
		m.subMeshes = append(m.subMeshes, subMesh)
	}

	m.modelMat = NewMat4Identity()
	m.tmpMat = NewMat4Zero()

	return &m, nil
}
*/

func readVec2(fields []string) Vec2 {
	var x, y float32
	fmt.Sscan(fields[0], &x)
	fmt.Sscan(fields[1], &y)
	return NewVec2(x, y)
}

func readVec3(fields []string) Vec3 {
	var z float32
	fmt.Sscan(fields[2], &z)
	return readVec2(fields[:2]).Vec3(z)
}

func readIndexedVertex(desc string) indexedVertex {
	var vert indexedVertex
	var inds [3]int = [3]int{0, 0, 0}
	fields := strings.Split(desc, "/")
	for i, field := range fields {
		if field != "" {
			fmt.Sscan(field, &inds[i])
		}
	}
	vert.v, vert.vt, vert.vn = inds[0], inds[1], inds[2]
	return vert
}

type indexedVertex struct {
	v int
	vt int
	vn int
}

type smoothingGroup struct {
	id int
	faces []indexedVertex
}

func newSmoothingGroup(id int) smoothingGroup {
	var sGroup smoothingGroup
	sGroup.id = id
	return sGroup
}

func ReadMeshObj(filename string) (*Mesh, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var positions []Vec3 = []Vec3{NewVec3(0, 0, 0)}
	var texCoords []Vec2 = []Vec2{NewVec2(0, 0)}
	var normals []Vec3 = []Vec3{NewVec3(0, 0, 0)}

	var sGroupInd int = 0
	var sGroups []smoothingGroup = []smoothingGroup{newSmoothingGroup(0)}

	// TODO: materials

	s := bufio.NewScanner(file)
	for s.Scan() {
		line := s.Text()
		fields := strings.Fields(line)

		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}

		switch fields[0] {
		case "v":
			positions = append(positions, readVec3(fields[1:]))
		case "vt":
			texCoords = append(texCoords, readVec2(fields[1:]))
		case "vn":
			normals = append(normals, readVec3(fields[1:]))
		case "s":
			var id int
			if fields[1] == "off" {
				id = 0
			} else {
				fmt.Sscan(fields[1], &id)
			}
			for sGroupInd = 0; sGroupInd < len(sGroups); sGroupInd++ {
				if sGroups[sGroupInd].id == id {
					break
				}
			}
			if sGroupInd == len(sGroups) {
				sGroups = append(sGroups, newSmoothingGroup(id))
			}
		case "f":
			vert1 := readIndexedVertex(fields[1])
			vert2 := readIndexedVertex(fields[2])
			for _, field := range fields[3:] {
				vert3 := readIndexedVertex(field)
				sGroup := &sGroups[sGroupInd]
				sGroup.faces = append(sGroup.faces, vert1, vert2, vert3)
				vert2 = vert3
			}
		case "mtllib":
		case "usemtl":
		default:
			println("ignoring line prefix", fields[0])
		}
	}

	var m Mesh
	var sm *SubMesh = NewSubMesh()
	var verts []Vertex
	var inds []int32

	for _, sGroup := range sGroups {
		var weightedNormals []Vec3 = make([]Vec3, len(positions) + 1)
		for i := 0; i < len(sGroup.faces); i += 3 {
			i1, i2, i3 := i + 0, i + 1, i + 2
			v1, v2, v3 := sGroup.faces[i1].v, sGroup.faces[i2].v, sGroup.faces[i3].v
			edge1 := positions[v3].Sub(positions[v1])
			edge2 := positions[v3].Sub(positions[v2])
			normal := edge1.Cross(edge2).Norm()
			weightedNormals[v1] = weightedNormals[v1].Add(normal)
			weightedNormals[v2] = weightedNormals[v2].Add(normal)
			weightedNormals[v3] = weightedNormals[v3].Add(normal)
		}

		for _, iVert := range sGroup.faces {
			pos := positions[iVert.v]
			texCoord := texCoords[iVert.vt]
			normal := weightedNormals[iVert.v].Norm()
			verts = append(verts, NewVertex(pos, texCoord, normal))
			inds = append(inds, int32(len(verts) - 1))
		}
	}

	sm.Finish(verts, inds)
	sm.mtl.Finish()
	m.subMeshes = append(m.subMeshes, sm)

	m.modelMat = NewMat4Identity()
	m.tmpMat = NewMat4Zero()

	return &m, nil
}

func (m *Mesh) ResetTransformations() {
	m.modelMat.Identity()
}

func (m *Mesh) Translate(d Vec3) {
	m.modelMat.MultLeft(m.tmpMat.Translation(d))
}

func (m *Mesh) Scale(factorX, factorY, factorZ float32) {
	m.modelMat.MultLeft(m.tmpMat.Scaling(factorX, factorY, factorZ))
}

func (m *Mesh) RotateX(ang float32) {
	m.modelMat.MultLeft(m.tmpMat.RotationX(ang))
}

func (m *Mesh) RotateY(ang float32) {
	m.modelMat.MultLeft(m.tmpMat.RotationY(ang))
}

func (m *Mesh) RotateZ(ang float32) {
	m.modelMat.MultLeft(m.tmpMat.RotationZ(ang))
}
