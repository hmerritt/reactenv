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
	// Files with occurrences (not every matched file will have an occurrence, so this may be less than `FilesMatchTotal`)
	Files []*fs.DirEntry

	// Total individual occurrences count
	Occurrences []Occurrence
	// Each file occurrence count + keys
	OccurrencesByFile []*FileOccurrences
	// Map of all unique environment variable keys
	OccurrenceKeys OccurrenceKeys
	// Map of all environment variable key values (keys will be replaced with these values)
	OccurrenceKeysReplacement OccurrenceKeysReplacement
}

type Occurrence = struct {
	FilePath string
	StartEnd []int
}
type OccurrenceKeys = map[string]bool
type OccurrenceKeysReplacement = map[string]string
type FileOccurrences = struct {
	Occurrences    []Occurrence
	OccurrenceKeys OccurrenceKeys
}

func NewReactenv(ui *ui.Ui) *Reactenv {
	return &Reactenv{
		UI:                        ui,
		Dir:                       "",
		Files:                     make([]*fs.DirEntry, 0),
		Occurrences:               make([]Occurrence, 0),
		OccurrencesByFile:         make([]*FileOccurrences, 0),
		OccurrenceKeys:            make(OccurrenceKeys),
		OccurrenceKeysReplacement: make(OccurrenceKeysReplacement),
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
	// Reset occurrence fields
	r.Occurrences = make([]Occurrence, 0)
	r.OccurrencesByFile = make([]*FileOccurrences, 0, len(r.Files))
	r.OccurrenceKeys = make(OccurrenceKeys)
	r.OccurrenceKeysReplacement = make(OccurrenceKeysReplacement)

	// Prep for removing files with no occurrences
	newFiles := make([]*fs.DirEntry, 0, len(r.Files))
	newOccurrencesTotalByFile := make([]*FileOccurrences, 0, len(r.Files))
	fileIndexesToRemove := make(map[int]int, 0)

	r.FilesWalkContents(func(fileIndex int, file fs.DirEntry, filePath string, fileContents []byte) error {
		fileOccurrences := regexp.MustCompile(REACTENV_FIND_EXPRESSION).FindAllIndex(fileContents, -1)

		fileOccurrencesToStore := make([]Occurrence, 0, len(fileOccurrences))
		for _, occurrence := range fileOccurrences {
			fileOccurrencesToStore = append(fileOccurrencesToStore, Occurrence{
				FilePath: filePath,
				StartEnd: occurrence,
			})
		}

		r.Occurrences = append(r.Occurrences, fileOccurrencesToStore...)
		r.OccurrencesByFile = append(r.OccurrencesByFile, &FileOccurrences{
			Occurrences:    fileOccurrencesToStore,
			OccurrenceKeys: make(OccurrenceKeys),
		})

		if len(fileOccurrences) == 0 {
			fileIndexesToRemove[fileIndex] = fileIndex
		}

		for _, occurrence := range fileOccurrences {
			occurrenceText := string(fileContents[occurrence[0]:occurrence[1]])
			envName := strings.Replace(occurrenceText, "__reactenv.", "", 1)
			envValue, envExists := os.LookupEnv(envName)

			if envExists {
				r.OccurrenceKeysReplacement[envName] = envValue
			}

			r.OccurrenceKeys[envName] = true
			r.OccurrencesByFile[fileIndex].OccurrenceKeys[envName] = true
		}

		return nil
	})

	// Remove files with no occurrences
	if len(fileIndexesToRemove) > 0 {
		for fileIndex, file := range r.Files {
			if _, ok := fileIndexesToRemove[fileIndex]; !ok {
				newFiles = append(newFiles, file)
				newOccurrencesTotalByFile = append(newOccurrencesTotalByFile, r.OccurrencesByFile[fileIndex])
			}
		}

		r.Files = newFiles
		r.OccurrencesByFile = newOccurrencesTotalByFile
	}
}

func (r *Reactenv) ReplaceOccurrences() {
	r.FilesWalkContents(func(fileIndex int, file fs.DirEntry, filePath string, fileContents []byte) error {
		fileContentsNew := make([]byte, 0, len(fileContents))
		fileOccurrences := r.OccurrencesByFile[fileIndex].Occurrences

		lastIndex := 0
		for _, occurrence := range fileOccurrences {
			start, end := occurrence.StartEnd[0], occurrence.StartEnd[1]
			occurrenceText := string(fileContents[start:end])
			envName := strings.Replace(occurrenceText, "__reactenv.", "", 1)
			envValue := r.OccurrenceKeysReplacement[envName]

			fileContentsNew = append(fileContentsNew, fileContents[lastIndex:start]...)
			fileContentsNew = append(fileContentsNew, envValue...)
			lastIndex = end
		}
		fileContentsNew = append(fileContentsNew, fileContents[lastIndex:]...)

		if err := os.WriteFile(filePath, fileContentsNew, 0644); err != nil {
			r.UI.Error(fmt.Sprintf("Error when writing to file '%s'.\n", filePath))
			r.UI.Error(fmt.Sprintf("%v", err))
			os.Exit(1)
		}

		return nil
	})
}
