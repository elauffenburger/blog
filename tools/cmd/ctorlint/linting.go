package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"golang.org/x/sync/errgroup"
)

func lintPkg(pkg string, files []string) error {
	for _, f := range files {
		f := f

		fileBytes, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}

		var (
			wg        errgroup.Group
			pkgStrcts []strct
			pkgCtors  []ctor
		)

		file := string(fileBytes)

		wg.Go(func() error {
			pkgStrcts, err = findStructs(strings.NewReader(file))
			return err
		})

		wg.Go(func() error {
			pkgCtors, err = findCtors(file)
			return err
		})

		wg.Wait()

		fmt.Printf("pkg: %s\n", pkg)
		fmt.Printf("pkgStrcts: %#v\n", pkgStrcts)
		fmt.Printf("pkgCtors: %#v\n", pkgCtors)

		var unmtchdStrcts []strct

	strctLoop:
		for _, s := range pkgStrcts {
			if s.vis == unexported || s.nolint {
				continue
			}

			for _, c := range pkgCtors {
				if c.MatchesStruct(s) {
					continue strctLoop
				}
			}

			unmtchdStrcts = append(unmtchdStrcts, s)
		}

		fmt.Printf("unmatched structs: %#v\n\n", unmtchdStrcts)
	}

	return nil
}
