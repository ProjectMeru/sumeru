package orm

type FieldType string

const (
	Char      FieldType = "char"
	Text      FieldType = "text"
	Integer   FieldType = "integer"
	Float     FieldType = "float"
	Float64   FieldType = "float64"
	Numeric   FieldType = "numeric" // For exact decimal precision (money)
	Boolean   FieldType = "boolean"
	Date      FieldType = "date"
	DateTime  FieldType = "datetime"
	Selection FieldType = "selection"
	Many2One  FieldType = "many2one"
	Many2Many FieldType = "many2many"
	One2Many  FieldType = "one2many"
	Json      FieldType = "json"
)

type FieldDefinition struct {
	Name               string
	Type               FieldType
	Required           bool
	Relation           string // For Many2One, Many2many, One2many
	RelationTable      string // For Many2many
	Column1            string // For Many2many (this model's FK in rel table)
	Column2            string // For Many2many (target model's FK in rel table)
	String             string // Label
	Help               string
	DefaultVal         interface{}
	Selection          [][]string // Options for Selection type: {{"key", "Value"}, ...}
	Unique             bool
	Index              bool   // Generate database index
	Size               int    // VARCHAR length for Char
	Precision          int    // NUMERIC precision
	Scale              int    // NUMERIC scale
	OnDelete           string // FK on delete policy (metadata)
	Readonly           bool
	Column             string // Explicit SQL column name override
	RelationModelField string // Many2OneReference: field holding target model name
	Widget             string // UI widget hint (email, phone, monetary, html, …)
	Min                *float64
	Max                *float64
	Currency           string // Money: sibling Many2One field name
	Domain             string // Many2One domain filter
	Groups             string // Field-level group XML id
	Related            string // related=relation.field path
	RelatedStore       bool   // store related value in DB
	Compute            string // compute handler name
	ComputeStore       bool   // persist computed value
	Virtual            bool   // no SQL column (related/compute without store)
}

type Model interface {
	ModelName() string
	Fields() []FieldDefinition
}
