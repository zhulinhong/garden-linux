package rootfs_provider

import "path/filepath"

//go:generate counterfeiter -o fake_namespacer/fake_namespacer.go . Namespacer
type Namespacer interface {
	Namespace(path string) (string, error)
}

type namespacer struct {
	Translator filepath.WalkFunc
}

func NewNamespacer(t filepath.WalkFunc) Namespacer {
	return &namespacer{t}
}

func (n *namespacer) Namespace(rootfs string) (mountpoint string, err error) {
	filepath.Walk(rootfs, n.Translator)
	return rootfs, nil
}
