package pronunciation

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"

	mp3 "github.com/hajimehoshi/go-mp3"
)

const MaxAudioSeconds = 60

// AudioQuality describes signal presence, not pronunciation/meaning quality.
// Thresholds reject near silence without trying to select preferred voices.
type AudioQuality struct {
	SampleRate      int     `json:"sampleRate"`
	DurationSeconds float64 `json:"durationSeconds"`
	Peak            float64 `json:"peak"`
	RMS             float64 `json:"rms"`
}

// ValidateAudio fully decodes bounded MPEG Layer III audio before it is saved.
// A strict frame walk is necessary because the decoder treats some truncated
// frames as EOF. No provider retry is performed when this gate fails.
func ValidateAudio(data []byte) (quality AudioQuality, err error) {
	err = ErrUnavailable
	defer func() {
		if recover() != nil {
			quality = AudioQuality{}
			err = ErrUnavailable
		}
	}()
	rate, expected, ok := audioFrames(data)
	if !ok {
		return quality, ErrUnavailable
	}
	decoder, decodeErr := mp3.NewDecoder(bytes.NewReader(data))
	if decodeErr != nil || decoder.SampleRate() != rate || decoder.Length() != int64(expected) {
		return quality, ErrUnavailable
	}
	var total, samples int
	var sumSquares, peak float64
	buffer := make([]byte, 8192)
	for {
		n, readErr := decoder.Read(buffer)
		if n < 0 || n > len(buffer) || n%4 != 0 || total+n > expected {
			return quality, ErrUnavailable
		}
		total += n
		for i := 0; i < n; i += 2 {
			value := float64(int16(binary.LittleEndian.Uint16(buffer[i:i+2]))) / 32768
			absolute := math.Abs(value)
			if absolute > peak {
				peak = absolute
			}
			sumSquares += value * value
			samples++
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil || n == 0 {
			return quality, ErrUnavailable
		}
	}
	if total != expected || samples == 0 {
		return quality, ErrUnavailable
	}
	quality = AudioQuality{SampleRate: rate, DurationSeconds: float64(total) / float64(rate*4), Peak: peak, RMS: math.Sqrt(sumSquares / float64(samples))}
	if quality.DurationSeconds < 0.12 || quality.DurationSeconds > MaxAudioSeconds || quality.Peak < 0.001 || quality.RMS < 0.0001 {
		return quality, ErrUnavailable
	}
	return quality, nil
}

// audioFrames permits ID3 metadata and complete MPEG-1/2 Layer III frames only.
// It bounds decoder work before any synthesis and rejects format changes.
func audioFrames(data []byte) (sampleRate, pcmBytes int, ok bool) {
	if !validMP3(data) {
		return 0, 0, false
	}
	start, end := 0, len(data)
	if len(data) >= 3 && string(data[:3]) == "ID3" {
		if len(data) < 10 || data[3] < 2 || data[3] > 4 {
			return 0, 0, false
		}
		size := 0
		for _, b := range data[6:10] {
			if b&0x80 != 0 {
				return 0, 0, false
			}
			size = size<<7 | int(b)
		}
		start = 10 + size
		if data[3] == 4 && data[5]&0x10 != 0 {
			start += 10
		}
		if start > end {
			return 0, 0, false
		}
	}
	if end-start >= 128 && string(data[end-128:end-125]) == "TAG" {
		end -= 128
	}
	firstVersion := -1
	bitrate1 := [16]int{0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 0}
	bitrate2 := [16]int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0}
	for start < end {
		if end-start < 4 {
			return 0, 0, false
		}
		header := binary.BigEndian.Uint32(data[start : start+4])
		version := int((header >> 19) & 3)
		layer := int((header >> 17) & 3)
		bitrateIndex := int((header >> 12) & 15)
		frequencyIndex := int((header >> 10) & 3)
		if header&0xffe00000 != 0xffe00000 || (version != 2 && version != 3) || layer != 1 || bitrateIndex == 0 || bitrateIndex == 15 || frequencyIndex == 3 || header&3 == 2 {
			return 0, 0, false
		}
		rate := [3]int{44100, 48000, 32000}[frequencyIndex]
		samples, coefficient, bitrate := 1152, 144000, bitrate1[bitrateIndex]
		if version == 2 {
			rate /= 2
			samples = 576
			coefficient = 72000
			bitrate = bitrate2[bitrateIndex]
		}
		if sampleRate == 0 {
			sampleRate = rate
			firstVersion = version
		}
		if rate != sampleRate || version != firstVersion || rate < 16000 || rate > 48000 {
			return 0, 0, false
		}
		frameSize := coefficient*bitrate/rate + int((header>>9)&1)
		if frameSize < 4 || frameSize > end-start {
			return 0, 0, false
		}
		pcmBytes += samples * 4
		if pcmBytes > rate*4*MaxAudioSeconds {
			return 0, 0, false
		}
		start += frameSize
	}
	return sampleRate, pcmBytes, sampleRate > 0 && pcmBytes > 0
}
