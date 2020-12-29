package syntaxtree

import (
	"go/ast"
	"reflect"

	"github.com/averageflow/godmt/pkg/godmt"
)

func parseStruct(d *ast.Ident) []godmt.ScannedStruct {
	var result []godmt.ScannedStruct

	structTypes := reflect.ValueOf(d.Obj.Decl).Elem().FieldByName("Type")
	if !structTypes.IsValid() {
		return result
	}

	switch structTypes.Interface().(type) {
	case *ast.StructType:
		fields := structTypes.Interface().(*ast.StructType)
		fieldList := fields.Fields.List

		var rawScannedFields []godmt.ScannedStructField
		for i := range fieldList {
			parsed := parseStructField(fieldList[i])
			if parsed != nil {
				rawScannedFields = append(rawScannedFields, *parsed)
			}
		}

		result = append(result, godmt.ScannedStruct{
			Doc:          nil,
			Name:         d.Name,
			Fields:       rawScannedFields,
			InternalType: godmt.StructType,
		})

	default:
		break
	}

	return result
}

func parseStructField(item *ast.Field) *godmt.ScannedStructField {
	switch item.Type.(type) {
	case *ast.Ident:
		return godmt.SimpleStructFieldParser(item)
	case *ast.StructType:
		return parseNestedStruct(item)
	case *ast.SelectorExpr:
		return godmt.ImportedStructFieldParser(item)
	default:
		return nil
	}
}

func parseNestedStruct(field *ast.Field) *godmt.ScannedStructField {
	nestedFields := reflect.ValueOf(field.Type).Elem().FieldByName("Fields").Interface().(*ast.FieldList)

	var parsedNestedFields []godmt.ScannedStructField

	for i := range nestedFields.List {
		parseStructField(nestedFields.List[i])
	}

	tag := field.Tag

	var tagValue string
	if tag != nil {
		tagValue = tag.Value
	}

	return &godmt.ScannedStructField{
		Name:          field.Names[0].Name,
		Kind:          "struct",
		Tag:           tagValue,
		Doc:           godmt.ExtractComments(field.Doc),
		ImportDetails: nil,
		SubFields:     parsedNestedFields,
	}
}

func parseConstantsAndVariables(d *ast.Ident) []godmt.ScannedType {
	var result []godmt.ScannedType

	objectValues := reflect.ValueOf(d.Obj.Decl).Elem().FieldByName("Values")
	if !objectValues.IsValid() {
		return result
	}

	values := objectValues.Interface().([]ast.Expr)

	for i := range values {
		switch item := values[i].(type) {
		case *ast.BasicLit:
			parsed := godmt.BasicTypeLiteralParser(d, item)
			result = append(result, parsed)

		case *ast.Ident:
			parsed := godmt.IdentifierParser(d, item)

			if parsed != nil {
				result = append(result, *parsed)
			}

		case *ast.CompositeLit:
			switch itemType := item.Type.(type) {
			case *ast.MapType:
				mapElements := reflect.ValueOf(item.Elts).Interface().([]ast.Expr)
				parsed := godmt.CompositeLiteralMapParser(d, mapElements, item)
				result = append(result, parsed)
			case *ast.ArrayType:
				sliceType := godmt.GetMapValueType(itemType.Elt)
				parsed := godmt.CompositeLiteralSliceParser(d, sliceType, item)
				result = append(result, parsed)
			}
		}
	}

	return result
}
