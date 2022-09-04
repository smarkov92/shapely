package model

import "strings"

type DataForShap struct {
	Path  []string
	Value uint64
}

func (data *DataForShap) GenerateKey() string {
	return strings.Join(data.Path, ",")
}
