package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
)

var (
	additionalDirs = []string{"~/pdev/taylormonacelli/northflier"} // specify additional directories here
	logFlag        bool
	gitPullFlag    bool
	files          StringArray
)

type StringArray []string

func (i *StringArray) String() string {
	return strings.Join(*i, ",")
}

func (i *StringArray) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func expandHomeDir(dirs []string) ([]string, error) {
	expandedDirs := make([]string, len(dirs))
	for i, dir := range dirs {
		dir, err := homedir.Expand(dir)
		if err != nil {
			return nil, err
		}
		expandedDirs[i] = dir
	}
	return expandedDirs, nil
}

func getCandidateDirs() ([]string, error) {
	var dirs []string
	for i := 1; i <= 10; i++ {
		dir := fmt.Sprintf("~/pdev/tmp/northflier%d", i)
		dirs = append(dirs, dir)
	}
	expandedDirs, err := expandHomeDir(additionalDirs)
	if err != nil {
		return nil, err
	}
	dirs = append(dirs, expandedDirs...)
	allExpandedDirs, err := expandHomeDir(dirs)
	if err != nil {
		return nil, err
	}
	return allExpandedDirs, nil
}

func checkForDuplicates(strSlice []string) (bool, string) {
	allKeys := make(map[string]bool)
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
		} else {
			return true, item
		}
	}
	return false, ""
}

func main() {
	flag.BoolVar(&logFlag, "log", false, "Enable logging")
	flag.BoolVar(&gitPullFlag, "git-pull", false, "Enable git pull")
	flag.Var(&files, "file", "File to search for (this flag can be set multiple times)")
	flag.Parse()

	if len(files) == 0 {
		files = append(files, "summary.txt")
	}

	// Warn about duplicate files
	if hasDuplicate, dupFile := checkForDuplicates(files); hasDuplicate {
		fmt.Printf("Warning: The file '%s' has been specified more than once.\n", dupFile)
	}

	dirs, err := getCandidateDirs()
	if err != nil {
		log.Fatalf("Error getting candidate directories: %v", err)
	}

	for _, dir := range dirs {
		if logFlag {
			log.Printf("Checking directory: %s", dir)
		}

		_, err := os.Stat(fmt.Sprintf("%s/data", dir))
		if os.IsNotExist(err) {
			if logFlag {
				log.Printf("Data directory does not exist: %s/data\n", dir)
			}
			continue
		}

		if logFlag {
			log.Printf("Data directory exists: %s/data\n", dir)
		}

		for _, file := range files {
			found := false

			err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if info.IsDir() {
					return nil
				}

				matched, err := filepath.Match(file, filepath.Base(path))
				if err != nil {
					return err
				}

				if matched {
					realPath, _ := filepath.Abs(path)
					if logFlag {
						log.Printf("Found %s file: %s\n", file, realPath)
					}
					fmt.Println(realPath)
					found = true
				}

				return nil
			})

			if err != nil {
				log.Printf("Error walking the path %v: %v\n", dir, err)
				continue
			}

			if !found {
				continue
			}

			_, err = os.Stat(dir)
			if os.IsNotExist(err) {
				continue
			}

			if gitPullFlag {
				if logFlag {
					log.Printf("Changing working directory to: %s\n", dir)
					log.Println("Executing git pull command...")
				}

				cmd := exec.Command("git", "-C", dir, "pull")
				cmd.Run()
			}
		}
	}
}
