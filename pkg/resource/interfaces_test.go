/*
Copyright 2021 The Crossplane Authors.

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

package resource

import (
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource/fake"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource/unstructured/claim"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource/unstructured/composed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource/unstructured/composite"
)

// We test that our fakes satisfy our interfaces here rather than in the fake
// package to avoid a cyclic dependency.

var (
	_ Managed             = &fake.Managed{}
	_ ProviderConfig      = &fake.ProviderConfig{}
	_ ProviderConfigUsage = &fake.ProviderConfigUsage{}
	_ ProviderConfigUsage = &fake.LegacyProviderConfigUsage{}

	_ ModernManaged             = &fake.ModernManaged{}
	_ TypedProviderConfigUsage  = &fake.ProviderConfigUsage{}
	_ LegacyManaged             = &fake.LegacyManaged{}
	_ LegacyProviderConfigUsage = &fake.LegacyProviderConfigUsage{}

	_ CompositeClaim = &fake.CompositeClaim{}
	_ Composite      = &fake.Composite{}
	_ Composed       = &fake.Composed{}

	_ CompositeClaim = &claim.Unstructured{}
	_ Composite      = &composite.Unstructured{}
	_ Composed       = &composed.Unstructured{}
)
