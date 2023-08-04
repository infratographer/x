// Copyright 2023 The Infratographer Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package entx

// EventsHookAnnotationName is the value of the annotation when read during ent compilation
var EventsHookAnnotationName = "INFRA9_EVENTHOOKS"

// EventsHookAnnotation provides a ent.Annotation spec. These shouldn't be set directly, you should use EventsHookAdditionalSubject() and EventsHookSubjectName instead
type EventsHookAnnotation struct {
	SubjectName               string
	AdditionalSubjectRelation string
}

// Name implements the ent Annotation interface.
func (a EventsHookAnnotation) Name() string {
	return EventsHookAnnotationName
}

// EventsHookAdditionalSubject marks this field as a field to return as an additional subject
func EventsHookAdditionalSubject(relation string) *EventsHookAnnotation {
	return &EventsHookAnnotation{
		AdditionalSubjectRelation: relation,
	}
}

// EventsHookSubjectName sets the subject name that is where the messages for this object will be sent
func EventsHookSubjectName(s string) *EventsHookAnnotation {
	return &EventsHookAnnotation{
		SubjectName: s,
	}
}
