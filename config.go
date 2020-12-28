package main

var outputPath = "./out"

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
}

var url = "https://maedevu.maemo.org/leste/dists"

var suites = []string{"beowulf"}
