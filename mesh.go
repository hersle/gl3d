package main

import (
	"os"
	"bufio"
	"fmt"
	"errors"
)

type Mesh struct {
	verts []Vertex
	faces []int
	// TODO: transformation matrix, etc.
}

func NewMesh(verts []Vertex, faces []int) *Mesh {
	var m Mesh
	m.verts = verts
	m.faces = faces
	return &m
}

func ReadMesh(filename string) (*Mesh, error) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	var m Mesh

	s := bufio.NewScanner(file)
	var x, y, z float64
	var u, v float64
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
		default:
			errMsg = "unexpected first character"
		}

		if errMsg != "" {
			err = errors.New(fmt.Sprintf("%s:%d: %s", filename, lineNo, errMsg))
			return nil, err
		}

		lineNo++
	}

	err = file.Close()
	if err != nil {
		panic(err)
	}

	return &m, nil
}
