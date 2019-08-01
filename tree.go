package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-yaml/yaml"
	"github.com/kovetskiy/ko"
	"github.com/reconquest/karma-go"
)

type Trees []*Tree

type Tree struct {
	Name        string   `yaml:"name" required:"true"`
	Dir         string   `yaml:"dir" required:"true"`
	MinDepth    int      `yaml:"min_depth"`
	MaxDepth    int      `yaml:"max_depth"`
	Ignore      []string `yaml:"ignore"`
	IncludeRoot bool     `yaml:"include_root"`

	ignoreMap map[string]struct{}
	votes     map[string]int
}

type Votes map[string]map[string]int

func LoadVotes(path string) (Votes, error) {
	var votes Votes
	err := ko.Load(expandHomeTilda(path), &votes, yaml.Unmarshal)
	if err != nil {
		if os.IsNotExist(err) {
			return make(Votes), nil
		}
		return nil, err
	}

	return votes, nil
}

func SaveVotes(path string, votes Votes) error {
	path = expandHomeTilda(path)

	contents, err := yaml.Marshal(votes)
	if err != nil {
		return karma.Format(
			err,
			"unable to encode to yaml",
		)
	}

	err = os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return karma.Format(
			err,
			"unable to mkdir",
		)
	}

	err = ioutil.WriteFile(path, contents, 0644)
	if err != nil {
		return err
	}

	return nil
}

func prepareTrees(trees Trees, votes Votes) Trees {
	for i, tree := range trees {
		tree.ignoreMap = makeMap(tree.Ignore)

		if _, ok := votes[tree.Name]; !ok {
			votes[tree.Name] = map[string]int{}
		}

		tree.votes = votes[tree.Name]

		tree.Dir = expandHomeTilda(tree.Dir)

		trees[i] = tree
	}

	return trees
}
