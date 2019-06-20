package utils

import (
	"image"
	"os"
)

func ReadImage(filename string) (image.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func ReadImages(filenames ...string) ([]image.Image, error) {
	imgs := make([]image.Image, len(filenames))
	for i, filename := range filenames {
		img, err := ReadImage(filename)
		if err != nil {
			return nil, err
		}
		imgs[i] = img
	}
	return imgs, nil
}
