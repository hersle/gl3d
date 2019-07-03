package object

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/hersle/gl3d/material"
	"github.com/hersle/gl3d/math"
	"os"
	"path"
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

type Geometry struct {
	Verts    []Vertex
	Faces    []int32
	Inds     int
	uploaded bool
}

type SubMesh struct {
	Mesh *Mesh
	bbox *Box
	Geo  *Geometry
	Mtl  *material.Material
}

type indexedVertex struct {
	v  int
	vt int
	vn int
}

type indexedTriangle struct {
	iVerts [3]indexedVertex
	mtlInd int
}

type smoothingGroup struct {
	id    int
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

func (m *Mesh) AddSubMesh(sm *SubMesh) {
	m.SubMeshes = append(m.SubMeshes, sm)
}

func (m *Mesh) invalidate() {
	for _, sm := range m.SubMeshes {
		sm.bbox = nil
	}
}

func (m *Mesh) Orient(unitX, unitY math.Vec3) {
	m.Object.Orient(unitX, unitY)
	m.invalidate()
}

func (m *Mesh) Place(position math.Vec3) {
	m.Object.Place(position)
	m.invalidate()
}

func (m *Mesh) Rotate(axis math.Vec3, ang float32) {
	m.Object.Rotate(axis, ang)
	m.invalidate()
}

func (m *Mesh) RotateX(ang float32) {
	m.Object.RotateX(ang)
	m.invalidate()
}

func (m *Mesh) RotateY(ang float32) {
	m.Object.RotateY(ang)
	m.invalidate()
}

func (m *Mesh) RotateZ(ang float32) {
	m.Object.RotateZ(ang)
	m.invalidate()
}

func (m *Mesh) Scale(factor math.Vec3) {
	m.Object.Scale(factor)
	m.invalidate()
}

func (m *Mesh) SetScale(scaling math.Vec3) {
	m.Object.SetScale(scaling)
	m.invalidate()
}

func NewSubMesh(geo *Geometry, mtl *material.Material, mesh *Mesh) *SubMesh {
	var sm SubMesh
	sm.Mesh = mesh
	sm.Geo = geo
	sm.Mtl = mtl
	return &sm
}

func (geo *Geometry) AddTriangle(vert1, vert2, vert3 Vertex) {
	i1, i2, i3 := len(geo.Verts)+0, len(geo.Verts)+1, len(geo.Verts)+2
	geo.Faces = append(geo.Faces, int32(i1), int32(i2), int32(i3))
	geo.Verts = append(geo.Verts, vert1, vert2, vert3)
	geo.Inds += 3
	geo.uploaded = false
}

func (sm *SubMesh) BoundingBox() *Box {
	if sm.Geo == nil || sm.Geo.Inds == 0 {
		return nil
	}

	if sm.bbox == nil {
		worldMatrix := sm.Mesh.WorldMatrix()
		pos := sm.Geo.Verts[0].Position.Vec4(1).Transform(worldMatrix).Vec3()
		minX := pos.X()
		minY := pos.Y()
		minZ := pos.Z()
		maxX := minX
		maxY := minY
		maxZ := minZ
		for _, v := range sm.Geo.Verts[1:] {
			pos := v.Position.Vec4(1).Transform(worldMatrix).Vec3()
			minX = math.Min(minX, pos.X())
			minY = math.Min(minY, pos.Y())
			minZ = math.Min(minZ, pos.Z())
			maxX = math.Max(maxX, pos.X())
			maxY = math.Max(maxY, pos.Y())
			maxZ = math.Max(maxZ, pos.Z())
		}
		sm.bbox = NewBoxAxisAligned(math.Vec3{minX, minY, minZ}, math.Vec3{maxX, maxY, maxZ})
	}

	return sm.bbox
}

func (sm *SubMesh) BoundingSphere() *Sphere {
	bbox := sm.BoundingBox()
	return NewSphere(bbox.Center(), bbox.DiagonalLength())
}

func (geo *Geometry) CalculateNormals() {
	for i, _ := range geo.Verts {
		geo.Verts[i].Normal = math.Vec3{0, 0, 0}
	}

	for i := 0; i < geo.Inds; i += 3 {
		v1 := geo.Verts[geo.Faces[i+0]]
		v2 := geo.Verts[geo.Faces[i+1]]
		v3 := geo.Verts[geo.Faces[i+2]]

		edge1 := v1.Position.Sub(v3.Position)
		edge2 := v2.Position.Sub(v3.Position)
		normal := edge1.Cross(edge2).Norm()
		v1.Normal = v1.Normal.Add(normal)
		v2.Normal = v2.Normal.Add(normal)
		v3.Normal = v3.Normal.Add(normal)

		geo.Verts[geo.Faces[i+0]] = v1
		geo.Verts[geo.Faces[i+1]] = v2
		geo.Verts[geo.Faces[i+2]] = v3
	}

	for i, v := range geo.Verts {
		geo.Verts[i].Normal = v.Normal.Norm()
	}

	geo.uploaded = false
}

func (geo *Geometry) CalculateTangents() {
	for i, _ := range geo.Verts {
		geo.Verts[i].Tangent = math.Vec3{0, 0, 0}
	}

	for i := 0; i < geo.Inds; i += 3 {
		v1 := geo.Verts[geo.Faces[i+0]]
		v2 := geo.Verts[geo.Faces[i+1]]
		v3 := geo.Verts[geo.Faces[i+2]]

		edge1 := v1.Position.Sub(v3.Position)
		edge2 := v2.Position.Sub(v3.Position)
		dTexCoord1 := v1.TexCoord.Sub(v3.TexCoord)
		dTexCoord2 := v2.TexCoord.Sub(v3.TexCoord)
		det := dTexCoord1.X()*dTexCoord2.Y() - dTexCoord2.X()*dTexCoord1.Y()
		if det != 0 {
			det = 1 / det
		}
		tangent := math.Vec3{
			dTexCoord2.Y()*edge1.X() - dTexCoord1.Y()*edge2.X(),
			dTexCoord2.Y()*edge1.Y() - dTexCoord1.Y()*edge2.Y(),
			dTexCoord2.Y()*edge1.Z() - dTexCoord1.Y()*edge2.Z()}.
			Scale(det)
		if det != 0 {
			// tangent is not zero vector
			tangent = tangent.Norm()
		}
		v1.Tangent = v1.Tangent.Add(tangent)
		v2.Tangent = v2.Tangent.Add(tangent)
		v3.Tangent = v3.Tangent.Add(tangent)

		geo.Verts[geo.Faces[i+0]] = v1
		geo.Verts[geo.Faces[i+1]] = v2
		geo.Verts[geo.Faces[i+2]] = v3
	}

	for i, v := range geo.Verts {
		if v.Tangent.Dot(v.Tangent) == 0 {
			// no tangent - generate arbitrary tangent vector
			// assume normal is not the zero vector
			if v.Normal.Dot(v.Normal) == 0 {
				panic("cannot find vector not parallell to zero vector")
			}
			if v.Normal.X() == 0 {
				v.Tangent = math.Vec3{1, 0, 0}
			} else {
				v.Tangent = math.Vec3{0, 1, 0}
			}
		} else {
			v.Tangent = v.Tangent.Norm()
		}
		v.Tangent = v.Tangent.Sub(v.Normal.Scale(v.Tangent.Dot(v.Normal))).Norm() // gram schmidt
		geo.Verts[i] = v
	}

	geo.uploaded = false
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
	return math.Vec2{x, y}
}

func readVec3(fields []string) math.Vec3 {
	var x, y, z float32
	fmt.Sscan(fields[0], &x)
	fmt.Sscan(fields[1], &y)
	fmt.Sscan(fields[2], &z)
	return math.Vec3{x, y, z}
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

func NewMesh(geo *Geometry, mtl *material.Material) *Mesh {
	if geo == nil && mtl != nil {
		panic("created mesh with material and without geometry")
	}

	var mesh Mesh
	mesh.Object = *NewObject()
	if geo != nil {
		if mtl == nil {
			mtl = material.NewDefaultMaterial("")
		}
		mesh.AddSubMesh(NewSubMesh(geo, mtl, &mesh))
	}
	return &mesh
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
	return m, err
}

func ReadMeshObj(filename string) (*Mesh, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var positions []math.Vec3 = []math.Vec3{math.Vec3{0, 0, 0}}
	var texCoords []math.Vec2 = []math.Vec2{math.Vec2{0, 0}}
	var normals []math.Vec3 = []math.Vec3{math.Vec3{0, 0, 0}}

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

	var geos []Geometry = make([]Geometry, len(mtls)) // one submesh per material

	for _, sGroup := range sGroups {
		var weightedNormals []math.Vec3 = make([]math.Vec3, len(positions)+1)
		var weightedTangents []math.Vec3 = make([]math.Vec3, len(positions)+1)
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
			det := (dTexCoord1.X()*dTexCoord2.Y() - dTexCoord2.X()*dTexCoord1.Y())
			if det != 0 {
				det = 1 / det
			}
			tangent := math.Vec3{
				dTexCoord2.Y()*edge1.X() - dTexCoord1.Y()*edge2.X(),
				dTexCoord2.Y()*edge1.Y() - dTexCoord1.Y()*edge2.Y(),
				dTexCoord2.Y()*edge1.Z() - dTexCoord1.Y()*edge2.Z()}.
				Scale(det)
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
				if tangent.Dot(tangent) == 0 {
					// no tangent - generate arbitrary tangent vector
					// assume normal is not the zero vector
					if normal.Dot(normal) == 0 {
						panic("cannot find vector not parallell to zero vector")
					}
					if normal.X() == 0 {
						tangent = math.Vec3{1, 0, 0}
					} else {
						tangent = math.Vec3{0, 1, 0}
					}
				} else {
					tangent = weightedTangents[iTri.iVerts[i].v].Norm()
				}
				tangent = tangent.Sub(normal.Scale(tangent.Dot(normal))).Norm() // gram schmidt
				verts[i] = NewVertex(pos, texCoord, normal, tangent)
			}
			geos[iTri.mtlInd].AddTriangle(verts[0], verts[1], verts[2])
		}
	}

	m := NewMesh(nil, nil)
	for i, _ := range geos {
		if len(geos[i].Verts) > 0 {
			println("submesh", i, "with", len(geos[i].Verts), "verts")
			if mtls[i] == nil {
				mtls[i] = material.NewDefaultMaterial("")
			}
			m.AddSubMesh(NewSubMesh(&geos[i], mtls[i], m))
		}
	}

	return m, nil
}
