package translators

import (
	"fmt"
	"strings"

	"github.com/averageflow/godmt/pkg/godmt"
)

var goPHPTypeMappings = map[string]string{ //nolint:gochecknoglobals
	"int":         "int",
	"int32":       "int",
	"int64":       "int",
	"float":       "float",
	"float32":     "float",
	"float64":     "float",
	"string":      "string",
	"bool":        "bool",
	"interface{}": "",
	"NullFloat64": "?float",
	"NullFloat32": "?float",
	"NullInt32":   "?int",
	"NullInt64":   "?int",
	"NullString":  "?string",
}

type PHPTranslator struct {
	Translator
}

func (t *PHPTranslator) Translate() string { //nolint:gocognit,gocyclo,funlen
	var imports string

	result := "<?php\n\n"

	for i := range t.Data.ConstantSort { //nolint:dupl
		entity := t.Data.ScanResult[t.Data.ConstantSort[i]]

		switch entity.InternalType {
		case godmt.ConstType:
			result += "/**\n" //nolint:goconst

			if len(entity.Doc) > 0 {
				for j := range entity.Doc {
					result += fmt.Sprintf(" * %s\n", strings.ReplaceAll(entity.Doc[j], "// ", ""))
				}
			}

			result += fmt.Sprintf(
				" * @const %s %s\n */\n",
				entity.Name,
				GetPHPCompatibleType(entity.Kind),
			)
			result += fmt.Sprintf(
				"const %s = %s;\n\n",
				entity.Name,
				entity.Value,
			)
		case godmt.MapType:
			result += "/**\n"

			if len(entity.Doc) > 0 {
				for j := range entity.Doc {
					result += fmt.Sprintf(" * %s\n", strings.ReplaceAll(entity.Doc[j], "// ", ""))
				}
			}

			result += fmt.Sprintf(" * @const array %s\n */\n", entity.Name)
			result += fmt.Sprintf(
				"const %s = [\n",
				entity.Name,
			)
			result += fmt.Sprintf("%s\n", MapValuesToPHPArray(entity.Value.(map[string]string)))
			result += "];\n\n" //nolint:goconst
		case godmt.SliceType:
			result += "/**\n"

			if len(entity.Doc) > 0 {
				for j := range entity.Doc {
					result += fmt.Sprintf(" * %s\n", strings.ReplaceAll(entity.Doc[j], "// ", ""))
				}
			}

			result += fmt.Sprintf(
				" * @const %s %s\n */\n",
				TransformSliceTypeToPHP(entity.Kind),
				entity.Name,
			)

			result += fmt.Sprintf(
				"const %s = [\n",
				entity.Name,
			)
			result += fmt.Sprintf("%s\n", godmt.SliceValuesToPrettyList(entity.Value.([]string)))
			result += "];\n\n"
		}
	}

	for i := range t.Data.StructSort {
		var extendsClasses []string

		entity := t.Data.StructScanResult[t.Data.StructSort[i]]
		for j := range entity.Fields {
			if IsEmbeddedStructForInheritance(&entity.Fields[j]) {
				extendsClasses = append(extendsClasses, entity.Fields[j].Name)
			}
		}

		if entity.Doc != nil {
			result += "\t/**\n" //nolint:goconst

			for k := range entity.Doc {
				result += fmt.Sprintf("\t * %s\n", strings.ReplaceAll(entity.Doc[k], "// ", ""))
			}

			result += "\t*/\n"
		}

		result += fmt.Sprintf("\nclass %s", entity.Name)

		if len(extendsClasses) > 0 {
			result += fmt.Sprintf(" extends %s", strings.Join(extendsClasses, ", "))
		}

		result += " {\n" //nolint:goconst

		for j := range entity.Fields {
			entityField := entity.Fields[j]
			if IsEmbeddedStructForInheritance(&entityField) {
				continue
			}

			tag := godmt.CleanTagName(entityField.Tag)
			if tag == "" || t.Preserve {
				tag = entityField.Name
			}

			if len(entityField.SubFields) == 0 {
				// Future: Support subfields (nested structs)
				switch entityField.InternalType {
				case godmt.MapType:
					result += "\t/**\n"

					if entityField.Doc != nil {
						for k := range entityField.Doc {
							result += fmt.Sprintf("\t * %s\n", strings.ReplaceAll(entityField.Doc[k], "// ", ""))
						}
					}

					result += fmt.Sprintf("\t * @var %s $%s\n", "array", tag)
					result += "\t */\n" //nolint:goconst
					result += fmt.Sprintf("\tpublic array $%s;\n\n", tag)
				case godmt.SliceType:
					result += "\t/**\n"

					if entityField.Doc != nil {
						for k := range entityField.Doc {
							result += fmt.Sprintf("\t * %s\n", strings.ReplaceAll(entityField.Doc[k], "// ", ""))
						}
					}

					result += fmt.Sprintf("\t * @var %s $%s\n", TransformSliceTypeToPHP(entityField.Kind), tag)
					result += "\t */\n"
					result += fmt.Sprintf("\tpublic array $%s;\n\n", tag)
				default:
					if entityField.Doc != nil {
						result += "\t/**\n"

						for k := range entityField.Doc {
							result += fmt.Sprintf("\t * %s\n", strings.ReplaceAll(entityField.Doc[k], "// ", ""))
						}

						result += "\t */\n"
					}

					result += fmt.Sprintf("\tpublic %s $%s;\n\n", GetPHPCompatibleType(entityField.Kind), tag)
				}
			}

			if entityField.ImportDetails != nil {
				imports += fmt.Sprintf(
					"import { %s } from \"%s\";\n",
					entityField.ImportDetails.EntityName,
					entityField.ImportDetails.PackageName,
				)
			}
		}

		result += "\n\tpublic function __construct(array $data) {\n"

		for j := range entity.Fields {
			if IsEmbeddedStructForInheritance(&entity.Fields[j]) {
				continue
			}

			tag := godmt.CleanTagName(entity.Fields[j].Tag)
			if tag == "" || t.Preserve {
				tag = entity.Fields[j].Name
			}

			result += fmt.Sprintf("\t\t$this->%s = $data['%s'];\n", tag, tag)
		}

		result += "\t}\n}\n"
	}

	if imports != "" {
		return fmt.Sprintf("%s\n\n%s", imports, result)
	}

	return result
}
