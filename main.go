package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// Result is the JSON output structure
type Result struct {
	Input      float64 `json:"input"`
	FromUnit   string  `json:"from_unit"`
	ToUnit     string  `json:"to_unit"`
	Output     float64 `json:"output"`
	Formatted  string  `json:"formatted"`
	Category   string  `json:"category"`
}

// --- Length conversions (base unit: metres) ---
var lengthToMetres = map[string]float64{
	"m":    1.0,
	"km":   1000.0,
	"cm":   0.01,
	"mm":   0.001,
	"mi":   1609.344,
	"mile": 1609.344,
	"miles": 1609.344,
	"ft":   0.3048,
	"feet": 0.3048,
	"foot": 0.3048,
	"in":   0.0254,
	"inch": 0.0254,
	"inches": 0.0254,
	"yd":   0.9144,
	"yard": 0.9144,
	"yards": 0.9144,
}

// --- Weight conversions (base unit: grams) ---
var weightToGrams = map[string]float64{
	"g":    1.0,
	"kg":   1000.0,
	"mg":   0.001,
	"lb":   453.592,
	"lbs":  453.592,
	"pound": 453.592,
	"pounds": 453.592,
	"oz":   28.3495,
	"ounce": 28.3495,
	"ounces": 28.3495,
	"t":    1_000_000.0,
	"ton":  1_000_000.0,
	"tonne": 1_000_000.0,
}

func convertLength(value float64, from, to string) (float64, error) {
	fromFactor, ok := lengthToMetres[from]
	if !ok {
		return 0, fmt.Errorf("unknown length unit: %q", from)
	}
	toFactor, ok := lengthToMetres[to]
	if !ok {
		return 0, fmt.Errorf("unknown length unit: %q", to)
	}
	metres := value * fromFactor
	return metres / toFactor, nil
}

func convertWeight(value float64, from, to string) (float64, error) {
	fromFactor, ok := weightToGrams[from]
	if !ok {
		return 0, fmt.Errorf("unknown weight unit: %q", from)
	}
	toFactor, ok := weightToGrams[to]
	if !ok {
		return 0, fmt.Errorf("unknown weight unit: %q", to)
	}
	grams := value * fromFactor
	return grams / toFactor, nil
}

func convertTemperature(value float64, from, to string) (float64, error) {
	// Normalise to Celsius first
	var celsius float64
	switch from {
	case "c", "celsius":
		celsius = value
	case "f", "fahrenheit":
		celsius = (value - 32) * 5 / 9
	case "k", "kelvin":
		celsius = value - 273.15
	default:
		return 0, fmt.Errorf("unknown temperature unit: %q", from)
	}

	switch to {
	case "c", "celsius":
		return celsius, nil
	case "f", "fahrenheit":
		return celsius*9/5 + 32, nil
	case "k", "kelvin":
		return celsius + 273.15, nil
	default:
		return 0, fmt.Errorf("unknown temperature unit: %q", to)
	}
}

func detectCategory(unit string) string {
	if _, ok := lengthToMetres[unit]; ok {
		return "length"
	}
	if _, ok := weightToGrams[unit]; ok {
		return "weight"
	}
	switch unit {
	case "c", "celsius", "f", "fahrenheit", "k", "kelvin":
		return "temperature"
	}
	return ""
}

func convert(value float64, from, to string) (float64, string, error) {
	cat := detectCategory(from)
	if cat == "" {
		return 0, "", fmt.Errorf("unrecognised unit: %q", from)
	}

	var result float64
	var err error
	switch cat {
	case "length":
		result, err = convertLength(value, from, to)
	case "weight":
		result, err = convertWeight(value, from, to)
	case "temperature":
		result, err = convertTemperature(value, from, to)
	}
	return result, cat, err
}

func formatNumber(f float64) string {
	// Use up to 6 significant figures, trim trailing zeros
	if f == 0 {
		return "0"
	}
	abs := math.Abs(f)
	var s string
	if abs >= 1000 || abs < 0.001 {
		s = strconv.FormatFloat(f, 'g', 6, 64)
	} else {
		s = strconv.FormatFloat(f, 'f', 6, 64)
		// Trim trailing zeros after decimal
		if strings.Contains(s, ".") {
			s = strings.TrimRight(s, "0")
			s = strings.TrimRight(s, ".")
		}
	}
	return s
}

func main() {
	// --- CLI flag mode ---
	valueFlag := flag.Float64("value", 0, "Numeric value to convert")
	fromFlag := flag.String("from", "", "Source unit (e.g. km, lbs, f)")
	toFlag := flag.String("to", "", "Target unit (e.g. miles, kg, c)")
	jsonFlag := flag.Bool("json", false, "Output as JSON")
	flag.Parse()

	var value float64
	var fromUnit, toUnit string

	if *fromFlag != "" && *toFlag != "" {
		// Flag mode
		value = *valueFlag
		fromUnit = strings.ToLower(*fromFlag)
		toUnit = strings.ToLower(*toFlag)
	} else {
		// Natural language mode: expects positional args like: 5 km miles
		args := flag.Args()
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage:")
			fmt.Fprintln(os.Stderr, "  Natural language : main.go <value> <from> <to>")
			fmt.Fprintln(os.Stderr, "  Flags            : main.go --value 5 --from km --to miles")
			os.Exit(1)
		}
		var err error
		value, err = strconv.ParseFloat(args[0], 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid value %q: %v\n", args[0], err)
			os.Exit(1)
		}
		fromUnit = strings.ToLower(args[1])
		toUnit = strings.ToLower(args[2])
	}

	result, category, err := convert(value, fromUnit, toUnit)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	formatted := fmt.Sprintf("%s %s = %s %s", formatNumber(value), fromUnit, formatNumber(result), toUnit)

	if *jsonFlag {
		out := Result{
			Input:     value,
			FromUnit:  fromUnit,
			ToUnit:    toUnit,
			Output:    result,
			Formatted: formatted,
			Category:  category,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(out)
	} else {
		fmt.Println(formatted)
	}
}
