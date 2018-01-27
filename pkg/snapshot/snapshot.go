package snapshot

// import (
// 	"archive/tar"
// 	"compress/gzip"
// 	"fmt"
// 	"io"
// 	"io/ioutil"
// 	"os"
// 	"path/filepath"
// 	"strings"
// )

// // TODO: should be /
// var directory = "/Users/priyawadhwa/go/src/github.com/priyawadhwa/kbuild/testexec"

// type Snapshotter struct {
// 	l         *LayeredMap
// 	directory string
// 	snapshots []string
// }

// func NewSnapshotter(l *LayeredMap, d string) *Snapshotter {
// 	return &Snapshotter{l: l, directory: d, snapshots: []string{}}
// }

// func (s *Snapshotter) Init() error {
// 	if err := s.snapShotFS(ioutil.Discard); err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (s *Snapshotter) TakeSnapshot() error {
// 	fmt.Println("taking snapshots in ", s.directory)
// 	path := filepath.Join(s.directory+"/work-dir", fmt.Sprintf("layer-%d.tar.gz", len(s.snapshots)))
// 	fmt.Println("Generating a snapshot in: ", path)
// 	f, err := os.Create(path)
// 	defer f.Close()
// 	if err != nil {
// 		return err
// 	}

// 	gz := gzip.NewWriter(f)
// 	defer gz.Close()
// 	if err := s.snapShotFS(gz); err != nil {
// 		return err
// 	}

// 	s.snapshots = append(s.snapshots, path)
// 	return nil
// }

// func (s *Snapshotter) snapShotFS(f io.Writer) error {
// 	s.l.Snapshot()

// 	w := tar.NewWriter(f)
// 	defer w.Close()

// 	return filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
// 		if ignorePath(path) {
// 			return nil
// 		}

// 		// Only add to the tar if we add it to the layeredmap.
// 		if s.l.MaybeAdd(path) {
// 			return addToTar(path, info, w)
// 		}
// 		return nil
// 	})
// }

// //TODO: remove this
// func ignorePath(p string) bool {
// 	for _, d := range []string{directory + "/dev", directory + "/sys", directory + "/proc", directory + "/work-dir", directory + "/dockerfile"} {
// 		if strings.HasPrefix(p, d) {
// 			return true
// 		}
// 	}
// 	return false
// }

// func addToTar(p string, i os.FileInfo, w *tar.Writer) error {
// 	linkDst := ""
// 	if i.Mode()&os.ModeSymlink != 0 {
// 		var err error
// 		linkDst, err = os.Readlink(p)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	hdr, err := tar.FileInfoHeader(i, linkDst)
// 	if err != nil {
// 		return err
// 	}
// 	hdr.Name = p
// 	w.WriteHeader(hdr)
// 	if !i.Mode().IsRegular() {
// 		return nil
// 	}
// 	r, err := os.Open(p)
// 	if err != nil {
// 		return err
// 	}
// 	if _, err := io.Copy(w, r); err != nil {
// 		return err
// 	}
// 	return nil
// }
