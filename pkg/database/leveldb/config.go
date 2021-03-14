package leveldb

const Type = "leveldb"

// Config defines database configuration.
type Config struct {
	DataDir   string `json:"data_dir" split_words:"true"`
	KeyPrefix string `json:"key_prefix" split_words:"true"`
}
