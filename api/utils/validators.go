package utils

import (
	"strconv"
	"strings"
)

func ValidateRUT(rut string) bool {
	rut = strings.TrimSpace(rut)
	rut = strings.ToUpper(rut)
	rut = strings.ReplaceAll(rut, ".", "")
	rut = strings.ReplaceAll(rut, "-", "")

	if len(rut) < 2 {
		return false
	}

	bodyStr := rut[:len(rut)-1]
	verifier := rut[len(rut)-1:]

	body, err := strconv.Atoi(bodyStr)
	if err != nil {
		return false
	}

	calculatedVerifier := calculateVerifier(body)
	return calculatedVerifier == verifier
}

func calculateVerifier(rutBody int) string {
	sum := 0
	multiplier := 2

	tempBody := rutBody
	for tempBody > 0 {
		digit := tempBody % 10
		tempBody = tempBody / 10
		sum += digit * multiplier
		multiplier++
		if multiplier > 7 {
			multiplier = 2
		}
	}

	mod := 11 - (sum % 11)
	if mod == 11 {
		return "0"
	}
	if mod == 10 {
		return "K"
	}
	return strconv.Itoa(mod)
}
