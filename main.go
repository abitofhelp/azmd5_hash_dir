package main

import (
	"context"
	mapset "github.com/deckarep/golang-set/v2"
	"golang.org/x/sync/errgroup"
	"io/fs"
	"os"
	"runtime"
	"sort"
	"sync/atomic"
	"time"
)

// This application demonstrates the correct way to calculate an MD5 hash
// that is compatible with Azure Blob Storage's Content-MD5.

import (
	"fmt"
	"github.com/abitofhelp/azmd5_hash_dir/hash/md5"
	"github.com/abitofhelp/azmd5_hash_dir/hash/model"
	"path/filepath"
)

const kTimeout = 30 * time.Second

func WalkDirectoryWithChannel(
	ctx context.Context,
	root string,
	excludeDirs mapset.Set[string],
	excludeFiles mapset.Set[string],
	paths chan<- string) error {
	cnt := 0
	if err := fs.WalkDir(os.DirFS(root), ".", func(p string, d fs.DirEntry, err error) error {

		filePath := root + "/" + p

		if err != nil {
			return fmt.Errorf("failed to walk the directory '%s': %w", filePath, err)
		}

		if excludeDirs.Contains(d.Name()) {
			return filepath.SkipDir
		}

		// Scan any directories that are not in excludeDir.
		if d.IsDir() {
			return nil
		}

		// Skip any files that are in excludeFiles.
		if !d.IsDir() && excludeFiles.Contains(d.Name()) {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case paths <- filepath.Join(p):
			cnt++
			fmt.Printf("prodcnt: %d\n", cnt)
		}

		return nil
	}); err == nil {
		return nil
	} else {
		return fmt.Errorf("failed to walk the directory '%s': %w", root, err)
	}
}

func main() {
	root := "/Users/mike/Downloads/clients/alm"
	excludeDirs := mapset.NewSet[string]()
	excludeDirs.Add("assets")
	excludeDirs.Add("aerials")
	excludeDirs.Add("projects")

	excludeFiles := mapset.NewSet[string]()
	excludeFiles.Add(".DS_Store")

	// find all files
	ctx, cancel := context.WithTimeout(context.Background(), kTimeout)
	eg, eqctx := errgroup.WithContext(ctx)
	paths := make(chan string, runtime.NumCPU())

	start := time.Now()
	// Producer: Get the paths to the files of interest within the root directory.
	eg.Go(func() error {
		defer close(paths)
		if err := WalkDirectoryWithChannel(eqctx, root, excludeDirs, excludeFiles, paths); err == nil {
			return nil
		} else {
			return fmt.Errorf("failed to walk through directory '%s': %w", root, err)
		}
	})

	// Consumer: Hash the files
	localFiles := make(chan *model.LocalFile, 10)
	workers := int64(runtime.NumCPU())
	cnt := -0
	for i := int64(0); i < workers; i++ {
		eg.Go(func() error {

			defer func() {
				// Close the channel when the last worker completes.
				if atomic.AddInt64(&workers, -1) == 0 {
					close(localFiles)
				}
			}()

			for p := range paths {
				// Calculate the base64 MD5 hash of the file.
				fullPath := filepath.Join(root, p)
				if azureMd5, err := md5.GenMd5HashAsBase64(fullPath); err == nil {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case localFiles <- model.NewLocalFile(p, azureMd5):
						cnt++
						fmt.Printf("hashcnt: %d\n", cnt)
					}
				} else {
					return fmt.Errorf("failed to generate a base64 hash of file '%s': %w", fullPath, err)
				}
			}
			return nil
		})
	}

	// Reduce & Sort: Slice of hashes ordered by path.
	var hashes []*model.LocalFile
	eg.Go(func() error {
		for lf := range localFiles {
			hashes = append(hashes, lf)
		}

		sort.Slice(hashes, func(i, j int) bool {
			return hashes[i].PathInsideDirectory() < hashes[j].PathInsideDirectory()
		})

		return nil
	})

	if err := eg.Wait(); err == nil {
		elapsed := time.Since(start)
		fmt.Println()
		for i, h := range hashes {
			fmt.Printf("(%4d) %s => %s\n", i+1, h.PathInsideDirectory(), h.AzureMd5())
		}

		fmt.Printf("Elapsed: %d ms\n", elapsed.Milliseconds())
	} else {
		panic(fmt.Errorf("failed to generate hashes for directory '%s': %w", root, err))
	}

	fmt.Println()
	cancel()
}
