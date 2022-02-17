package regex

import "regexp"

var Regex, _ = regexp.Compile(`(\pL+nathan)\s*`)

var BadwordRegex, _ = regexp.Compile("[Nn]egro")
