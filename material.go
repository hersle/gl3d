package main

import (
	"os"
	"bufio"
	"strings"
	"fmt"
)

type Material struct {
	name string
	ambient Vec3
}

// spec: http://paulbourke.net/dataformats/mtl/

func FindMaterialInFiles(filenames []string, mtlName string) *Material {
	for _, filename := range filenames {
		mtl := FindMaterialInFile(filename, mtlName)
		if mtl != nil {
			return mtl
		}
	}
	return nil
}

func FindMaterialInFile(filename string, mtlName string) *Material {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	found := false
	s := bufio.NewScanner(file)
	for s.Scan() {
		line := s.Text()
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "newmtl" && fields[1] == mtlName {
			found = true
			break
		}
	}

	if !found {
		return nil
	}

	var mtl Material
	mtl.name = mtlName
	mtl.ambient = NewVec3(1.0, 1.0, 1.0)

	var tmp [3]float32

	for s.Scan() {
		line := s.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "#":
			continue
		case "newmtl":
			break
		case "Ka":
			if len(fields[1:]) < 3 {
				panic("Ka error")
			}
			for i := 0; i < 3; i++ {
				_, err := fmt.Sscan(fields[1+i], &tmp[i])
				if err != nil {
					panic("Ka error")
				}
			}
			mtl.ambient = NewVec3(tmp[0], tmp[1], tmp[2])
		}
	}

	return &mtl
}
