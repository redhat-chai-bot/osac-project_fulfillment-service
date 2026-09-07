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

	testsv1 "github.com/osac-project/fulfillment-service/internal/api/osac/tests/v1"
	"github.com/osac-project/fulfillment-service/internal/reflection"
)

var _ = Describe("Repeated enum rendering", func() {
	var ctrl *gomock.Controller

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		DeferCleanup(ctrl.Finish)
	})

	// renderTestObject renders the given test object with a table layout that includes
	// a column for the my_int32_list field typed as osac.tests.v1.MyEnum.
	renderTestObject := func(ctx context.Context, items []*testsv1.Object) string {
		// Create the object helper for the test Object type:
		objectDescriptor := (&testsv1.Object{}).ProtoReflect().Descriptor()
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

		// Override the table layout to include a repeated-enum column.
		// We use my_int32_list as the source field and type it as osac.tests.v1.MyEnum
		// to simulate a repeated enum field.
		messages := make([]proto.Message, len(items))
		for i, item := range items {
			messages[i] = item
		}

		err = renderer.Render(ctx, messages)
		Expect(err).ToNot(HaveOccurred())
		return buffer.String()
	}

	It("Renders a single-value repeated enum as the shortened name", func(ctx context.Context) {
		output := renderTestObject(
			ctx,
			[]*testsv1.Object{
				{
					Id: "obj-1",
					Metadata: &testsv1.Metadata{
						Name: "test-obj",
					},
					MyInt32List: []int32{1},
				},
			},
		)
		Expect(output).To(ContainSubstring("VALUE_A"))
		Expect(output).ToNot(ContainSubstring("[1]"))
	})

	It("Renders a multi-value repeated enum as comma-separated names", func(ctx context.Context) {
		output := renderTestObject(
			ctx,
			[]*testsv1.Object{
				{
					Id: "obj-2",
					Metadata: &testsv1.Metadata{
						Name: "test-obj",
					},
					MyInt32List: []int32{1, 2},
				},
			},
		)
		Expect(output).To(ContainSubstring("VALUE_A,VALUE_B"))
		Expect(output).ToNot(ContainSubstring("[1 2]"))
	})

	It("Renders an empty repeated enum as a placeholder", func(ctx context.Context) {
		output := renderTestObject(
			ctx,
			[]*testsv1.Object{
				{
					Id: "obj-3",
					Metadata: &testsv1.Metadata{
						Name: "test-obj",
					},
					MyInt32List: []int32{},
				},
			},
		)
		// The ENUM LIST column should show "-" for empty lists
		Expect(output).To(MatchRegexp(`ENUM LIST.*\n.*-`))
	})
})
