package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/docopt/docopt-go"
	"github.com/kovetskiy/lorg"
	logger "github.com/reconquest/structured-logger-go"
)

var (
	version = "[manual build]"
	usage   = "favor " + version + os.ExpandEnv(`

Usage:
  favor [options]
  favor -h | --help
  favor --version

Options:
  -c --config <file>  Read following configuration file.
                       [default: $HOME/.config/favor/favor.conf]
  --debug             Enable debug messages.
  -h --help           Show this screen.
  --version           Show version.
`)
)

var (
	log *logger.Logger
)

func initLogger(args map[string]interface{}) {
	stderr := lorg.NewLog()
	stderr.SetIndentLines(true)
	stderr.SetFormat(
		lorg.NewFormat("${time} ${level:[%s]:right:short} ${prefix}%s"),
	)

	debug := args["--debug"].(bool)

	if debug {
		stderr.SetLevel(lorg.LevelDebug)
	}

	log = logger.NewLogger(stderr)
}

func main() {
	args, err := docopt.Parse(usage, nil, true, version, false)
	if err != nil {
		panic(err)
	}

	initLogger(args)

	config, err := LoadConfig(args["--config"].(string))
	if err != nil {
		log.Fatalf(
			err,
			"unable to load configuration file: %s", args["--config"].(string),
		)
	}

	votes, err := LoadVotes(config.VotesPath)
	if err != nil {
		log.Fatalf(
			err,
			"unable to load votes file: %s", config.VotesPath,
		)
	}

	var (
		trees = prepareTrees(config.Trees, votes)

		scanner = &Scanner{
			ignoreMap: makeMap(config.IgnoreGlobal),
		}

		scheduler = &Scheduler{
			max: config.Threads,
		}
	)

	scanner.scheduler = scheduler
	scheduler.scanner = scanner

	for _, tree := range trees {
		scheduler.Schedule(tree, ".")
	}

	scheduler.Wait()

	items := sortScanItems(scanner.items)

	choose, err := pick(config.Picker, trees, items)
	if err != nil {
		log.Fatal(err)
	}

	if choose == nil {
		return
	}

	choose.tree.votes[choose.dir] += 1

	err = SaveVotes(config.VotesPath, votes)
	if err != nil {
		log.Fatalf(
			err,
			"unable to save votes file: %s", config.VotesPath,
		)
	}

	fmt.Println(filepath.Join(choose.tree.Dir, choose.dir))
}

func sortScanItems(items []*ScanItem) []*ScanItem {
	sort.SliceStable(items, func(i, j int) bool {
		x := items[i]
		y := items[j]

		// NOTE: > used for votes instead of < function
		if x.votes > y.votes {
			return true
		}

		if x.votes < y.votes {
			return false
		}

		if x.tree.Name < y.tree.Name {
			return true
		}

		if x.tree.Name > y.tree.Name {
			return false
		}

		if x.dir < y.dir {
			return true
		}

		if x.dir > y.dir {
			return false
		}

		return false
	})

	return items
}

func expandHomeTilda(target string) string {
	if strings.HasPrefix(target, "~/") {
		target = os.Getenv("HOME") + target[1:]
	}

	return target
}
