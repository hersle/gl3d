package main

import (
	"os"
	"path"
	"bufio"
	"fmt"
	"errors"
	"strings"
)

type Mesh struct {
	Object
	subMeshes []*SubMesh
}

type SubMesh struct {
	verts []Vertex
	faces []int32
	vbo *Buffer
	ibo *Buffer
	inds int
	mtl *Material
}

type indexedVertex struct {
	v int
	vt int
	vn int
}

type indexedTriangle struct {
	iVerts [3]indexedVertex
	mtlInd int
}

type smoothingGroup struct {
	id int
	iTris []indexedTriangle
}

func NewSubMesh() *SubMesh {
	var sm SubMesh
	return &sm
}

func (sm *SubMesh) AddTriangle(vert1, vert2, vert3 Vertex) {
	i1, i2, i3 := len(sm.verts) + 0, len(sm.verts) + 1, len(sm.verts) + 2
	sm.faces = append(sm.faces, int32(i1), int32(i2), int32(i3))
	sm.verts = append(sm.verts, vert1, vert2, vert3)
}

func (sm *SubMesh) Finish() {
	sm.vbo = NewBuffer()
	sm.ibo = NewBuffer()
	sm.vbo.SetData(sm.verts, 0)
	sm.ibo.SetData(sm.faces, 0)
	sm.inds = len(sm.faces)
	if sm.mtl == nil {
		sm.mtl = NewDefaultMaterial("")
	}
}

func newIndexedTriangle(iv1, iv2, iv3 indexedVertex, mtlInd int) indexedTriangle {
	var iTri indexedTriangle
	iTri.iVerts[0], iTri.iVerts[1], iTri.iVerts[2] = iv1, iv2, iv3
	iTri.mtlInd = mtlInd
	return iTri
}

func newSmoothingGroup(id int) smoothingGroup {
	var sGroup smoothingGroup
	sGroup.id = id
	return sGroup
}

func readVec2(fields []string) Vec2 {
	var x, y float32
	fmt.Sscan(fields[0], &x)
	fmt.Sscan(fields[1], &y)
	return NewVec2(x, y)
}

func readVec3(fields []string) Vec3 {
	var x, y, z float32
	fmt.Sscan(fields[0], &x)
	fmt.Sscan(fields[1], &y)
	fmt.Sscan(fields[2], &z)
	return NewVec3(x, y, z)
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

func ReadMesh(filename string) (*Mesh, error) {
	var m *Mesh
	var err error
	switch path.Ext(filename) {
	case ".obj":
		m, err = ReadMeshObj(filename)
	default:
		return nil, errors.New(fmt.Sprintf("%s has unknown format", filename))
	}
	m.Object.Init()
	return m, err
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

	var mtlLib []*Material
	var mtlInd int = 0
	var mtls []*Material = []*Material{NewDefaultMaterial("")}

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
				iTri := newIndexedTriangle(vert1, vert2, vert3, mtlInd)
				sGroup.iTris = append(sGroup.iTris, iTri)
				vert2 = vert3
			}
		case "mtllib":
			for i, _ := range fields[1:] {
				if !path.IsAbs(fields[1:][i]) {
					fields[1:][i] = path.Join(path.Dir(filename), fields[1:][i])
				}
			}
			mtlLib = ReadMaterials(fields[1:])
		case "usemtl":
			// find material in current library with given name
			var mtl *Material
			for _, mtl = range mtlLib {
				if mtl.name == fields[1] {
					break
				}
			}

			// if the material has been used before, use it again
			for mtlInd = 0; mtlInd < len(mtls); mtlInd++ {
				if mtls[mtlInd] == mtl {
					break
				}
			}

			// otherwise, make a new material
			if mtlInd == len(mtls) {
				mtls = append(mtls, mtl)
			}
		case "g", "o":
			continue // ignore without warning - no effect on appearance
		default:
			println("ignoring line prefix", fields[0])
		}
	}

	var m Mesh
	m.subMeshes = make([]*SubMesh, len(mtls)) // one submesh per material
	for i, _ := range m.subMeshes {
		m.subMeshes[i] = NewSubMesh()
		m.subMeshes[i].mtl = mtls[i]
	}

	for _, sGroup := range sGroups {
		var weightedNormals []Vec3 = make([]Vec3, len(positions) + 1)
		for i := 0; i < len(sGroup.iTris); i++ {
			iTri := sGroup.iTris[i]
			v1, v2, v3 := iTri.iVerts[0].v, iTri.iVerts[1].v, iTri.iVerts[2].v
			edge1 := positions[v3].Sub(positions[v1])
			edge2 := positions[v3].Sub(positions[v2])
			normal := edge1.Cross(edge2).Norm()
			weightedNormals[v1] = weightedNormals[v1].Add(normal)
			weightedNormals[v2] = weightedNormals[v2].Add(normal)
			weightedNormals[v3] = weightedNormals[v3].Add(normal)
		}

		for _, iTri := range sGroup.iTris {
			var verts [3]Vertex
			for i := 0; i < 3; i++ {
				pos := positions[iTri.iVerts[i].v]
				texCoord := texCoords[iTri.iVerts[i].vt]
				normal := weightedNormals[iTri.iVerts[i].v].Norm()
				verts[i] = NewVertex(pos, texCoord, normal)
			}
			m.subMeshes[iTri.mtlInd].AddTriangle(verts[0], verts[1], verts[2])
		}
	}

	if len(m.subMeshes[0].verts) == 0 {
		m.subMeshes = m.subMeshes[1:]
	}

	for i, _ := range m.subMeshes {
		println("submesh", i, "with", len(m.subMeshes[i].verts), "verts")
		m.subMeshes[i].Finish()
		m.subMeshes[i].mtl.Finish()
	}

	return &m, nil
}

func (m *Mesh) Reset() {
	m.Object.Reset()
}
