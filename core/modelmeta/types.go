package modelmeta

// ModelMeta is embedded in model structs for schema registration.
type ModelMeta struct{}

// Any is a placeholder type parameter for cross-package relations.
type Any struct {
	ModelMeta `sumeru:"model=-"`
}

// Scalar and relational field markers (zero-size).
type (
	String            struct{}
	Text              struct{}
	HTML              struct{}
	Email             struct{}
	Phone             struct{}
	URL               struct{}
	UUID              struct{}
	Boolean           struct{}
	Integer           struct{}
	Float             struct{}
	Float64           struct{}
	Numeric           struct{}
	Money             struct{}
	Date              struct{}
	Time              struct{}
	DateTime          struct{}
	Duration          struct{}
	Json              struct{}
	Properties        struct{}
	Binary            struct{}
	Image             struct{}
	Reference         struct{}
	Many2OneReference struct{}
)

type Many2One[T any] struct{}
type One2Many[T any] struct{}
type Many2Many[T any] struct{}

// Selection is a type-safe selection marker. Options are auto-discovered from typed
// const blocks in the same package, or set explicitly with a selection= struct tag.
type Selection[T ~string] struct{}
