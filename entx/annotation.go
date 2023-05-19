package entx

// EventsHookAnnotationName is the value of the annotation when read during ent compilation
var EventsHookAnnotationName = "INFRA9_EVENTHOOKS"

// EventsHookAnnotation provides a ent.Annotation spec. These shouldn't be set directly, you should use EventsHookAdditionalSubject() and EventsHookSubjectName instead
type EventsHookAnnotation struct {
	SubjectName              string
	IsAdditionalSubjectField bool
}

// Name implements the ent Annotation interface.
func (a EventsHookAnnotation) Name() string {
	return EventsHookAnnotationName
}

// EventsHookAdditionalSubject marks this field as a field to return as an additional subject
func EventsHookAdditionalSubject() *EventsHookAnnotation {
	return &EventsHookAnnotation{
		IsAdditionalSubjectField: true,
	}
}

// EventsHookSubjectName sets the subject name that is where the messages for this object will be sent
func EventsHookSubjectName(s string) *EventsHookAnnotation {
	return &EventsHookAnnotation{
		SubjectName: s,
	}
}
