package utils

import (
	"image"
	"image/png"
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

func WriteImagePNG(img image.Image, filename string) error {
	out, err := os.Create(filename)
	if err != nil {
		return err
	}

	png.Encode(out, img)
	out.Close()

	return nil
}
