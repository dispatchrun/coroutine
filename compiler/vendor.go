package compiler

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/go/packages"
)

const copyConcurrency = 16

// vendor copies GOROOT packages into a new directory.
func vendor(newRoot string, pkgs []*packages.Package) error {
	goroot := runtime.GOROOT()

	var dirs []string
	var scanErr error
	packages.Visit(pkgs, func(p *packages.Package) bool {
		path := filepath.Dir(p.GoFiles[0])
		rel, err := filepath.Rel(goroot, path)
		if err != nil {
			scanErr = err
			return false
		}
		if strings.HasPrefix(rel, "..") {
			// TODO: these could be dependencies in GOPATH, in which
			//  case they could be vendored under ./vendor?
			scanErr = fmt.Errorf("package %s (%s) is not in GOROOT (%s)", p.PkgPath, path, goroot)
			return false
		}
		dirs = append(dirs, rel)
		return scanErr == nil
	}, nil)
	if scanErr != nil {
		return scanErr
	}

	type copyOperation struct{ src, dst string }
	ops := make(chan copyOperation, 256)
	var group errgroup.Group
	group.Go(func() error {
		for _, rel := range dirs {
			srcDir := filepath.Join(goroot, rel)
			dstDir := filepath.Join(newRoot, rel)
			if err := os.MkdirAll(dstDir, 0755); err != nil {
				return err
			}
			entries, err := os.ReadDir(srcDir)
			if err != nil {
				return err
			}
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				name := entry.Name()
				ops <- copyOperation{
					src: filepath.Join(srcDir, name),
					dst: filepath.Join(dstDir, name),
				}
			}
		}
		close(ops)
		return nil
	})
	for i := 0; i < copyConcurrency; i++ {
		group.Go(func() error {
			for op := range ops {
				if err := copyFile(op.dst, op.src); err != nil {
					return err
				}
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return err
	}

	// Rewrite GoFiles paths.
	packages.Visit(pkgs, func(p *packages.Package) bool {
		for i, path := range p.GoFiles {
			rel, _ := filepath.Rel(goroot, path)
			p.GoFiles[i] = filepath.Join(newRoot, rel)
		}
		return true
	}, nil)

	// Symlink $GOROOT/pkg, which contains directories required
	// at compile time.
	err := os.Symlink(filepath.Join(goroot, "pkg"), filepath.Join(newRoot, "pkg"))
	if errors.Is(err, os.ErrExist) {
		err = nil
	}
	return err
}

func copyFile(dst, src string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return dstFile.Close()
}
