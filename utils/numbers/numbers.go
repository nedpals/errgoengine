package numbers

import (
	"strings"
)

func ToWords(num int) string {
	if num == 0 {
		return "zero"
	}

	billions := num / 1000000000
	millions := (num % 1000000000) / 1000000
	thousands := (num % 1000000) / 1000
	hundreds := (num % 1000) / 100
	tens := (num % 100) / 10
	ones := num % 10

	var sb strings.Builder

	if billions > 0 {
		sb.WriteString(threeDigitNumberToWords(billions))
		sb.WriteString(" billion")
	}

	if millions > 0 {
		sb.WriteString(threeDigitNumberToWords(millions))
		sb.WriteString(" million")
	}

	if thousands > 0 {
		sb.WriteString(threeDigitNumberToWords(thousands))
		sb.WriteString(" thousand")
	}

	if hundreds > 0 {
		sb.WriteString(twoDigitNumberToWords(hundreds))
		sb.WriteString(" hundred")
	}

	if tens > 0 || ones > 0 {
		sb.WriteString(twoDigitNumberToWords(tens*10 + ones))
	}

	return sb.String()
}

func threeDigitNumberToWords(num int) string {
	if num == 0 {
		return ""
	}

	hundreds := num / 100
	tens := (num % 100) / 10
	ones := num % 10

	var sb strings.Builder

	sb.WriteString(twoDigitNumberToWords(hundreds*100 + tens*10 + ones))

	return sb.String()
}

var tensInOne = []string{
	"ten",
	"eleven",
	"twelve",
	"thirteen",
	"fourteen",
	"fifteen",
	"sixteen",
	"seventeen",
	"eighteen",
	"nineteen",
}

var tensPrefixes = map[int]string{
	1: "",
	2: "twenty",
	3: "thirty",
	4: "forty",
	5: "fifty",
	6: "sixty",
	7: "seventy",
	8: "eighty",
	9: "ninety",
}

var onesWords = map[int]string{
	1: "one",
	2: "two",
	3: "three",
	4: "four",
	5: "five",
	6: "six",
	7: "seven",
	8: "eight",
	9: "nine",
}

func twoDigitNumberToWords(num int) string {
	if num == 0 {
		return ""
	}

	tens := num / 10
	ones := num % 10

	if tens == 1 {
		return tensInOne[ones]
	} else if prefix, ok := tensPrefixes[tens]; ok {
		return prefix + " " + oneDigitNumberToWords(ones)
	}

	return oneDigitNumberToWords(ones)
}

func oneDigitNumberToWords(num int) string {
	return onesWords[num]
}
