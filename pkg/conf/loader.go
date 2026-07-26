package conf

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

func LoadTOML(path string, dst any) error {
	meta, err := LoadTOMLWithMeta(path, dst)
	if err != nil {
		return err
	}
	if err := rejectRemovedKeys(meta); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if err := rejectUnknownKeys(meta); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	return nil
}

func LoadTOMLWithMeta(path string, dst any) (toml.MetaData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return toml.MetaData{}, fmt.Errorf("%w: read %s: %v", ErrRead, path, err)
	}
	meta, err := toml.Decode(string(data), dst)
	if err != nil {
		return toml.MetaData{}, fmt.Errorf("%w: parse %s: %v", ErrParse, path, err)
	}
	return meta, nil
}

func rejectUnknownKeys(meta toml.MetaData) error {
	keys := meta.Undecoded()
	if len(keys) == 0 {
		return nil
	}
	names := make([]string, len(keys))
	for i, key := range keys {
		names[i] = key.String()
	}
	return fmt.Errorf("unknown TOML key(s): %s", strings.Join(names, ", "))
}
