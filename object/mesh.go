package object

import (
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/material"
	"os"
	"path"
	"bufio"
	"fmt"
	"errors"
	"strings"
	"unsafe"
)

type Vertex struct {
	Position math.Vec3
	TexCoord math.Vec2
	Normal   math.Vec3
	Tangent  math.Vec3
}

type Mesh struct {
	Object
	SubMeshes []*SubMesh
}

type SubMesh struct {
	Verts []Vertex
	Faces []int32
	Vbo *graphics.Buffer
	Ibo *graphics.Buffer
	Inds int
	Mtl *material.Material
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

func NewVertex(position math.Vec3, texCoord math.Vec2, normal, tangent math.Vec3) Vertex {
	var vert Vertex
	vert.Position = position
	vert.TexCoord = texCoord
	vert.Normal = normal
	vert.Tangent = tangent
	return vert
}

func (v *Vertex) Bitangent() math.Vec3 {
	return v.Normal.Cross(v.Tangent)
}

func (_ *Vertex) Size() int {
	return int(unsafe.Sizeof(Vertex{}))
}

func (_ *Vertex) PositionOffset() int {
	return int(unsafe.Offsetof(Vertex{}.Position))
}

func (_ *Vertex) NormalOffset() int {
	return int(unsafe.Offsetof(Vertex{}.Normal))
}

func (_ *Vertex) TexCoordOffset() int {
	return int(unsafe.Offsetof(Vertex{}.TexCoord))
}

func (_ *Vertex) TangentOffset() int {
	return int(unsafe.Offsetof(Vertex{}.Tangent))
}

func NewSubMesh() *SubMesh {
	var sm SubMesh
	return &sm
}

func (sm *SubMesh) AddTriangle(vert1, vert2, vert3 Vertex) {
	i1, i2, i3 := len(sm.Verts) + 0, len(sm.Verts) + 1, len(sm.Verts) + 2
	sm.Faces = append(sm.Faces, int32(i1), int32(i2), int32(i3))
	sm.Verts = append(sm.Verts, vert1, vert2, vert3)
}

func (sm *SubMesh) Finish() {
	sm.Vbo = graphics.NewBuffer()
	sm.Ibo = graphics.NewBuffer()
	sm.Vbo.SetData(sm.Verts, 0)
	sm.Ibo.SetData(sm.Faces, 0)
	sm.Inds = len(sm.Faces)
	if sm.Mtl == nil {
		sm.Mtl = material.NewDefaultMaterial("")
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

func readVec2(fields []string) math.Vec2 {
	var x, y float32
	fmt.Sscan(fields[0], &x)
	fmt.Sscan(fields[1], &y)
	return math.NewVec2(x, y)
}

func readVec3(fields []string) math.Vec3 {
	var x, y, z float32
	fmt.Sscan(fields[0], &x)
	fmt.Sscan(fields[1], &y)
	fmt.Sscan(fields[2], &z)
	return math.NewVec3(x, y, z)
}

func readIndexedVertex(desc string, nv, nvt, nvn int) indexedVertex {
	var vert indexedVertex
	var inds [3]int = [3]int{0, 0, 0}
	fields := strings.Split(desc, "/")
	for i, field := range fields {
		if field != "" {
			fmt.Sscan(field, &inds[i])
		}
	}
	vert.v, vert.vt, vert.vn = inds[0], inds[1], inds[2]
	if vert.v < 0 {
		vert.v = nv + vert.v
	}
	if vert.vt < 0 {
		vert.vt = nvt + vert.vt
	}
	if vert.vn < 0 {
		vert.vn = nvn + vert.vn
	}
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
	m.Object.Reset()
	return m, err
}

func ReadMeshObj(filename string) (*Mesh, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var positions []math.Vec3 = []math.Vec3{math.NewVec3(0, 0, 0)}
	var texCoords []math.Vec2 = []math.Vec2{math.NewVec2(0, 0)}
	var normals []math.Vec3 = []math.Vec3{math.NewVec3(0, 0, 0)}

	var sGroupInd int = 0
	var sGroups []smoothingGroup = []smoothingGroup{newSmoothingGroup(0)}

	var mtlLib []*material.Material
	var mtlInd int = 0
	var mtls []*material.Material = []*material.Material{material.NewDefaultMaterial("")}

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
			nv, nvt, nvn := len(positions), len(texCoords), len(normals)
			vert1 := readIndexedVertex(fields[1], nv, nvt, nvn)
			vert2 := readIndexedVertex(fields[2], nv, nvt, nvn)
			for _, field := range fields[3:] {
				vert3 := readIndexedVertex(field, nv, nvt, nvn)
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
			mtlLib = material.ReadMaterials(fields[1:])
		case "usemtl":
			// find material in current library with given name
			var mtl *material.Material
			for _, mtl = range mtlLib {
				if mtl.Name == fields[1] {
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
	m.SubMeshes = make([]*SubMesh, len(mtls)) // one submesh per material
	for i, _ := range m.SubMeshes {
		m.SubMeshes[i] = NewSubMesh()
		m.SubMeshes[i].Mtl = mtls[i]
	}

	for _, sGroup := range sGroups {
		var weightedNormals []math.Vec3 = make([]math.Vec3, len(positions) + 1)
		var weightedTangents []math.Vec3 = make([]math.Vec3, len(positions) + 1)
		for i := 0; i < len(sGroup.iTris); i++ {
			iTri := sGroup.iTris[i]
			v1, v2, v3 := iTri.iVerts[0].v, iTri.iVerts[1].v, iTri.iVerts[2].v
			edge1 := positions[v1].Sub(positions[v3])
			edge2 := positions[v2].Sub(positions[v3])
			normal := edge1.Cross(edge2).Norm()
			weightedNormals[v1] = weightedNormals[v1].Add(normal)
			weightedNormals[v2] = weightedNormals[v2].Add(normal)
			weightedNormals[v3] = weightedNormals[v3].Add(normal)

			vt1, vt2, vt3 := iTri.iVerts[0].vt, iTri.iVerts[1].vt, iTri.iVerts[2].vt
			dTexCoord1 := texCoords[vt1].Sub(texCoords[vt3])
			dTexCoord2 := texCoords[vt2].Sub(texCoords[vt3])
			det := (dTexCoord1.X() * dTexCoord2.Y() - dTexCoord2.X() * dTexCoord1.Y())
			if det != 0 {
				det = 1 / det
			}
			tangent := math.NewVec3(
				dTexCoord2.Y() * edge1.X() - dTexCoord1.Y() * edge2.X(),
				dTexCoord2.Y() * edge1.Y() - dTexCoord1.Y() * edge2.Y(),
				dTexCoord2.Y() * edge1.Z() - dTexCoord1.Y() * edge2.Z(),
			).Scale(det)
			if det != 0 {
				// tangent is not zero vector
				tangent = tangent.Norm()
			}
			weightedTangents[v1] = weightedTangents[v1].Add(tangent)
			weightedTangents[v2] = weightedTangents[v2].Add(tangent)
			weightedTangents[v3] = weightedTangents[v3].Add(tangent)
		}

		for _, iTri := range sGroup.iTris {
			var verts [3]Vertex
			for i := 0; i < 3; i++ {
				pos := positions[iTri.iVerts[i].v]
				texCoord := texCoords[iTri.iVerts[i].vt]
				normal := weightedNormals[iTri.iVerts[i].v].Norm()
				var tangent math.Vec3
				if tangent.X() == 0 && tangent.Y() == 0 && tangent.Z() == 0 {
					// no tangent - generate arbitrary tangent vector
					// assume normal is not the zero vector
					if normal.X() == 0 && normal.Y() == 0 && normal.Z() == 0 {
						panic("cannot find vector not parallell to zero vector")
					}
					if normal.X() == 0 {
						tangent = math.NewVec3(1, 0, 0)
					} else {
						tangent = math.NewVec3(0, 1, 0)
					}
				} else {
					tangent = weightedTangents[iTri.iVerts[i].v].Norm()
				}
				tangent = tangent.Sub(normal.Scale(tangent.Dot(normal))).Norm() // gram schmidt
				verts[i] = NewVertex(pos, texCoord, normal, tangent)
			}
			m.SubMeshes[iTri.mtlInd].AddTriangle(verts[0], verts[1], verts[2])
		}
	}

	if len(m.SubMeshes[0].Verts) == 0 {
		m.SubMeshes = m.SubMeshes[1:]
	}

	for i, _ := range m.SubMeshes {
		println("submesh", i, "with", len(m.SubMeshes[i].Verts), "verts")
		m.SubMeshes[i].Finish()
		m.SubMeshes[i].Mtl.Finish()
	}

	return &m, nil
}
