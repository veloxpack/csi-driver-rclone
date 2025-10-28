package asset

import (
	"errors"
	"fmt"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

const queryString = "_a"
const algoVersion = "A" // The version of the algorithm
const sdkCode = "Q"     // Cloudinary Go SDK

var sdkVersion = api.Version
var techVersion = strings.Join(strings.Split(strings.TrimPrefix(runtime.Version(), "go"), ".")[:2], ".")

const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
const binaryPadSize = 6

var charCodes = map[string]string{}
var initialize sync.Once

var analyticsSignature = ""

func sdkAnalyticsSignature() string {
	if analyticsSignature != "" {
		return analyticsSignature
	}
	sdkVersionStr, err := encodeVersion(sdkVersion)
	if err != nil {
		analyticsSignature = "E"

		return analyticsSignature
	}

	techVersionStr, err := encodeVersion(techVersion)
	if err != nil {
		analyticsSignature = "E"

		return analyticsSignature
	}

	analyticsSignature = fmt.Sprintf("%s%s%s%s", algoVersion, sdkCode, sdkVersionStr, techVersionStr)

	return analyticsSignature
}

func encodeVersion(version string) (string, error) {

	parts := strings.Split(version, ".")

	paddedParts := make([]string, len(parts))
	for i, v := range parts {
		vInt, _ := strconv.Atoi(v)
		paddedParts[i] = fmt.Sprintf("%02d", vInt)
	}

	// reverse (in this case swap first and last elements)
	paddedParts[0], paddedParts[len(paddedParts)-1] = paddedParts[len(paddedParts)-1], paddedParts[0]

	num, _ := strconv.Atoi(strings.Join(paddedParts, ""))
	paddedBinary := intToPaddedBin(num, len(parts)*binaryPadSize)

	if len(paddedBinary)%binaryPadSize != 0 {
		return "", errors.New("version must be smaller than 43.21.26")
	}

	encodedChars := make([]string, len(parts))

	for i := 0; i < len(parts); i++ {
		encodedChars[i] = getKey(paddedBinary[i*binaryPadSize : (i+1)*binaryPadSize])
	}

	return strings.Join(encodedChars, ""), nil
}

func initializeCharCodes() {
	for i, char := range chars {
		charCodes[intToPaddedBin(i, binaryPadSize)] = string(char)
	}
}

func getKey(binaryValue string) string {
	initialize.Do(initializeCharCodes)

	return charCodes[binaryValue]
}

func intToPaddedBin(integer int, padNum int) string {
	return fmt.Sprintf("%0*b", padNum, integer)
}
