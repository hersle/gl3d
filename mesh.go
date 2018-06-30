package main

import (
	"os"
	"bufio"
	"fmt"
	"errors"
	"strings"
	"strconv"
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

	var pos [3]float32
	var texCoord [2]float32
	var inds [3]int
	var positions []Vec3
	var texCoords []Vec2

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
			i1 := len(m.verts)
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
				vert := Vertex{positions[inds[0]-1], RGBAColor{}, texCoords[inds[1]-1]}
				i2 := len(m.verts) - 1
				i3 := len(m.verts)
				m.faces = append(m.faces, i1, i2, i3)
				m.verts = append(m.verts, vert)
			}
		default:
		}

		if errMsg != "" {
			err = errors.New(fmt.Sprintf("%s:%d: %s", filename, lineNo, errMsg))
			return nil, err
		}
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
