package data

// ScannedType represents a basic entity to be translated.
// More specifically const and var items.
type ScannedType struct {
	Name         string      `json:"name"`
	Kind         string      `json:"kind"`
	Value        interface{} `json:"value"`
	Doc          []string    `json:"doc"`
	InternalType int         `json:"internalType"`
}

type ExtendedScannedType struct {
	ScannedType
	AnExtraField bool `json:"anExtraField"`
}

// ScannedStruct represents the details of a scanned struct.
type ScannedStruct struct {
	Doc          []string `json:"doc" binding:"required" validation:"required"`
	Name         string   `json:"name" binding:"required" validation:"required"`
	Fields       []bool   `json:"fields" binding:"required" validation:"required"`
	InternalType int      `xml:"internalType" binding:"required" validation:"required"`
}

type EmbeddedType struct {
	ID   string `json:"id"`
	Data struct {
		Test string `json:"test"`
	} `json:"data"`
}