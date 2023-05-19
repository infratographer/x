package entx

// PubsubAnnotationName is the value of the annotation when read during ent compilation
var PubsubAnnotationName = "INFRA9_PUBSUBHOOK"

// PubsubAnnotation provides a ent.Annotation spec. These shouldn't be set directly, you should use PubsubAdditionalSubject() and PubsubSubjectName instead
type PubsubAnnotation struct {
	SubjectName              string
	IsAdditionalSubjectField bool
}

// Name implements the ent Annotation interface.
func (a PubsubAnnotation) Name() string {
	return PubsubAnnotationName
}

// PubsubAdditionalSubject marks this field as a field to return as an additional subject
func PubsubAdditionalSubject() *PubsubAnnotation {
	return &PubsubAnnotation{
		IsAdditionalSubjectField: true,
	}
}

// PubsubSubjectName sets the subject name that is where the messages for this object will be sent
func PubsubSubjectName(s string) *PubsubAnnotation {
	return &PubsubAnnotation{
		SubjectName: s,
	}
}
