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
type Mesh struct {
	vbo *Buffer
	ibo *Buffer
	modelMat *Mat4
	tmpMat *Mat4
	mtl *Material
	inds int
	// TODO: transformation matrix, etc.
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
	var inds [3]int
	var positions []Vec3
	var texCoords []Vec2
	var mtlFilenames []string

	var verts[]Vertex
	var faces []int32

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
		case "f":
			i1 := len(verts)
			for _, field := range fields[1:] {
				indStrs := strings.Split(field, "/")
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
				if err != nil || inds[0] == 0 || inds[1] == 0 {
					errMsg = "face data error"
					break
				}
				vert := Vertex{positions[inds[0]-1], RGBAColor{}, NewVec2(0, 0)}
				i2 := len(verts) - 1
				i3 := len(verts)
				faces = append(faces, int32(i1), int32(i2), int32(i3))
				verts = append(verts, vert)
			}
		case "mtllib":
			mtlFilenames = mtlFilenames[:0]
			mtlFilenames = append(mtlFilenames, fields[1:]...)
		case "usemtl":
			if len(fields) != 2 {
				errMsg = "material data error"
			} else {
				mtlName := fields[1]
				m.mtl = FindMaterialInFiles(mtlFilenames, mtlName)
				if m.mtl == nil {
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

	m.inds = len(faces)
	m.vbo = NewBuffer()
	m.ibo = NewBuffer()
	m.vbo.SetData(verts, 0)
	m.ibo.SetData(faces, 0)

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
