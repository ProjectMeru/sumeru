package sdk

import "sumeru/core/modelmeta"

// Model is embedded in addon model structs.
type Model = modelmeta.ModelMeta

// Any is a placeholder for Many2One when the target type cannot be imported.
type Any = modelmeta.Any

type (
	String            = modelmeta.String
	Text              = modelmeta.Text
	HTML              = modelmeta.HTML
	Email             = modelmeta.Email
	Phone             = modelmeta.Phone
	URL               = modelmeta.URL
	UUID              = modelmeta.UUID
	Boolean           = modelmeta.Boolean
	Integer           = modelmeta.Integer
	Float             = modelmeta.Float
	Float64           = modelmeta.Float64
	Numeric           = modelmeta.Numeric
	Money             = modelmeta.Money
	Date              = modelmeta.Date
	Time              = modelmeta.Time
	DateTime          = modelmeta.DateTime
	Duration          = modelmeta.Duration
	Json              = modelmeta.Json
	Properties        = modelmeta.Properties
	Binary            = modelmeta.Binary
	Image             = modelmeta.Image
	Reference         = modelmeta.Reference
	Many2OneReference = modelmeta.Many2OneReference
)

type Many2One[T any] = modelmeta.Many2One[T]
type One2Many[T any] = modelmeta.One2Many[T]
type Many2Many[T any] = modelmeta.Many2Many[T]
type Selection[T ~string] = modelmeta.Selection[T]
