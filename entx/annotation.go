package entx

// AnnotationName is the value of the annotation when read during ent compilation
var AnnotationName = "I12R_ENTX"

// Annotation provides a ent.Annotaion spec
type Annotation struct {
	IsNamespacedDataJSONField bool
}

// Name implements the ent Annotation interface.
func (a Annotation) Name() string {
	return AnnotationName
}
