package main

var outputPath = "./"

var architectures = []string{"binary-amd64", "binary-armhf", "binary-arm64"}

var components = []string{
	"main",
	"contrib",
	"non-free",
	"n900",
	"droid4",
	"bionic",
	"n9",
	"n950",
	"lima",
	"pinephone",
	"raspberrypi",
	"sunxi",
	"pocophone-f1",
}

var mainurl = "https://maedevu.maemo.org/leste/dists"
var extrasurl = "https://maedevu.maemo.org/extras/dists"

var suites = []string{"beowulf", "beowulf-devel", "chimaera", "chimaera-testing", "chimaera-devel", "chimaera-experimental", "daedalus", "daedalus-testing", "daedalus-devel", "daedalus-experimental", "excalibur", "excalibur-testing", "excalibur-devel", "excalibur-experimental"}
