package main

import (
	"os"
	"bufio"
	"fmt"
	"errors"
	"strings"
	"image"
	_ "image/png"
)

type Mesh struct {
	verts []Vertex
	faces []int
	tex *Texture2D
	// TODO: transformation matrix, etc.
}

func NewMesh(verts []Vertex, faces []int, tex *Texture2D) *Mesh {
	var m Mesh
	m.verts = verts
	m.faces = faces
	m.tex = tex
	return &m
}

func ReadMesh(filename string) (*Mesh, error) {
	if strings.HasSuffix(filename, ".3d") {
		return ReadMeshCustom(filename)
	} else if strings.HasSuffix(filename, ".obj") {
		return ReadMeshObj(filename)
	} else {
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

	var x, y, z float32
	var u, v float32
	var v1, vt1, vn1 int
	var positions []Vec3
	var texCoords []Vec2

	errMsg := ""
	lineNo := 1
	s := bufio.NewScanner(file)
	for s.Scan() {
		line := s.Text()

		if len(line) == 0 || strings.HasPrefix(line, "#") {
			// empty line or comment
		} else if strings.HasPrefix(line, "vn") {
		} else if strings.HasPrefix(line, "vt") {
			n, err := fmt.Sscanf(line, "vt %f %f", &u, &v)
			if n == 2 && err == nil {
				texCoords = append(texCoords, NewVec2(u, v))
			} else {
				errMsg = "texture data error"
			}
		} else if strings.HasPrefix(line, "v") {
			n, err := fmt.Sscanf(line, "v %f %f %f", &x, &y, &z)
			if n == 3 && err == nil {
				positions = append(positions, NewVec3(x, y, z))
			} else {
				errMsg = "vertex data error"
			}
		} else if strings.HasPrefix(line, "f") {
			vertSpecs := strings.Fields(line)[1:]
			i1 := len(m.verts)
			for _, vertSpec := range vertSpecs {
				n, _ := fmt.Sscanf(vertSpec, "%d/%d/%d", &v1, &vt1, &vn1)
				if (n == 2 || n == 3) {
					vert := Vertex{positions[v1-1], RGBAColor{}, texCoords[vt1-1]}
					i2 := len(m.verts) - 1
					i3 := len(m.verts)
					m.faces = append(m.faces, i1, i2, i3)
					m.verts = append(m.verts, vert)
				} else {
					errMsg = "face data error"
					break
				}
			}
		} else {
			fmt.Printf("warning: unknown line prefix: %s\n", strings.Fields(line)[0])
		}

		if errMsg != "" {
			err = errors.New(fmt.Sprintf("%s:%d: %s", filename, lineNo, errMsg))
			return nil, err
		}

		lineNo++
	}

	m.tex = NewTexture2D()
	err = m.tex.ReadImage(strings.TrimSuffix(filename, ".obj") + "_texture.png")
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func ReadMeshCustom(filename string) (*Mesh, error) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	var m Mesh

	s := bufio.NewScanner(file)
	var x, y, z float32
	var u, v float32
	var r, g, b uint8
	var i1, i2, i3 int
	errMsg := ""
	lineNo := 1
	for s.Scan() {
		line := s.Text()

		if len(line) == 0 {
			continue
		}

		switch line[0] {
		case '#':
			continue
		case 'v':
			line = line[1:]
			n, err := fmt.Sscanf(line, "%f %f %f %02x%02x%02x %f %f", &x, &y, &z, &r, &g, &b, &u, &v)
			if n == 8 && err == nil {
				pos := NewVec3(x, y, z)
				color := NewColor(r, g, b, 0xff)
				texCoord := NewVec2(u, v)
				vert := Vertex{pos, color, texCoord}
				m.verts = append(m.verts, vert)
			} else {
				errMsg = "vertex data error"
			}
		case 'f':
			line = line[1:]
			n, err := fmt.Sscanf(line, "%d %d %d", &i1, &i2, &i3)
			if n == 3 && err == nil {
				m.faces = append(m.faces, i1 - 1, i2 - 1, i3 - 1)
			} else {
				errMsg = "face data error"
			}
		case 't':
			if m.tex != nil {
				errMsg = "multiple textures specified"
			} else {
				line = line[1:]
				var texFilename string
				n, err := fmt.Sscanf(line, "%s", &texFilename)
				if n == 1 && err == nil {
					m.tex = NewTexture2D()
					file, err := os.Open(texFilename)
					if err != nil {
						errMsg = err.Error()
					}
					defer file.Close()
					img, _, err := image.Decode(file)
					if err != nil {
						errMsg = err.Error()
					} else {
						m.tex.SetImage(img)
					}
				} else {
					errMsg = "texture specification error"
				}
			}
		default:
			errMsg = "unexpected first character"
		}

		if errMsg != "" {
			err = errors.New(fmt.Sprintf("%s:%d: %s", filename, lineNo, errMsg))
			return nil, err
		}

		lineNo++
	}

	if m.tex == nil {
		err = errors.New("no texture specified")
		return nil, err
	}

	err = file.Close()
	if err != nil {
		panic(err)
	}

	return &m, nil
}
