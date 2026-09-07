/*
Copyright (c) 2025 Red Hat Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the
License. You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific
language governing permissions and limitations under the License.
*/

package rendering

import (
	"bytes"
	"context"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"

	privatev1 "github.com/osac-project/fulfillment-service/internal/api/osac/private/v1"
	"github.com/osac-project/fulfillment-service/internal/reflection"
)

var _ = Describe("Repeated enum rendering", func() {
	var ctrl *gomock.Controller

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		DeferCleanup(ctrl.Finish)
	})

	// renderDiskImages renders the given DiskImage objects using the table layout
	// that includes an ARCHITECTURE column typed as osac.private.v1.Architecture.
	renderDiskImages := func(ctx context.Context, items []*privatev1.DiskImage) string {
		// Create the object helper for the DiskImage type:
		objectDescriptor := (&privatev1.DiskImage{}).ProtoReflect().Descriptor()
		objectFullName := objectDescriptor.FullName()
		objectHelper := reflection.NewMockObjectHelper(ctrl)
		objectHelper.EXPECT().FullName().
			Return(objectFullName).
			AnyTimes()
		objectHelper.EXPECT().Descriptor().
			Return(objectDescriptor).
			AnyTimes()
		objectHelper.EXPECT().String().
			Return(string(objectFullName)).
			AnyTimes()
		objectHelper.EXPECT().IsTenantScoped().
			Return(false).
			AnyTimes()

		// Create the reflection helper:
		helper := reflection.NewMockHelper(ctrl)
		helper.EXPECT().
			Lookup(gomock.Any()).
			Return(objectHelper).
			AnyTimes()

		// Build the renderer:
		buffer := &bytes.Buffer{}
		renderer, err := NewTableRenderer().
			SetLogger(logger).
			SetHelper(helper).
			SetWriter(buffer).
			Build()
		Expect(err).ToNot(HaveOccurred())

		// Render the DiskImage objects:
		messages := make([]proto.Message, len(items))
		for i, item := range items {
			messages[i] = item
		}

		err = renderer.Render(ctx, messages)
		Expect(err).ToNot(HaveOccurred())
		return buffer.String()
	}

	It("Renders a single-value repeated enum as AMD64", func(ctx context.Context) {
		output := renderDiskImages(
			ctx,
			[]*privatev1.DiskImage{
				privatev1.DiskImage_builder{
					Id: "img-1",
					Metadata: privatev1.Metadata_builder{
						Name: "rhel-9",
					}.Build(),
					Spec: privatev1.DiskImageSpec_builder{
						Architectures: []privatev1.Architecture{
							privatev1.Architecture_ARCHITECTURE_AMD64,
						},
					}.Build(),
				}.Build(),
			},
		)
		Expect(output).To(ContainSubstring("AMD64"))
		Expect(output).ToNot(ContainSubstring("[1]"))
	})

	It("Renders a multi-value repeated enum as AMD64,ARM64", func(ctx context.Context) {
		output := renderDiskImages(
			ctx,
			[]*privatev1.DiskImage{
				privatev1.DiskImage_builder{
					Id: "img-2",
					Metadata: privatev1.Metadata_builder{
						Name: "rhel-9-multi",
					}.Build(),
					Spec: privatev1.DiskImageSpec_builder{
						Architectures: []privatev1.Architecture{
							privatev1.Architecture_ARCHITECTURE_AMD64,
							privatev1.Architecture_ARCHITECTURE_ARM64,
						},
					}.Build(),
				}.Build(),
			},
		)
		Expect(output).To(ContainSubstring("AMD64,ARM64"))
		Expect(output).ToNot(ContainSubstring("[1 2]"))
	})

	It("Renders an empty repeated enum as a placeholder", func(ctx context.Context) {
		output := renderDiskImages(
			ctx,
			[]*privatev1.DiskImage{
				privatev1.DiskImage_builder{
					Id: "img-3",
					Metadata: privatev1.Metadata_builder{
						Name: "no-arch",
					}.Build(),
					Spec: privatev1.DiskImageSpec_builder{}.Build(),
				}.Build(),
			},
		)
		// The ARCHITECTURE column should show "-" for empty lists
		Expect(output).To(MatchRegexp(`ARCHITECTURE.*\n.*-`))
	})
})
