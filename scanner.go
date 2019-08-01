package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type ScanItem struct {
	tree  *Tree
	dir   string
	votes int
}

type Scanner struct {
	scheduler *Scheduler
	ignoreMap map[string]struct{}
	mutex     sync.Mutex
	items     []*ScanItem
}

func makeMap(slice []string) map[string]struct{} {
	hash := map[string]struct{}{}
	for _, item := range slice {
		hash[item] = struct{}{}
	}

	return hash
}

func (scanner *Scanner) append(tree *Tree, dir string) {
	scanner.mutex.Lock()
	scanner.items = append(scanner.items, &ScanItem{
		tree:  tree,
		dir:   dir,
		votes: tree.votes[dir],
	})
	scanner.mutex.Unlock()
}

func (scanner *Scanner) Scan(tree *Tree, dir string) {
	depth := strings.Count(dir, "/") + 1

	if depth >= tree.MinDepth {
		if dir != "." || tree.IncludeRoot {
			scanner.append(tree, dir)
		}
	}

	if tree.MaxDepth  == 0 && tree.IncludeRoot {
		return
	}

	if depth < tree.MaxDepth || dir == "." {
		path := filepath.Join(tree.Dir, dir)

		names, err := readdir(path)
		if err != nil {
			log.Errorf(err, "unable to read: %s", path)
		}

		for _, name := range names {
			_, skip := scanner.ignoreMap[name]
			if skip {
				continue
			}

			_, skip = tree.ignoreMap[name]
			if skip {
				continue
			}

			scanner.scheduler.Schedule(tree, filepath.Join(dir, name))
		}
	}
}

func readdir(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	infos, err := file.Readdir(-1)
	if err != nil {
		return nil, err
	}

	names := []string{}
	for _, info := range infos {
		if info.IsDir() {
			names = append(names, info.Name())
			continue
		}

		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			realpath, err := filepath.EvalSymlinks(
				filepath.Join(path, info.Name()),
			)
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return nil, err
			}

			realstat, err := os.Stat(realpath)
			if err != nil {
				return nil, err
			}

			if realstat.IsDir() {
				names = append(names, info.Name())
			}
		}
	}

	return names, nil
}
