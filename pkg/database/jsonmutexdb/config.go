package jsonmutexdb

const Type = "jsonmutexdb"

// Config defines database configuration.
type Config struct {
	DataDir        string `json:"data_dir" split_words:"true"`
	FilenamePrefix string `json:"filename_prefix" split_words:"true"`
}
