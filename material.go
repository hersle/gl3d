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
	diffuse Vec3
	specular Vec3
	shine float32
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
	mtl.diffuse = NewVec3(1.0, 1.0, 1.0)
	mtl.specular = NewVec3(1.0, 1.0, 1.0)
	mtl.shine = 100 // TODO: change default?

	var tmp [3]float32

	reachedNewMaterial := false
	for s.Scan() && !reachedNewMaterial {
		line := s.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "#":
			continue
		case "newmtl":
			reachedNewMaterial = true
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
		case "Kd":
			if len(fields[1:]) < 3 {
				panic("Kd error")
			}
			for i := 0; i < 3; i++ {
				_, err := fmt.Sscan(fields[1+i], &tmp[i])
				if err != nil {
					panic("Kd error")
				}
			}
			mtl.diffuse = NewVec3(tmp[0], tmp[1], tmp[2])
		case "Ks":
			if len(fields[1:]) < 3 {
				panic("Ks error")
			}
			for i := 0; i < 3; i++ {
				_, err := fmt.Sscan(fields[1+i], &tmp[i])
				if err != nil {
					panic("Ks error")
				}
			}
			mtl.specular = NewVec3(tmp[0], tmp[1], tmp[2])
		case "Ns":
			if len(fields[1:]) != 1 {
				panic("shine error")
			}
			_, err := fmt.Sscan(fields[1], &mtl.shine)
			if err != nil {
				panic("shine error")
			}
		}
	}

	return &mtl
}
