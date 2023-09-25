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
		return scanErr == nil
	}, nil)
	if scanErr != nil {
		return scanErr
	}

	// Copy the entire GOROOT/src directory.
	if err := copyDir(filepath.Join(newRoot, "src"), filepath.Join(goroot, "src")); err != nil {
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

type copyOperation struct{ src, dst string }

func copyDir(dst, src string) error {
	ops := make(chan copyOperation, 256)

	var group errgroup.Group
	group.Go(func() error {
		err := scanDir(dst, src, ops)
		close(ops)
		return err
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
	return group.Wait()
}

func scanDir(dst, src string, ops chan<- copyOperation) error {
	if err := os.MkdirAll(dst, 0755); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			dstChild := filepath.Join(dst, name)
			srcChild := filepath.Join(src, name)
			if err := scanDir(dstChild, srcChild, ops); err != nil {
				return err
			}
		} else {
			ops <- copyOperation{
				src: filepath.Join(src, name),
				dst: filepath.Join(dst, name),
			}
		}
	}
	return nil
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
