package main

import (
	"os"
	"path"
	"bufio"
	"fmt"
	"errors"
	"strings"
	"strconv"
)

// TODO: upload mesh data to GPU only once!
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

	var pos [3]float32
	var texCoord [2]float32
	var normal [3]float32
	var inds [3]int
	var positions []Vec3
	var texCoords []Vec2
	var normals []Vec3
	var mtlFilenames []string

	var verts[]Vertex
	var faces []int32

	var subMesh *SubMesh

	errMsg := ""
	lineNo := 0
	s := bufio.NewScanner(file)
	for s.Scan() {
		line := s.Text()
		lineNo++
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "#":
			continue
		case "v":
			if len(fields[1:]) < 3 {
				errMsg = "vertex data error"
			}
			for i := 0; i < 3; i++ {
				_, err := fmt.Sscan(fields[1+i], &pos[i])
				if err != nil {
					errMsg = "vertex data error"
					break
				}
			}
			positions = append(positions, NewVec3(pos[0], pos[1], pos[2]))
		case "vt":
			if len(fields[1:]) < 2 || len(fields[1:]) > 3 {
				errMsg = "texture coordinate data error"
			}
			for i := 0; i < 2; i++ {
				_, err := fmt.Sscan(fields[1+i], &texCoord[i])
				if err != nil {
					errMsg = "texture coordinate data error"
					break
				}
			}
			texCoords = append(texCoords, NewVec2(texCoord[0], texCoord[1]))
		case "vn":
			if len(fields[1:]) < 3 {
				errMsg = "vertex normal data error"
			}
			for i := 0; i < 3; i++ {
				_, err := fmt.Sscan(fields[1+i], &normal[i])
				if err != nil {
					errMsg = "vertex normal data error"
					break
				}
			}
			normals = append(normals, NewVec3(normal[0], normal[1], normal[2]))
		case "f":
			i1 := len(verts)
			for _, field := range fields[1:] {
				indStrs := strings.Split(field, "/")
				inds[0], inds[1], inds[2] = 0, 0, 0
				for i, indStr := range indStrs {
					if indStr == "" {
						inds[i] = 0
					} else {
						inds[i], err = strconv.Atoi(indStr)
						if err != nil {
							break
						}
					}
				}
				if err != nil || inds[0] == 0 {
					errMsg = "face data error"
					break
				}
				p := positions[inds[0]-1]
				c := RGBAColor{}
				t := NewVec2(0, 0)
				n := NewVec3(0, 0, 0)
				if inds[1] != 0 {
					t = texCoords[inds[1]-1]
				}
				if inds[2] != 0 {
					n = normals[inds[2]-1]
				}
				vert := Vertex{p, c, t, n}
				fmt.Printf("%+v\n", vert)
				i2 := len(verts) - 1
				i3 := len(verts)
				faces = append(faces, int32(i1), int32(i2), int32(i3))
				verts = append(verts, vert)
			}
		case "mtllib":
			mtlFilenames = mtlFilenames[:0]
			mtlFilenames = append(mtlFilenames, fields[1:]...)
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
				subMesh.mtl = FindMaterialInFiles(mtlFilenames, mtlName)
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
		subMesh.vbo = NewBuffer()
		subMesh.ibo = NewBuffer()
		subMesh.vbo.SetData(verts, 0)
		subMesh.ibo.SetData(faces, 0)
		subMesh.inds = len(faces)
		m.subMeshes = append(m.subMeshes, subMesh)
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
