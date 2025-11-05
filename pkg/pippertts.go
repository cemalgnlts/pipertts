package pipertts

/*
#cgo CFLAGS: -I${SRCDIR}/../internal/libpiper/install/include  -std=c11
#cgo LDFLAGS: -L${SRCDIR}/../internal/libpiper/install -lpiper
#cgo LDFLAGS: -L${SRCDIR}/../internal/libpiper/install/lib -lonnxruntime
#cgo LDFLAGS: -Wl,-rpath,${SRCDIR}/../internal/libpiper/install
#include <stdlib.h>
#include <stdint.h>
#include "piper.h"
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path"
	"unsafe"
)

func Generate(model, text string) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	modelJSON := fmt.Sprintf("%s.json", model)

	espeakData := path.Join(cwd, "internal", "libpiper", "install", "espeak-ng-data")
	model = path.Join(cwd, model)
	modelJSON = path.Join(cwd, modelJSON)

	cModel := C.CString(model)
	cModelJSON := C.CString(modelJSON)
	cEspeak := C.CString(espeakData)
	// free C strings after use
	defer C.free(unsafe.Pointer(cModel))
	defer C.free(unsafe.Pointer(cModelJSON))
	defer C.free(unsafe.Pointer(cEspeak))

	synth := C.piper_create(cModel, cModelJSON, cEspeak)
	if synth == nil {
		fmt.Fprintln(os.Stderr, "piper_create returned NULL")
		os.Exit(1)
	}
	// Ensure resources are freed
	defer C.piper_free(synth)

	// README örneği: metni sentezleyip WAV formatında output.wav oluşturur
	if err := synthesizeToWav(synth, text, "output.wav"); err != nil {
		fmt.Fprintln(os.Stderr, "synthesis error:", err)
		os.Exit(1)
	}

	fmt.Println("Synthesis complete: output.wav")
}

func synthesizeToWav(synth *C.piper_synthesizer, text string, outPath string) error {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	// default options
	opts := C.piper_default_synthesize_options(synth)
	C.piper_synthesize_start(synth, cText, &opts)

	// if C.piper_synthesize_start(synth, cText, &opts) == C.PIPER_ERROR {
	// return fmt.Errorf("piper_synthesize_start failed")
	// }

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// placeholder for WAV header (44 bytes)
	if _, err := f.Write(make([]byte, 44)); err != nil {
		return err
	}

	// Use sample rate from README/example; change if your model uses different rate
	sampleRate := uint32(22050)
	numChannels := uint16(1)
	bitsPerSample := uint16(16)

	var totalDataBytes uint32
	var chunk C.piper_audio_chunk

	for {
		ret := C.piper_synthesize_next(synth, &chunk)
		if ret == C.PIPER_DONE {
			break
		}
		// if ret == C.PIPER_ERROR {
		// 	return fmt.Errorf("piper_synthesize_next returned error")
		// }

		n := int(chunk.num_samples)
		bytesLen := C.int(n * int(unsafe.Sizeof(float32(0))))
		bs := C.GoBytes(unsafe.Pointer(chunk.samples), bytesLen)

		// convert each float32 (assumed little-endian IEEE754) to int16 PCM and write
		for i := range n {
			off := i * 4
			fbits := binary.LittleEndian.Uint32(bs[off : off+4])
			fval := math.Float32frombits(fbits)

			// scale to int16 range with clipping
			sample := max(min(int32(math.Round(float64(fval*32767.0))), 32767), -32768)

			var tmp [2]byte
			binary.LittleEndian.PutUint16(tmp[:], uint16(int16(sample)))
			if _, err := f.Write(tmp[:]); err != nil {
				return err
			}
			totalDataBytes += 2
		}
	}

	// write WAV header with correct sizes
	dataLen := totalDataBytes
	byteRate := sampleRate * uint32(numChannels) * uint32(bitsPerSample) / 8
	blockAlign := numChannels * bitsPerSample / 8
	riffChunkSize := 36 + dataLen

	header := make([]byte, 44)
	copy(header[0:], []byte("RIFF"))
	binary.LittleEndian.PutUint32(header[4:], riffChunkSize)
	copy(header[8:], []byte("WAVE"))
	copy(header[12:], []byte("fmt "))
	binary.LittleEndian.PutUint32(header[16:], 16)            // Subchunk1Size for PCM
	binary.LittleEndian.PutUint16(header[20:], 1)             // AudioFormat 1 = PCM
	binary.LittleEndian.PutUint16(header[22:], numChannels)   // NumChannels
	binary.LittleEndian.PutUint32(header[24:], sampleRate)    // SampleRate
	binary.LittleEndian.PutUint32(header[28:], byteRate)      // ByteRate
	binary.LittleEndian.PutUint16(header[32:], blockAlign)    // BlockAlign
	binary.LittleEndian.PutUint16(header[34:], bitsPerSample) // BitsPerSample
	copy(header[36:], []byte("data"))
	binary.LittleEndian.PutUint32(header[40:], dataLen) // Subchunk2Size

	// go back to start and write header
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	if _, err := f.Write(header); err != nil {
		return err
	}

	return nil
}
