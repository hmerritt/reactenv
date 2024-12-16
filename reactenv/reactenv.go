package reactenv

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/hmerritt/reactenv/ui"
)

const (
	REACTENV_PREFIX          = "__reactenv"
	REACTENV_FIND_EXPRESSION = `(__reactenv\.[a-zA-Z_$][0-9a-zA-Z_$]*)`
)

type Reactenv struct {
	UI *ui.Ui

	// Path of directory to scan
	Dir string

	// Total file count (that match `REACTENV_FIND_EXPRESSION`, within `Dir`)
	FilesMatchTotal int
	// Files with occurrences (that match `REACTENV_FIND_EXPRESSION`, within `Dir`)
	Files []*fs.DirEntry

	// Total individual occurrences count
	OccurrencesTotal int
	// Each file occurrence count + keys
	OccurrencesTotalByFile map[int]*FileOccurrences
	// Map of all unique environment variable keys
	OccurrenceKeys OccurrenceKeys
	// Map of all environment variable key values (keys will be replaced with these values)
	OccurrenceKeysReplacement OccurrenceKeysReplacement
}

type OccurrenceKeys = map[string]bool
type OccurrenceKeysReplacement = map[string]string
type FileOccurrences = struct {
	Total          int
	OccurrenceKeys OccurrenceKeys
}

func NewReactenv(ui *ui.Ui) *Reactenv {
	return &Reactenv{
		UI:                     ui,
		Dir:                    "",
		Files:                  make([]*fs.DirEntry, 0),
		OccurrenceKeys:         make(OccurrenceKeys),
		OccurrencesTotal:       0,
		OccurrencesTotalByFile: make(map[int]*FileOccurrences),
	}
}

// Populates `Reactenv.Files` with all files that match `fileMatchExpression`
func (r *Reactenv) FindFiles(dir string, fileMatchExpression string) error {
	r.Dir = dir
	files, err := os.ReadDir(r.Dir)

	if err != nil {
		return err
	}

	fileMatcher, err := regexp.Compile(fileMatchExpression)

	if err != nil {
		return err
	}

	for _, file := range files {
		if fileMatcher.MatchString(file.Name()) && !file.IsDir() {
			r.Files = append(r.Files, &file)
		}
	}

	r.FilesMatchTotal = len(r.Files)

	return nil
}

// Run a callback for each File
func (r *Reactenv) FilesWalk(fileCb func(fileIndex int, file fs.DirEntry, filePath string) error) error {
	for fileIndex, file := range r.Files {
		filePath := path.Join(r.Dir, (*file).Name())
		err := fileCb(fileIndex, *file, filePath)
		if err != nil {
			return err
		}
	}

	return nil
}

// Run a callback for each File, passing in the file contents
func (r *Reactenv) FilesWalkContents(fileCb func(fileIndex int, file fs.DirEntry, filePath string, fileContents []byte) error) error {
	for fileIndex, file := range r.Files {
		filePath := path.Join(r.Dir, (*file).Name())
		fileContents, err := os.ReadFile(filePath)

		if err != nil {
			r.UI.Error(fmt.Sprintf("Error when reading file '%s'.\n", (*file).Name()))
			r.UI.Error(fmt.Sprintf("%v", err))
			os.Exit(1)
		}

		err = fileCb(fileIndex, *file, filePath, fileContents)

		if err != nil {
			return err
		}
	}

	return nil
}

// Walks every file and populates `Reactenv.Occurrences*` fields.
func (r *Reactenv) FindOccurrences() {
	r.OccurrenceKeys = make(OccurrenceKeys)
	r.OccurrencesTotal = 0
	r.OccurrencesTotalByFile = make(map[int]*FileOccurrences)

	fileIndexesToRemove := make([]int, 0)

	r.FilesWalkContents(func(fileIndex int, file fs.DirEntry, filePath string, fileContents []byte) error {
		fileOccurrences := regexp.MustCompile(REACTENV_FIND_EXPRESSION).FindAllIndex(fileContents, -1)

		r.OccurrencesTotal += len(fileOccurrences)
		r.OccurrencesTotalByFile[fileIndex] = &FileOccurrences{
			Total:          len(fileOccurrences),
			OccurrenceKeys: make(OccurrenceKeys),
		}

		if len(fileOccurrences) == 0 {
			fileIndexesToRemove = append(fileIndexesToRemove, fileIndex)
		}

		for _, occurrence := range fileOccurrences {
			occurrenceText := string(fileContents[occurrence[0]:occurrence[1]])
			envName := strings.Replace(occurrenceText, "__reactenv.", "", 1)
			envValue, envExists := os.LookupEnv(envName)

			if envExists {
				r.OccurrenceKeysReplacement[envName] = envValue
			}

			r.OccurrenceKeys[envName] = true
			r.OccurrencesTotalByFile[fileIndex].OccurrenceKeys[envName] = true
		}

		return nil
	})

	// Remove files with no occurrences
	for _, fileIndex := range fileIndexesToRemove {
		r.Files = append(r.Files[:fileIndex], r.Files[fileIndex+1:]...)
	}
}

// func (r *Reactenv) ReplaceOccurrences() {
// 	r.FilesWalkContents(func(fileIndex int, file fs.DirEntry, filePath string, fileContents []byte) error {
// 		fileOccurrences := regexp.MustCompile(REACTENV_FIND_EXPRESSION).FindAllIndex(fileContents, -1)
// 		occurrenceReplacementValues := make([]string, len(occurrences))
// 	}
// }
