package lib

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var transliterationMap map[rune]string

func init() {
	transliterationMapString := "ÀA,ÁA,ÂA,ÃA,ÄA,ÅA,ÆAE,ÇC,ÈE,ÉE,ÊE,ËE,ÌI,ÍI,ÎI,ÏI,ÐD,ÑN,ÒO,ÓO,ÔO,ÕO,ÖO,ØO,ÙU,ÚU,ÛU,ÜU,ÝY,ßss,àa,áa,âa,ãa,äa,åa,æae,çc,èe,ée,êe,ëe,ìi,íi,îi,ïi,ðd,ñn,òo,óo,ôo,õo,öo,øo,ùu,úu,ûu,üu,ýy,ÿy,ĀA,āa,ĂA,ăa,ĄA,ąa,ĆC,ćc,ĈC,ĉc,ĊC,ċc,ČC,čc,ĎD,ďd,ĐD,đd,ĒE,ēe,ĔE,ĕe,ĖE,ėe,ĘE,ęe,ĚE,ěe,ĜG,ĝg,ĞG,ğg,ĠG,ġg,ĢG,ģg,ĤH,ĥh,ĦH,ħh,ĨI,ĩi,ĪI,īi,ĬI,ĭi,ĮI,įi,İI,ĲIJ,ĳij,ĴJ,ĵj,ĶK,ķk,ĹL,ĺl,ĻL,ļl,ĽL,ľl,ŁL,łl,ŃN,ńn,ŅN,ņn,ŇN,ňn,ŉ'n,ŌO,ōo,ŎO,ŏo,ŐO,őo,ŒOE,œoe,ŔR,ŕr,ŖR,ŗr,ŘR,řr,ŚS,śs,ŜS,ŝs,ŞS,şs,ŠS,šs,ŢT,ţt,ŤT,ťt,ŨU,ũu,ŪU,ūu,ŬU,ŭu,ŮU,ůu,ŰU,űu,ŲU,ųu,ŴW,ŵw,ŶY,ŷy,ŸY,ŹZ,źz,ŻZ,żz,ŽZ,žz"
	transliterationMap = make(map[rune]string)
	pairs := strings.Split(transliterationMapString, ",")

	for _, pair := range pairs {
		runes := []rune(pair)
		transliterationMap[runes[0]] = string(runes[1:])
	}
}

func transliterate(r rune, transliterationMap map[rune]string) string {
	if result, ok := transliterationMap[r]; ok {
		return result
	}
	return string(r)
}

func NormalizeForComparison(path string) string {
	// Normalize Algorithm
	path = strings.ReplaceAll(path, "\x00", "")
	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.Trim(path, "/")
	path = strings.Join(strings.FieldsFunc(path, func(r rune) bool { return r == '/' }), "/")

	// Normalize For Comparison Algorithm
	path = norm.NFKC.String(path)

	var transliteratedPath strings.Builder
	for _, r := range path {
		transliteratedPath.WriteString(transliterate(r, transliterationMap))
	}
	path = strings.Map(unicode.ToLower, transliteratedPath.String())

	path = strings.TrimRightFunc(path, unicode.IsSpace)

	return path
}
