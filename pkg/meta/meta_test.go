/*
Copyright 2019 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package meta

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/errors"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
)

const (
	group        = "coolstuff"
	version      = "v1"
	groupVersion = group + "/" + version
	kind         = "coolresource"
	namespace    = "coolns"
	name         = "cool"
	uid          = types.UID("definitely-a-uuid")
)

func TestReferenceTo(t *testing.T) {
	type args struct {
		o  metav1.Object
		of schema.GroupVersionKind
	}

	tests := map[string]struct {
		args
		want *corev1.ObjectReference
	}{
		"WithTypeMeta": {
			args: args{
				o: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: namespace,
						Name:      name,
						UID:       uid,
					},
				},
				of: schema.GroupVersionKind{
					Group:   group,
					Version: version,
					Kind:    kind,
				},
			},
			want: &corev1.ObjectReference{
				APIVersion: groupVersion,
				Kind:       kind,
				Namespace:  namespace,
				Name:       name,
				UID:        uid,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := ReferenceTo(tc.o, tc.of)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ReferenceTo(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestTypedReferenceTo(t *testing.T) {
	type args struct {
		o  metav1.Object
		of schema.GroupVersionKind
	}

	tests := map[string]struct {
		args
		want *xpv1.TypedReference
	}{
		"WithTypeMeta": {
			args: args{
				o: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: namespace,
						Name:      name,
						UID:       uid,
					},
				},
				of: schema.GroupVersionKind{
					Group:   group,
					Version: version,
					Kind:    kind,
				},
			},
			want: &xpv1.TypedReference{
				APIVersion: groupVersion,
				Kind:       kind,
				Name:       name,
				UID:        uid,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := TypedReferenceTo(tc.o, tc.of)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("TypedReferenceTo(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestAsOwner(t *testing.T) {
	tests := map[string]struct {
		r    *xpv1.TypedReference
		want metav1.OwnerReference
	}{
		"Successful": {
			r: &xpv1.TypedReference{
				APIVersion: groupVersion,
				Kind:       kind,
				Name:       name,
				UID:        uid,
			},
			want: metav1.OwnerReference{
				APIVersion: groupVersion,
				Kind:       kind,
				Name:       name,
				UID:        uid,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := AsOwner(tc.r)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("AsOwner(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestAsController(t *testing.T) {
	flag := true

	tests := map[string]struct {
		r    *xpv1.TypedReference
		want metav1.OwnerReference
	}{
		"Successful": {
			r: &xpv1.TypedReference{
				APIVersion: groupVersion,
				Kind:       kind,
				Name:       name,
				UID:        uid,
			},
			want: metav1.OwnerReference{
				APIVersion:         groupVersion,
				Kind:               kind,
				Name:               name,
				UID:                uid,
				Controller:         &flag,
				BlockOwnerDeletion: &flag,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := AsController(tc.r)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("AsController(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestHaveSameController(t *testing.T) {
	controller := true

	controllerA := metav1.OwnerReference{
		UID:        uid,
		Controller: &controller,
	}

	controllerB := metav1.OwnerReference{
		UID:        types.UID("a-different-uuid"),
		Controller: &controller,
	}

	cases := map[string]struct {
		a    metav1.Object
		b    metav1.Object
		want bool
	}{
		"SameController": {
			a: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{controllerA},
				},
			},
			b: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{controllerA},
				},
			},
			want: true,
		},
		"AHasNoController": {
			a: &corev1.Pod{},
			b: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{controllerB},
				},
			},
			want: false,
		},
		"BHasNoController": {
			a: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{controllerA},
				},
			},
			b:    &corev1.Pod{},
			want: false,
		},
		"ControllersDiffer": {
			a: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{controllerA},
				},
			},
			b: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{controllerB},
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := HaveSameController(tc.a, tc.b)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("HaveSameController(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestNamespacedNameOf(t *testing.T) {
	cases := map[string]struct {
		r    *corev1.ObjectReference
		want types.NamespacedName
	}{
		"Success": {
			r:    &corev1.ObjectReference{Namespace: namespace, Name: name},
			want: types.NamespacedName{Namespace: namespace, Name: name},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := NamespacedNameOf(tc.r)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("NamespacedNameOf(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestAddOwnerReference(t *testing.T) {
	owner := metav1.OwnerReference{UID: uid}
	other := metav1.OwnerReference{UID: "a-different-uuid"}
	ctrlr := metav1.OwnerReference{UID: uid, Controller: func() *bool { c := true; return &c }()}

	type args struct {
		o metav1.Object
		r metav1.OwnerReference
	}

	cases := map[string]struct {
		args args
		want []metav1.OwnerReference
	}{
		"NoExistingOwners": {
			args: args{
				o: &corev1.Pod{},
				r: owner,
			},
			want: []metav1.OwnerReference{owner},
		},
		"UpdateExistingOwner": {
			args: args{
				o: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						OwnerReferences: []metav1.OwnerReference{ctrlr},
					},
				},
				r: owner,
			},
			want: []metav1.OwnerReference{owner},
		},
		"OwnedByAnotherObject": {
			args: args{
				o: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						OwnerReferences: []metav1.OwnerReference{other},
					},
				},
				r: owner,
			},
			want: []metav1.OwnerReference{other, owner},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			AddOwnerReference(tc.args.o, tc.args.r)

			got := tc.args.o.GetOwnerReferences()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("tc.args.o.GetOwnerReferences(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestAddControllerReference(t *testing.T) {
	owner := metav1.OwnerReference{UID: uid}
	other := metav1.OwnerReference{UID: "a-different-uuid"}
	ctrlr := metav1.OwnerReference{UID: uid, Controller: func() *bool { c := true; return &c }()}
	otrlr := metav1.OwnerReference{
		Kind:       "lame",
		Name:       "othercontroller",
		UID:        "a-different-uuid",
		Controller: func() *bool { c := true; return &c }(),
	}

	type args struct {
		o metav1.Object
		r metav1.OwnerReference
	}

	type want struct {
		owners []metav1.OwnerReference
		err    error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"NoExistingOwners": {
			args: args{
				o: &corev1.Pod{},
				r: owner,
			},
			want: want{
				owners: []metav1.OwnerReference{owner},
			},
		},
		"UpdateExistingOwner": {
			args: args{
				o: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						OwnerReferences: []metav1.OwnerReference{ctrlr},
					},
				},
				r: owner,
			},
			want: want{
				owners: []metav1.OwnerReference{owner},
			},
		},
		"OwnedByAnotherObject": {
			args: args{
				o: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						OwnerReferences: []metav1.OwnerReference{other},
					},
				},
				r: owner,
			},
			want: want{
				owners: []metav1.OwnerReference{other, owner},
			},
		},
		"ControlledByAnotherObject": {
			args: args{
				o: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:            name,
						OwnerReferences: []metav1.OwnerReference{otrlr},
					},
				},
				r: owner,
			},
			want: want{
				owners: []metav1.OwnerReference{otrlr},
				err:    errors.Errorf("%s is already controlled by %s %s (UID %s)", name, otrlr.Kind, otrlr.Name, otrlr.UID),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := AddControllerReference(tc.args.o, tc.args.r)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("AddControllerReference(...): -want error, +got error:\n%s", diff)
			}

			got := tc.args.o.GetOwnerReferences()
			if diff := cmp.Diff(tc.want.owners, got); diff != "" {
				t.Errorf("tc.args.o.GetOwnerReferences(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestAddFinalizer(t *testing.T) {
	finalizer := "fin"
	funalizer := "fun"

	type args struct {
		o         metav1.Object
		finalizer string
	}

	cases := map[string]struct {
		args args
		want []string
	}{
		"NoExistingFinalizers": {
			args: args{
				o:         &corev1.Pod{},
				finalizer: finalizer,
			},
			want: []string{finalizer},
		},
		"FinalizerAlreadyExists": {
			args: args{
				o: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Finalizers: []string{finalizer},
					},
				},
				finalizer: finalizer,
			},
			want: []string{finalizer},
		},
		"AnotherFinalizerExists": {
			args: args{
				o: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Finalizers: []string{funalizer},
					},
				},
				finalizer: finalizer,
			},
			want: []string{funalizer, finalizer},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			AddFinalizer(tc.args.o, tc.args.finalizer)

			got := tc.args.o.GetFinalizers()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("tc.args.o.GetFinalizers(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestRemoveFinalizer(t *testing.T) {
	finalizer := "fin"
	funalizer := "fun"

	type args struct {
		o         metav1.Object
		finalizer string
	}

	cases := map[string]struct {
		args args
		want []string
	}{
		"NoExistingFinalizers": {
			args: args{
				o:         &corev1.Pod{},
				finalizer: finalizer,
			},
			want: nil,
		},
		"FinalizerExists": {
			args: args{
				o: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Finalizers: []string{finalizer},
					},
				},
				finalizer: finalizer,
			},
			want: []string{},
		},
		"AnotherFinalizerExists": {
			args: args{
				o: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Finalizers: []string{finalizer, funalizer},
					},
				},
				finalizer: finalizer,
			},
			want: []string{funalizer},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			RemoveFinalizer(tc.args.o, tc.args.finalizer)

			got := tc.args.o.GetFinalizers()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("tc.args.o.GetFinalizers(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestFinalizerExists(t *testing.T) {
	finalizer := "fin"
	funalizer := "fun"

	type args struct {
		o         metav1.Object
		finalizer string
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"NoExistingFinalizers": {
			args: args{
				o:         &corev1.Pod{},
				finalizer: finalizer,
			},
			want: false,
		},
		"FinalizerExists": {
			args: args{
				o: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Finalizers: []string{finalizer},
					},
				},
				finalizer: finalizer,
			},
			want: true,
		},
		"AnotherFinalizerExists": {
			args: args{
				o: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Finalizers: []string{funalizer},
					},
				},
				finalizer: finalizer,
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if diff := cmp.Diff(tc.want, FinalizerExists(tc.args.o, tc.args.finalizer)); diff != "" {
				t.Errorf("tc.args.o.GetFinalizers(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestAddLabels(t *testing.T) {
	key, value := "key", "value"
	existingKey, existingValue := "ekey", "evalue"

	type args struct {
		o      metav1.Object
		labels map[string]string
	}

	cases := map[string]struct {
		args args
		want map[string]string
	}{
		"ExistingLabels": {
			args: args{
				o: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							existingKey: existingValue,
						},
					},
				},
				labels: map[string]string{key: value},
			},
			want: map[string]string{
				existingKey: existingValue,
				key:         value,
			},
		},
		"NoExistingLabels": {
			args: args{
				o:      &corev1.Pod{},
				labels: map[string]string{key: value},
			},
			want: map[string]string{key: value},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			AddLabels(tc.args.o, tc.args.labels)

			got := tc.args.o.GetLabels()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("tc.args.o.GetLabels(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestRemoveLabels(t *testing.T) {
	keyA, valueA := "keyA", "valueA"
	keyB, valueB := "keyB", "valueB"

	type args struct {
		o      metav1.Object
		labels []string
	}

	cases := map[string]struct {
		args args
		want map[string]string
	}{
		"ExistingLabels": {
			args: args{
				o: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							keyA: valueA,
							keyB: valueB,
						},
					},
				},
				labels: []string{keyA},
			},
			want: map[string]string{keyB: valueB},
		},
		"NoExistingLabels": {
			args: args{
				o:      &corev1.Pod{},
				labels: []string{keyA},
			},
			want: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			RemoveLabels(tc.args.o, tc.args.labels...)

			got := tc.args.o.GetLabels()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("tc.args.o.GetLabels(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestAddAnnotations(t *testing.T) {
	key, value := "key", "value"
	existingKey, existingValue := "ekey", "evalue"

	type args struct {
		o           metav1.Object
		annotations map[string]string
	}

	cases := map[string]struct {
		args args
		want map[string]string
	}{
		"ExistingAnnotations": {
			args: args{
				o: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							existingKey: existingValue,
						},
					},
				},
				annotations: map[string]string{key: value},
			},
			want: map[string]string{
				existingKey: existingValue,
				key:         value,
			},
		},
		"NoExistingAnnotations": {
			args: args{
				o:           &corev1.Pod{},
				annotations: map[string]string{key: value},
			},
			want: map[string]string{key: value},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			AddAnnotations(tc.args.o, tc.args.annotations)

			got := tc.args.o.GetAnnotations()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("tc.args.o.GetAnnotations(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestRemoveAnnotations(t *testing.T) {
	keyA, valueA := "keyA", "valueA"
	keyB, valueB := "keyB", "valueB"

	type args struct {
		o           metav1.Object
		annotations []string
	}

	cases := map[string]struct {
		args args
		want map[string]string
	}{
		"ExistingAnnotations": {
			args: args{
				o: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							keyA: valueA,
							keyB: valueB,
						},
					},
				},
				annotations: []string{keyA},
			},
			want: map[string]string{keyB: valueB},
		},
		"NoExistingAnnotations": {
			args: args{
				o:           &corev1.Pod{},
				annotations: []string{keyA},
			},
			want: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			RemoveAnnotations(tc.args.o, tc.args.annotations...)

			got := tc.args.o.GetAnnotations()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("tc.args.o.GetAnnotations(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestWasDeleted(t *testing.T) {
	now := metav1.Now()

	cases := map[string]struct {
		o    metav1.Object
		want bool
	}{
		"ObjectWasDeleted": {
			o:    &corev1.Pod{ObjectMeta: metav1.ObjectMeta{DeletionTimestamp: &now}},
			want: true,
		},
		"ObjectWasNotDeleted": {
			o:    &corev1.Pod{ObjectMeta: metav1.ObjectMeta{DeletionTimestamp: nil}},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := WasDeleted(tc.o)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("WasDeleted(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestWasCreated(t *testing.T) {
	now := metav1.Now()
	zero := metav1.Time{}

	cases := map[string]struct {
		o    metav1.Object
		want bool
	}{
		"ObjectWasCreated": {
			o:    &corev1.Pod{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: now}},
			want: true,
		},
		"ObjectWasNotCreated": {
			o:    &corev1.Pod{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: zero}},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := WasCreated(tc.o)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("WasCreated(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGetExternalName(t *testing.T) {
	cases := map[string]struct {
		o    metav1.Object
		want string
	}{
		"ExternalNameExists": {
			o:    &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{AnnotationKeyExternalName: name}}},
			want: name,
		},
		"NoExternalName": {
			o:    &corev1.Pod{},
			want: "",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GetExternalName(tc.o)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GetExternalName(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestSetExternalName(t *testing.T) {
	cases := map[string]struct {
		o    metav1.Object
		name string
		want metav1.Object
	}{
		"SetsTheCorrectKey": {
			o:    &corev1.Pod{},
			name: name,
			want: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{AnnotationKeyExternalName: name}}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			SetExternalName(tc.o, tc.name)

			if diff := cmp.Diff(tc.want, tc.o); diff != "" {
				t.Errorf("SetExternalName(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGetExternalCreatePending(t *testing.T) {
	now := time.Now().Round(time.Second)

	cases := map[string]struct {
		o    metav1.Object
		want time.Time
	}{
		"ExternalCreatePendingExists": {
			o:    &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{AnnotationKeyExternalCreatePending: now.Format(time.RFC3339)}}},
			want: now,
		},
		"NoExternalCreatePending": {
			o:    &corev1.Pod{},
			want: time.Time{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GetExternalCreatePending(tc.o)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GetExternalCreatePending(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestSetExternalCreatePending(t *testing.T) {
	now := time.Now()

	cases := map[string]struct {
		o    metav1.Object
		t    time.Time
		want metav1.Object
	}{
		"SetsTheCorrectKey": {
			o:    &corev1.Pod{},
			t:    now,
			want: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{AnnotationKeyExternalCreatePending: now.Format(time.RFC3339)}}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			SetExternalCreatePending(tc.o, tc.t)

			if diff := cmp.Diff(tc.want, tc.o); diff != "" {
				t.Errorf("SetExternalCreatePending(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGetExternalCreateSucceeded(t *testing.T) {
	now := time.Now().Round(time.Second)

	cases := map[string]struct {
		o    metav1.Object
		want time.Time
	}{
		"ExternalCreateTimeExists": {
			o:    &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{AnnotationKeyExternalCreateSucceeded: now.Format(time.RFC3339)}}},
			want: now,
		},
		"NoExternalCreateTime": {
			o:    &corev1.Pod{},
			want: time.Time{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GetExternalCreateSucceeded(tc.o)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GetExternalCreateSucceeded(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestSetExternalCreateSucceeded(t *testing.T) {
	now := time.Now()

	cases := map[string]struct {
		o    metav1.Object
		t    time.Time
		want metav1.Object
	}{
		"SetsTheCorrectKey": {
			o:    &corev1.Pod{},
			t:    now,
			want: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{AnnotationKeyExternalCreateSucceeded: now.Format(time.RFC3339)}}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			SetExternalCreateSucceeded(tc.o, tc.t)

			if diff := cmp.Diff(tc.want, tc.o); diff != "" {
				t.Errorf("SetExternalCreateSucceeded(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGetExternalCreateFailed(t *testing.T) {
	now := time.Now().Round(time.Second)

	cases := map[string]struct {
		o    metav1.Object
		want time.Time
	}{
		"ExternalCreateFailedExists": {
			o:    &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{AnnotationKeyExternalCreateFailed: now.Format(time.RFC3339)}}},
			want: now,
		},
		"NoExternalCreateFailed": {
			o:    &corev1.Pod{},
			want: time.Time{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GetExternalCreateFailed(tc.o)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GetExternalCreateFailed(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestSetExternalCreateFailed(t *testing.T) {
	now := time.Now()

	cases := map[string]struct {
		o    metav1.Object
		t    time.Time
		want metav1.Object
	}{
		"SetsTheCorrectKey": {
			o:    &corev1.Pod{},
			t:    now,
			want: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{AnnotationKeyExternalCreateFailed: now.Format(time.RFC3339)}}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			SetExternalCreateFailed(tc.o, tc.t)

			if diff := cmp.Diff(tc.want, tc.o); diff != "" {
				t.Errorf("SetExternalCreateFailed(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestExternalCreateSucceededDuring(t *testing.T) {
	type args struct {
		o metav1.Object
		d time.Duration
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"NotYetSuccessfullyCreated": {
			args: args{
				o: &corev1.Pod{},
				d: 1 * time.Minute,
			},
			want: false,
		},
		"SuccessfullyCreatedTooLongAgo": {
			args: args{
				o: func() metav1.Object {
					o := &corev1.Pod{}
					t := time.Now().Add(-2 * time.Minute)
					SetExternalCreateSucceeded(o, t)
					return o
				}(),
				d: 1 * time.Minute,
			},
			want: false,
		},
		"SuccessfullyCreatedWithinDuration": {
			args: args{
				o: func() metav1.Object {
					o := &corev1.Pod{}
					t := time.Now().Add(-30 * time.Second)
					SetExternalCreateSucceeded(o, t)
					return o
				}(),
				d: 1 * time.Minute,
			},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := ExternalCreateSucceededDuring(tc.args.o, tc.args.d)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ExternalCreateSucceededDuring(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestExternalCreateIncomplete(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	earlier := time.Now().Add(-1 * time.Second).Format(time.RFC3339)
	evenEarlier := time.Now().Add(-1 * time.Minute).Format(time.RFC3339)

	cases := map[string]struct {
		reason string
		o      metav1.Object
		want   bool
	}{
		"CreateNeverPending": {
			reason: "If we've never called Create it can't be incomplete.",
			o:      &corev1.Pod{},
			want:   false,
		},
		"CreateSucceeded": {
			reason: "If Create succeeded since it was pending, it's complete.",
			o: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				AnnotationKeyExternalCreateFailed:    evenEarlier,
				AnnotationKeyExternalCreatePending:   earlier,
				AnnotationKeyExternalCreateSucceeded: now,
			}}},
			want: false,
		},
		"CreateFailed": {
			reason: "If Create failed since it was pending, it's complete.",
			o: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				AnnotationKeyExternalCreateSucceeded: evenEarlier,
				AnnotationKeyExternalCreatePending:   earlier,
				AnnotationKeyExternalCreateFailed:    now,
			}}},
			want: false,
		},
		"CreateNeverCompleted": {
			reason: "If Create was pending but never succeeded or failed, it's incomplete.",
			o: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				AnnotationKeyExternalCreatePending: earlier,
			}}},
			want: true,
		},
		"RecreateNeverCompleted": {
			reason: "If Create is pending and there's an older success we're probably trying to recreate a deleted external resource, and it's incomplete.",
			o: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				AnnotationKeyExternalCreateSucceeded: earlier,
				AnnotationKeyExternalCreatePending:   now,
			}}},
			want: true,
		},
		"RetryNeverCompleted": {
			reason: "If Create is pending and there's an older failure we're probably trying to recreate a deleted external resource, and it's incomplete.",
			o: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				AnnotationKeyExternalCreateFailed:  earlier,
				AnnotationKeyExternalCreatePending: now,
			}}},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := ExternalCreateIncomplete(tc.o)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ExternalCreateIncomplete(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsPaused(t *testing.T) {
	cases := map[string]struct {
		o    metav1.Object
		want bool
	}{
		"HasPauseAnnotationSetTrue": {
			o: func() metav1.Object {
				p := &corev1.Pod{}
				p.SetAnnotations(map[string]string{
					AnnotationKeyReconciliationPaused: "true",
				})
				return p
			}(),
			want: true,
		},
		"NoPauseAnnotation": {
			o:    &corev1.Pod{},
			want: false,
		},
		"HasEmptyPauseAnnotation": {
			o: func() metav1.Object {
				p := &corev1.Pod{}
				p.SetAnnotations(map[string]string{
					AnnotationKeyReconciliationPaused: "",
				})
				return p
			}(),
			want: false,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsPaused(tc.o)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("IsPaused(...): -want, +got:\n%s", diff)
			}
		})
	}
}
