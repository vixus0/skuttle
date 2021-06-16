package file_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vixus0/skuttle/v2/internal/provider/file"
)

var _ = Describe("File provider", func() {
	var (
		provider *file.Provider
	)

	BeforeEach(func() {
		provider = &file.Provider{
			Nodes: []string{
				"node1",
				"node2",
			},
		}
	})

	Describe("Checking an instance exists", func() {
		It("should be true only for instances in the node list", func() {
			Expect(provider.InstanceExists("file://node1")).To(BeTrue())
			Expect(provider.InstanceExists("file://node2")).To(BeTrue())
			Expect(provider.InstanceExists("file://node3")).To(BeFalse())
		})
	})
})
