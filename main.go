package main

import (
	mapset "github.com/deckarep/golang-set/v2"
	"io/fs"
	"os"
	"runtime"
	"sync"
)

// This application demonstrates the correct way to calculate an MD5 hash
// that is compatible with Azure Blob Storage's Content-MD5.

import (
	"fmt"
	"github.com/abitofhelp/azmd5_hash_dir/hash/md5"
	"github.com/abitofhelp/azmd5_hash_dir/hash/model"
	"path/filepath"
)

func WalkDirectory(
	root string,
	excludeDirs mapset.Set[string],
	excludeFiles mapset.Set[string]) (mapset.Set[*model.LocalFile], error) {

	localPaths := mapset.NewSet[*model.LocalFile]()

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

		// Calculate the base64 hash for a file of interest.
		if azureMd5, err := md5.GenMd5HashAsBase64(filePath); err == nil {
			lf := model.NewLocalFile(p, azureMd5)
			localPaths.Add(lf)
		} else {
			return fmt.Errorf("failed to generate a hash for '%s': %w", filePath, err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to walk the directory '%s': %w", root, err)
	}

	return localPaths, nil
}

func WalkDirectoryWithChannel(
	root string,
	excludeDirs mapset.Set[string],
	excludeFiles mapset.Set[string],
	c chan<- string) {

	defer close(c)

	fs.WalkDir(os.DirFS(root), ".", func(p string, d fs.DirEntry, err error) error {

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

		// Calculate the base64 hash for a file of interest.
		c <- filepath.Join(p)
		return nil
	})
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
	c := make(chan string, runtime.NumCPU())
	go WalkDirectoryWithChannel(root, excludeDirs, excludeFiles, c)

	// hash all the files
	wg := sync.WaitGroup{}
	rc := make(chan model.LocalFile, runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range c {

				// Calculate the base64 hash for a file of interest.
				if azureMd5, err := md5.GenMd5HashAsBase64(filepath.Join(root, filePath) { //; err == nil {
					lf := model.NewLocalFile(filePath, azureMd5)
					rc <- *lf
				} //else {
					//return fmt.Errorf("failed to generate a hash for '%s': %w", filePath, err)
				//}
			}
		}()
	}

	// collect all the file hashes
	m := make(map[string][]string)
	go func() {
		for r := range rc {
			m[r.PathInsideDirectory()] = append(m[r.PathInsideDirectory()], r.AzureMd5())
		}
	}()
	wg.Wait()
	close(rc)

	//paths := localPaths.ToSlice()
	//sort.Slice(paths, func(i, j int) bool {
	//	return paths[i].PathInsideDirectory() < paths[j].PathInsideDirectory()
	//})

	cnt := 0
	for k, v := range m {
		cnt++
		fmt.Printf("(%00d)%s\n%s\n\n", cnt, k, v)
	}

	//
	//for _, p := range paths {
	//
	//	fmt.Printf("(%00d)%s\n%s\n\n", cnt, p.PathInsideDirectory(), p.AzureMd5())
	//}
}
