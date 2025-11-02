package util

import (
	"math/rand"
	"time"

	"github.com/JeongWoo-Seo/pcBook/pb"
	"github.com/google/uuid"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandomKeyboardLayout() pb.Keyboard_Layout {
	switch rand.Intn(3) {
	case 1:
		return pb.Keyboard_QWERTY
	case 2:
		return pb.Keyboard_QWERTZ
	default:
		return pb.Keyboard_AZERTY
	}
}

func RandomBool() bool {
	return rand.Intn(2) == 1
}

func RandomCPUBrand() string {
	return randomStringFromSet("Intel", "AMD")
}

func RandomGPUBrand() string {
	return randomStringFromSet("NVIDIA", "AMD")
}

func randomStringFromSet(set ...string) string {
	if len(set) == 0 {
		return ""
	}
	return set[rand.Intn(len(set))]
}

func RandomCPUName(brand string) string {
	if brand == "Intel" {
		return randomStringFromSet(
			"Core i7-11800H",
			"Core i9-11900K",
		)
	}
	return randomStringFromSet(
		"Ryzen 9 5900HX",
		"Ryzen 7 5800X",
	)
}

func RandomGPUName(brand string) string {
	if brand == "NVIDIA" {
		return randomStringFromSet(
			"RTX 4080",
			"GTX 2060 ",
		)
	}

	return randomStringFromSet(
		"RX 590",
		"RX 5800",
	)
}

func RandomInt(min, max uint32) uint32 {
	if min >= max {
		return min
	}
	return min + uint32(rand.Intn(int(max-min+1)))
}

func RandomFloat64(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func RandomFloat32(min, max float32) float32 {
	return min + rand.Float32()*(max-min)
}

func RandomScreenPanel() pb.Screen_Panel {
	if rand.Intn(2) == 1 {
		return pb.Screen_IPS
	}
	return pb.Screen_OLED
}

func RandomResolution() *pb.Screen_Resolution {
	height := RandomInt(1000, 4320)
	width := height * 16 / 9

	return &pb.Screen_Resolution{
		Height: uint32(height),
		Width:  uint32(width),
	}
}

func RandomID() string {
	return uuid.New().String()
}

func RandomLaptopBrand() string {
	return randomStringFromSet("Apple", "Dell", "Lenovo")
}

func RandomLaptopName(brand string) string {
	switch brand {
	case "Apple":
		return randomStringFromSet("Macbook Air", "Macbook Pro")
	case "Dell":
		return randomStringFromSet("Latitude", "Vostro", "XPS", "Alienware")
	default:
		return randomStringFromSet("Thinkpad X1", "Thinkpad P1", "Thinkpad P53")
	}
}

func RandomLaptopScore() float64 {
	return float64(RandomInt(1, 10))
}
