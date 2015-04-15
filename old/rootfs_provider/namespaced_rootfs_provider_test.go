package rootfs_provider_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry-incubator/garden-linux/old/rootfs_provider"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Namespacer", func() {
	var rootfs string

	BeforeEach(func() {
		var err error
		rootfs, err = ioutil.TempDir("", "rootfs")
		Expect(err).NotTo(HaveOccurred())

		os.MkdirAll(filepath.Join(rootfs, "foo", "bar", "baz"), 0755)
		ioutil.WriteFile(filepath.Join(rootfs, "foo", "beans"), []byte("jam"), 0721)
	})

	It("translates all of the uids in the rootfs", func() {
		var translated []translation
		namespacer := rootfs_provider.NewNamespacer(
			func(path string, info os.FileInfo, err error) error {
				translated = append(translated, translation{
					path: path,
					info: info,
					err:  err,
				})

				return nil
			},
		)

		_, err := namespacer.Namespace(rootfs)
		Expect(err).NotTo(HaveOccurred())

		Expect(translated).NotTo(BeEmpty())

		info, err := os.Stat(filepath.Join(rootfs, "foo", "bar", "baz"))
		Expect(err).NotTo(HaveOccurred())
		Expect(translated).To(ContainElement(translation{
			path: filepath.Join(rootfs, "foo", "bar", "baz"),
			info: info,
			err:  nil,
		}))

		info, err = os.Stat(filepath.Join(rootfs, "foo", "beans"))
		Expect(err).NotTo(HaveOccurred())
		Expect(translated).To(ContainElement(translation{
			path: filepath.Join(rootfs, "foo", "beans"),
			info: info,
			err:  nil,
		}))
	})
})

type translation struct {
	path string
	info os.FileInfo
	err  error
}
