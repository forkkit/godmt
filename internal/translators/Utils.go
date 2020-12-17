package translators

import (
	"fmt"
	"strings"

	"github.com/averageflow/goschemaconverter/internal/syntaxtree"
)

func CleanTagName(rawTagName string) string {
	var result string

	result = strings.ReplaceAll(rawTagName, ",string", "")
	result = strings.ReplaceAll(result, "`json:\"", "")
	result = strings.ReplaceAll(result, ",omitempty", "")
	result = strings.ReplaceAll(result, "\"`", "")
	result = strings.ReplaceAll(result, `binding:"required"`, ``)
	return strings.TrimSpace(result)
}

func isEmbeddedStructForInheritance(field syntaxtree.ScannedStructField) bool {
	return field.Kind == "struct" && field.Tag == ""
}

func getTypescriptCompatibleType(goType string) string {
	result, ok := goTypeScriptTypeMappings[goType]
	if !ok {
		return goType
	}

	return result
}

func getSwiftCompatibleType(goType string) string {
	result, ok := goSwiftTypeMappings[goType]
	if !ok {
		return goType
	}

	return result
}

func getRecordType(goMapType string) string {
	var result string

	result = strings.ReplaceAll(goMapType, "map[", "")
	recordTypes := strings.Split(result, "]")

	for i := range recordTypes {
		recordTypes[i] = getTypescriptCompatibleType(recordTypes[i])
	}

	result = strings.Join(recordTypes, ", ")

	return fmt.Sprintf("Record<%s>", result)
}

func getDictionaryType(goMapType string) string {
	var result string

	result = strings.ReplaceAll(goMapType, "map[", "")
	recordTypes := strings.Split(result, "]")

	for i := range recordTypes {
		recordTypes[i] = getSwiftCompatibleType(recordTypes[i])
	}

	result = strings.Join(recordTypes, ", ")

	return fmt.Sprintf("Dictionary<%s>", result)
}

func mapValuesToTypeScriptRecord(rawMap map[string]string) string {
	var entries []string
	for i := range rawMap {
		entries = append(entries, fmt.Sprintf("\t%s: %s", i, rawMap[i]))
	}

	return strings.Join(entries, ",\n")
}

func transformSliceTypeToTypeScript(rawSliceType string) string {
	var result string

	result = strings.ReplaceAll(rawSliceType, "[]", "")
	return fmt.Sprintf("%s[]", getTypescriptCompatibleType(result))
}

func transformSliceTypeToSwift(rawSliceType string) string {
	var result string

	result = strings.ReplaceAll(rawSliceType, "[]", "")
	return fmt.Sprintf("[%s]", getSwiftCompatibleType(result))
}

func sliceValuesToPrettyList(raw []string) string {
	var result []string

	for i := range raw {
		result = append(result, fmt.Sprintf("\t%s", raw[i]))
	}

	return strings.Join(result, ",\n")
}
