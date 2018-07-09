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

	var tmp [3]float32
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
			if len(fields[1:]) < 3 {
				errMsg = "vertex data error"
			}
			for i := 0; i < 3; i++ {
				_, err := fmt.Sscan(fields[1+i], &tmp[i])
				if err != nil {
					errMsg = "vertex data error"
					break
				}
			}
			positions = append(positions, NewVec3(tmp[0], tmp[1], tmp[2]))
		case "vt":
			if len(fields[1:]) < 2 || len(fields[1:]) > 3 {
				errMsg = "texture coordinate data error"
			}
			for i := 0; i < 2; i++ {
				_, err := fmt.Sscan(fields[1+i], &tmp[i])
				if err != nil {
					errMsg = "texture coordinate data error"
					break
				}
			}
			texCoords = append(texCoords, NewVec2(tmp[0], tmp[1]))
		case "vn":
			if len(fields[1:]) < 3 {
				errMsg = "vertex normal data error"
			}
			for i := 0; i < 3; i++ {
				_, err := fmt.Sscan(fields[1+i], &tmp[i])
				if err != nil {
					errMsg = "vertex normal data error"
					break
				}
			}
			normals = append(normals, NewVec3(tmp[0], tmp[1], tmp[2]))
		case "f":
			if len(fields[1:]) < 3 {
				errMsg = "face data error"
				break
			}
			i1 := len(verts)
			for _, field := range fields[1:] {
				var ind int
				var vert Vertex
				indStrs := strings.Split(field, "/")
				_, err := fmt.Sscan(indStrs[0], &ind)
				if err != nil {
					panic("face data error")
				}
				if ind > 0 {
					vert.pos = positions[ind-1]
				} else if ind < 0 {
					vert.pos = positions[len(positions)+ind]
				} else {
					panic("face data error")
				}
				if len(indStrs) < 2 || indStrs[1] == "" {
					vert.texCoord = NewVec2(0, 0)
				} else {
					_, err := fmt.Sscan(indStrs[1], &ind)
					if err != nil {
						panic("face data error")
					}
					if ind > 0 {
						vert.texCoord = texCoords[ind-1]
					} else if ind < 0 {
						vert.texCoord = texCoords[len(texCoords)+ind]
					} else {
						panic("face data error")
					}
				}
				if len(indStrs) < 3 || indStrs[2] == "" {
					vert.normal = NewVec3(0, 0, 0)
				} else {
					_, err := fmt.Sscan(indStrs[2], &ind)
					if err != nil {
						panic("face data error")
					}
					if ind > 0 {
						vert.normal = normals[ind-1]
					} else if ind < 0 {
						vert.normal = normals[len(positions)+ind]
					} else {
						panic("face data error")
					}
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
