// Copyright Infratographer, Inc. and/or licensed to Infratographer, Inc. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package entx

// EventsHookAnnotationName is the value of the annotation when read during ent compilation
var EventsHookAnnotationName = "INFRA9_EVENTHOOKS"
var NamespacedAnnotationName = "INFRA9_ENTX"

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

// Annotation provides a ent.Annotaion spec
type NamespacedAnnotation struct {
	IsNamespacedDataJSONField bool
}

// Name implements the ent Annotation interface.
func (a NamespacedAnnotation) Name() string {
	return NamespacedAnnotationName
}
