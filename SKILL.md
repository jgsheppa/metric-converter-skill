---
name: unit-converter
description: >
  Use this skill whenever the user wants to convert units of measurement —
  including length (km, miles, feet, inches, yards, cm, mm), weight (kg, lbs,
  oz, grams, tonnes), and temperature (°C, °F, Kelvin). Trigger on phrases
  like "convert X to Y", "how many miles is X km", "what is X°F in Celsius",
  "X lbs in kg", or any time the user mentions two units of the same category.
  Always use this skill rather than doing unit math manually.
---

# Unit Converter Skill

Converts values between units of **length**, **weight**, and **temperature**
using a Go script. Supports both natural-language-style positional args and
explicit CLI flags. Can output plain text or JSON.

---

## Supported Units

| Category    | Units |
|-------------|-------|
| Length      | `m`, `km`, `cm`, `mm`, `mi`/`miles`, `ft`/`feet`, `in`/`inches`, `yd`/`yards` |
| Weight      | `g`, `kg`, `mg`, `lb`/`lbs`/`pounds`, `oz`/`ounces`, `t`/`tonne` |
| Temperature | `c`/`celsius`, `f`/`fahrenheit`, `k`/`kelvin` |

---

## How to Invoke the Script

The script lives at `scripts/main.go` relative to this skill directory.
Run it with `go run` from that directory.

### Natural language mode (positional args)
```bash
go run scripts/main.go <value> <from_unit> <to_unit>
```
Examples:
```bash
go run scripts/main.go 5 km miles
go run scripts/main.go 70 kg lbs
go run scripts/main.go 100 f c
```

### Flag mode
```bash
go run scripts/main.go --value <n> --from <unit> --to <unit>
```
Examples:
```bash
go run scripts/main.go --value 6 --from feet --to m
go run scripts/main.go --value 212 --from f --to celsius
```

### JSON output
Add `--json` to either invocation style:
```bash
go run scripts/main.go --value 180 --from lbs --to kg --json
```
Returns:
```json
{
  "input": 180,
  "from_unit": "lbs",
  "to_unit": "kg",
  "output": 81.64656,
  "formatted": "180 lbs = 81.64656 kg",
  "category": "weight"
}
```

---

## Steps

1. Parse the user's request to extract: numeric value, source unit, target unit.
2. Normalise unit names to lowercase (e.g. "Fahrenheit" → `fahrenheit`, "KG" → `kg`).
3. Run the script using the natural language or flag invocation style — either works.
4. Present the `formatted` field from the output to the user in a friendly sentence.
5. If the conversion fails (unknown unit, wrong category), report the error message
   from stderr and ask the user to clarify the units.

---

## Notes

- Units are **case-insensitive** — pass them lowercase.
- Do **not** mix categories (e.g. `km` → `lbs` will error — that's correct behaviour).
- For temperature, `k` means Kelvin; do not confuse with `km`.
- The script uses stdlib only — no `go.mod` needed, just `go run`.
