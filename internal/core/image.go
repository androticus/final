package core

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
)

type Image struct {
	image.Image
}

func (img *Image) FinalType() *Type { return imageType }

func savePngToFile(im image.Image, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, im)
}

func saveJpegToFile(im image.Image, filepath string, quality int) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	return jpeg.Encode(file, im, &jpeg.Options{Quality: quality})
}

func saveGifToFile(im image.Image, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	return gif.Encode(file, im, nil)
}

func savePngToStream(im image.Image, stream *ByteWriter) error {
	var buf bytes.Buffer
	if err := png.Encode(&buf, im); err != nil {
		return err
	}
	stream.WriteBytes(buf.Bytes())
	return nil
}

func saveJpegToStream(im image.Image, stream *ByteWriter, quality int) error {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, im, &jpeg.Options{Quality: quality}); err != nil {
		return err
	}
	stream.WriteBytes(buf.Bytes())
	return nil
}

func saveGifToStream(im image.Image, stream *ByteWriter) error {
	var buf bytes.Buffer
	if err := gif.Encode(&buf, im, nil); err != nil {
		return err
	}
	stream.WriteBytes(buf.Bytes())
	return nil
}

func newImageFromFile(filepath string) (*Image, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return &Image{img}, nil
}

// registerImageCodecMembers adds image encode instance methods to t.
// Handlers cast ths.Obj to T (must satisfy image.Image). Not for class factories.
func registerImageCodecMembers[T image.Image](t *Type) {
	t.registerProperty("modelName", "String",
		func(inst Any) Any {
			model := inst.(T).ColorModel()
			switch model {
			case color.RGBAModel:
				return String("RGBA")
			case color.NRGBAModel:
				return String("NRGBA")
			case color.YCbCrModel:
				return String("YCbCr")
			default:
				return String(fmt.Sprintf("%T", model))
			}
		},
		nil,
	)
	t.registerNativeMethodSec("savePngToFile", func(f *RunFrame, ths *TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(String))
		if err := savePngToFile(ths.Obj.(T), path); err != nil {
			panic(err)
		}
	}, SR_FileWrite)
	t.registerNativeMethodSec("saveJpegToFile", func(f *RunFrame, ths *TypeObjPair) {
		quality := int(f.RP.DataStack.Pop().(Int))
		path := string(f.RP.DataStack.Pop().(String))
		if err := saveJpegToFile(ths.Obj.(T), path, quality); err != nil {
			panic(err)
		}
	}, SR_FileWrite)
	t.registerNativeMethodSec("saveGifToFile", func(f *RunFrame, ths *TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(String))
		if err := saveGifToFile(ths.Obj.(T), path); err != nil {
			panic(err)
		}
	}, SR_FileWrite)
	t.registerNativeMethod("savePngToStream", func(f *RunFrame, ths *TypeObjPair) {
		stream := f.RP.DataStack.Pop().(*ByteWriter)
		if err := savePngToStream(ths.Obj.(T), stream); err != nil {
			panic(err)
		}
	})
	t.registerNativeMethod("saveJpegToStream", func(f *RunFrame, ths *TypeObjPair) {
		quality := int(f.RP.DataStack.Pop().(Int))
		stream := f.RP.DataStack.Pop().(*ByteWriter)
		if err := saveJpegToStream(ths.Obj.(T), stream, quality); err != nil {
			panic(err)
		}
	})
	t.registerNativeMethod("saveGifToStream", func(f *RunFrame, ths *TypeObjPair) {
		stream := f.RP.DataStack.Pop().(*ByteWriter)
		if err := saveGifToStream(ths.Obj.(T), stream); err != nil {
			panic(err)
		}
	})
}

func registerImageMembers(t *Type) {
	registerImageCodecMembers[*Image](t)
	// ( path:s -- img:Image )
	t.registerNativeClassMethodSec("newFromFile", func(f *RunFrame, _ *TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(String))
		img, err := newImageFromFile(path)
		if err != nil {
			Throw(err.Error(), f)
			return
		}
		f.RP.DataStack.Push(img)
	}, SR_FileRead)
}

func CreateRegisterImageType() {
	t := &Type{
		name:            "Image",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[Image]("Image"),
	}
	registerImageMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	imageType = t
}

var (
	imageType *Type
	_         Typed       = (*Image)(nil)
	_         image.Image = (*Image)(nil)
)

func init() {
	CreateRegisterImageType()
}
