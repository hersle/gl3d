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
	var subMesh *SubMesh = &SubMesh{}
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
				vert.pos = positions[ind-1]
				if indStrs[1] == "" {
					vert.texCoord = NewVec2(0, 0)
				} else {
					_, err := fmt.Sscan(indStrs[1], &ind)
					if err != nil {
						panic("face data error")
					}
					vert.texCoord = texCoords[ind-1]
				}
				if indStrs[2] == "" {
					vert.normal = NewVec3(0, 0, 0)
				} else {
					_, err := fmt.Sscan(indStrs[2], &ind)
					if err != nil {
						panic("face data error")
					}
					vert.normal = normals[ind-1]
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
				// store submesh
				subMesh.vbo = NewBuffer()
				subMesh.ibo = NewBuffer()
				subMesh.vbo.SetData(verts, 0)
				subMesh.ibo.SetData(faces, 0)
				subMesh.inds = len(faces)
				m.subMeshes = append(m.subMeshes, subMesh)
			}
			// start new submesh
			subMesh = &SubMesh{}
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
		// store submesh
		// TODO: repeated here, make function or only occur once
		subMesh.vbo = NewBuffer()
		subMesh.ibo = NewBuffer()
		subMesh.vbo.SetData(verts, 0)
		subMesh.ibo.SetData(faces, 0)
		subMesh.inds = len(faces)
		m.subMeshes = append(m.subMeshes, subMesh)
		if subMesh.mtl == nil {
			subMesh.mtl = NewDefaultMaterial("")
		}
	}

	m.modelMat = NewMat4Identity()
	m.tmpMat = NewMat4Zero()

	return &m, nil
}

// TODO: implement matrix left multiplication, 
// TODO: so transformations can be done in a natural order

func (m *Mesh) ResetTransformations() {
	m.modelMat.Identity()
}

func (m *Mesh) Translate(d Vec3) {
	m.modelMat.Mult(m.tmpMat.Translation(d))
}

func (m *Mesh) Scale(factorX, factorY, factorZ float32) {
	m.modelMat.Mult(m.tmpMat.Scaling(factorX, factorY, factorZ))
}

func (m *Mesh) RotateX(ang float32) {
	m.modelMat.Mult(m.tmpMat.RotationX(ang))
}

func (m *Mesh) RotateY(ang float32) {
	m.modelMat.Mult(m.tmpMat.RotationY(ang))
}

func (m *Mesh) RotateZ(ang float32) {
	m.modelMat.Mult(m.tmpMat.RotationZ(ang))
}
