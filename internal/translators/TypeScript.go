package translators

import (
	"fmt"
	"strings"

	"github.com/averageflow/goschemaconverter/internal/syntaxtree"
)

var goTypeScriptTypeMappings = map[string]string{
	"int":         "number",
	"int32":       "number",
	"int64":       "number",
	"float":       "number",
	"float32":     "number",
	"float64":     "number",
	"string":      "string",
	"bool":        "boolean",
	"interface{}": "any",
}

type TypeScriptTranslator struct {
	preserve       bool
	scannedTypes   map[string]syntaxtree.ScannedType
	scannedStructs map[string]syntaxtree.ScannedStruct
}

func (t *TypeScriptTranslator) Setup(preserve bool, d map[string]syntaxtree.ScannedType, s map[string]syntaxtree.ScannedStruct) {
	t.preserve = preserve
	t.scannedTypes = d
	t.scannedStructs = s
}

func (t *TypeScriptTranslator) Translate() string {
	fmt.Println(`
----------------------------------
Performing TypeScript translation!
----------------------------------
	`)

	var imports string
	var result string

	for i := range t.scannedTypes {
		if len(t.scannedTypes[i].Doc) > 0 {
			for j := range t.scannedTypes[i].Doc {
				result += fmt.Sprintf("%s\n", t.scannedTypes[i].Doc[j])
			}
		}

		switch t.scannedTypes[i].InternalType {
		case syntaxtree.ConstType:
			result += fmt.Sprintf(
				"export const %s: %s = %s;\n\n",
				t.scannedTypes[i].Name,
				getTypescriptCompatibleType(t.scannedTypes[i].Kind),
				t.scannedTypes[i].Value,
			)
		case syntaxtree.MapType:
			result += fmt.Sprintf(
				"export const %s: %s = {\n",
				t.scannedTypes[i].Name,
				getRecordType(t.scannedTypes[i].Kind),
			)
			result += fmt.Sprintf("%s\n", mapValuesToTypeScriptRecord(t.scannedTypes[i].Value.(map[string]string)))
			result += fmt.Sprint("};\n\n")
		case syntaxtree.SliceType:
			result += fmt.Sprintf(
				"export const %s: %s = [\n",
				t.scannedTypes[i].Name,
				transformSliceTypeToTypeScript(t.scannedTypes[i].Kind),
			)
			result += fmt.Sprintf("%s\n", sliceValuesToPrettyList(t.scannedTypes[i].Value.([]string)))

			result += fmt.Sprint("];\n\n")
		}

	}

	for i := range t.scannedStructs {
		var extendsClasses []string
		for j := range t.scannedStructs[i].Fields {
			if isEmbeddedStructForInheritance(t.scannedStructs[i].Fields[j]) {
				extendsClasses = append(extendsClasses, t.scannedStructs[i].Fields[j].Name)
			}
		}

		result += fmt.Sprintf("\nexport interface %s", t.scannedStructs[i].Name)
		if len(extendsClasses) > 0 {
			result += fmt.Sprintf(" extends %s", strings.Join(extendsClasses, ", "))
		}

		result += fmt.Sprint(" {\n")

		for j := range t.scannedStructs[i].Fields {
			if isEmbeddedStructForInheritance(t.scannedStructs[i].Fields[j]) {
				continue
			}

			tag := CleanTagName(t.scannedStructs[i].Fields[j].Tag)
			if tag == "" || t.preserve {
				tag = t.scannedStructs[i].Fields[j].Name
			}

			if t.scannedStructs[i].Fields[j].Doc != nil {
				for k := range t.scannedStructs[i].Fields[j].Doc {
					result += fmt.Sprintf("\t%s\n", t.scannedStructs[i].Fields[j].Doc[k])
				}
			}

			result += fmt.Sprintf("\t%s: %s;\n", tag, getTypescriptCompatibleType(t.scannedStructs[i].Fields[j].Kind))

			if t.scannedStructs[i].Fields[j].ImportDetails != nil {
				imports += fmt.Sprintf(
					"import { %s } from \"%s\";\n",
					t.scannedStructs[i].Fields[j].ImportDetails.EntityName,
					t.scannedStructs[i].Fields[j].ImportDetails.PackageName,
				)
			}
		}

		result += "}\n"
	}

	if imports != "" {
		return fmt.Sprintf("%s\n\n%s", imports, result)
	}

	return result
}
