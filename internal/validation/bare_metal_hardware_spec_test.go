/*
Copyright (c) 2025 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the
License. You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific
language governing permissions and limitations under the License.
*/

package validation

import (
	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	privatev1 "github.com/osac-project/fulfillment-service/internal/api/osac/private/v1"
	publicv1 "github.com/osac-project/fulfillment-service/internal/api/osac/public/v1"
)

var _ = Describe("BareMetalHardwareSpec validation", func() {
	var validator protovalidate.Validator

	BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		Expect(err).ToNot(HaveOccurred())
	})

	// validCPU returns a valid BareMetalCPUSpec for use in test helpers.
	validPrivateCPU := func() *privatev1.BareMetalCPUSpec {
		return privatev1.BareMetalCPUSpec_builder{
			Cores:          4,
			Architecture:   "x86_64",
			ThreadsPerCore: 2,
		}.Build()
	}

	validPrivateMemory := func() *privatev1.BareMetalMemorySpec {
		return privatev1.BareMetalMemorySpec_builder{
			TotalGb: 64,
		}.Build()
	}

	validPublicCPU := func() *publicv1.BareMetalCPUSpec {
		return publicv1.BareMetalCPUSpec_builder{
			Cores:          4,
			Architecture:   "x86_64",
			ThreadsPerCore: 2,
		}.Build()
	}

	validPublicMemory := func() *publicv1.BareMetalMemorySpec {
		return publicv1.BareMetalMemorySpec_builder{
			TotalGb: 64,
		}.Build()
	}

	Describe("Private API", func() {
		It("Accepts hardware spec with a fabric port", func() {
			spec := privatev1.BareMetalHardwareSpec_builder{
				Cpu:    validPrivateCPU(),
				Memory: validPrivateMemory(),
				NetworkPorts: []*privatev1.BareMetalNetworkPortSpec{
					privatev1.BareMetalNetworkPortSpec_builder{
						Name:  "data-0",
						Role:  "fabric",
						Type:  "Ethernet",
						Speed: "100Gbps",
					}.Build(),
				},
			}.Build()
			err := validator.Validate(spec)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Accepts hardware spec with multiple ports including fabric", func() {
			spec := privatev1.BareMetalHardwareSpec_builder{
				Cpu:    validPrivateCPU(),
				Memory: validPrivateMemory(),
				NetworkPorts: []*privatev1.BareMetalNetworkPortSpec{
					privatev1.BareMetalNetworkPortSpec_builder{
						Name:  "data-0",
						Role:  "fabric",
						Type:  "Ethernet",
						Speed: "100Gbps",
					}.Build(),
					privatev1.BareMetalNetworkPortSpec_builder{
						Name:  "mgmt-0",
						Role:  "management",
						Type:  "Ethernet",
						Speed: "1Gbps",
					}.Build(),
					privatev1.BareMetalNetworkPortSpec_builder{
						Name:  "bmc-0",
						Role:  "lifecycle",
						Type:  "Ethernet",
						Speed: "1Gbps",
					}.Build(),
				},
			}.Build()
			err := validator.Validate(spec)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Accepts hardware spec with multiple fabric ports", func() {
			spec := privatev1.BareMetalHardwareSpec_builder{
				Cpu:    validPrivateCPU(),
				Memory: validPrivateMemory(),
				NetworkPorts: []*privatev1.BareMetalNetworkPortSpec{
					privatev1.BareMetalNetworkPortSpec_builder{
						Name:  "data-0",
						Role:  "fabric",
						Type:  "Ethernet",
						Speed: "100Gbps",
					}.Build(),
					privatev1.BareMetalNetworkPortSpec_builder{
						Name:  "data-1",
						Role:  "fabric",
						Type:  "Ethernet",
						Speed: "100Gbps",
					}.Build(),
				},
			}.Build()
			err := validator.Validate(spec)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Rejects hardware spec with no network ports", func() {
			spec := privatev1.BareMetalHardwareSpec_builder{
				Cpu:    validPrivateCPU(),
				Memory: validPrivateMemory(),
			}.Build()
			err := validator.Validate(spec)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("at least one network port with role 'fabric' is required"))
		})

		It("Rejects hardware spec with only non-fabric ports", func() {
			spec := privatev1.BareMetalHardwareSpec_builder{
				Cpu:    validPrivateCPU(),
				Memory: validPrivateMemory(),
				NetworkPorts: []*privatev1.BareMetalNetworkPortSpec{
					privatev1.BareMetalNetworkPortSpec_builder{
						Name:  "mgmt-0",
						Role:  "management",
						Type:  "Ethernet",
						Speed: "1Gbps",
					}.Build(),
					privatev1.BareMetalNetworkPortSpec_builder{
						Name:  "bmc-0",
						Role:  "lifecycle",
						Type:  "Ethernet",
						Speed: "1Gbps",
					}.Build(),
				},
			}.Build()
			err := validator.Validate(spec)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("at least one network port with role 'fabric' is required"))
		})

		It("Rejects hardware spec with single non-fabric port", func() {
			spec := privatev1.BareMetalHardwareSpec_builder{
				Cpu:    validPrivateCPU(),
				Memory: validPrivateMemory(),
				NetworkPorts: []*privatev1.BareMetalNetworkPortSpec{
					privatev1.BareMetalNetworkPortSpec_builder{
						Name:  "stor-0",
						Role:  "storage",
						Type:  "InfiniBand",
						Speed: "100Gbps",
					}.Build(),
				},
			}.Build()
			err := validator.Validate(spec)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("at least one network port with role 'fabric' is required"))
		})

		It("Still validates unique port names alongside fabric requirement", func() {
			spec := privatev1.BareMetalHardwareSpec_builder{
				Cpu:    validPrivateCPU(),
				Memory: validPrivateMemory(),
				NetworkPorts: []*privatev1.BareMetalNetworkPortSpec{
					privatev1.BareMetalNetworkPortSpec_builder{
						Name:  "data-0",
						Role:  "fabric",
						Type:  "Ethernet",
						Speed: "100Gbps",
					}.Build(),
					privatev1.BareMetalNetworkPortSpec_builder{
						Name:  "data-0",
						Role:  "fabric",
						Type:  "Ethernet",
						Speed: "100Gbps",
					}.Build(),
				},
			}.Build()
			err := validator.Validate(spec)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("network port names must be unique"))
		})
	})

	Describe("Public API", func() {
		It("Accepts hardware spec with a fabric port", func() {
			spec := publicv1.BareMetalHardwareSpec_builder{
				Cpu:    validPublicCPU(),
				Memory: validPublicMemory(),
				NetworkPorts: []*publicv1.BareMetalNetworkPortSpec{
					publicv1.BareMetalNetworkPortSpec_builder{
						Name:  "data-0",
						Role:  "fabric",
						Type:  "Ethernet",
						Speed: "100Gbps",
					}.Build(),
				},
			}.Build()
			err := validator.Validate(spec)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Rejects hardware spec with no network ports", func() {
			spec := publicv1.BareMetalHardwareSpec_builder{
				Cpu:    validPublicCPU(),
				Memory: validPublicMemory(),
			}.Build()
			err := validator.Validate(spec)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("at least one network port with role 'fabric' is required"))
		})

		It("Rejects hardware spec with only non-fabric ports", func() {
			spec := publicv1.BareMetalHardwareSpec_builder{
				Cpu:    validPublicCPU(),
				Memory: validPublicMemory(),
				NetworkPorts: []*publicv1.BareMetalNetworkPortSpec{
					publicv1.BareMetalNetworkPortSpec_builder{
						Name:  "mgmt-0",
						Role:  "management",
						Type:  "Ethernet",
						Speed: "1Gbps",
					}.Build(),
				},
			}.Build()
			err := validator.Validate(spec)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("at least one network port with role 'fabric' is required"))
		})
	})
})
