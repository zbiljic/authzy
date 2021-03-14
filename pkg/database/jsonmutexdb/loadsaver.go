package jsonmutexdb

import (
	"fmt"
	"sync"

	"github.com/spf13/afero"
)

type LoadSaver interface {
	Load(filename string) ([]byte, error)
	Save(filename string, data []byte) error
}

func NewLoadSaver(dataDir string) (LoadSaver, error) {
	var fs afero.Fs

	if dataDir == "" {
		fs = afero.NewMemMapFs()
	} else {
		afs := afero.NewOsFs()

		dirExists, err := afero.DirExists(afs, dataDir)
		if err != nil {
			return nil, err
		}

		if !dirExists {
			return nil, fmt.Errorf("dir does not exists: %s", dataDir)
		}

		fs = afero.NewBasePathFs(afs, dataDir)
	}

	return &loadSave{fs: fs}, nil
}

type loadSave struct {
	fs afero.Fs
	mu sync.Mutex
}

func (ls *loadSave) Load(filename string) ([]byte, error) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	exists, err := afero.Exists(ls.fs, filename)
	if err != nil {
		return nil, err
	}

	if exists {
		return afero.ReadFile(ls.fs, filename)
	}

	return []byte{}, nil
}

func (ls *loadSave) Save(filename string, data []byte) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	return afero.WriteFile(ls.fs, filename, data, 0644)
}
