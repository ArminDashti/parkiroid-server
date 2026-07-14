package store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	ModelParamFileName = "model.param"
	ModelBinFileName   = "model.bin"
)

func ModelDir(modelsRoot, modelID string) string {
	return filepath.Join(modelsRoot, modelID)
}

func ModelParamPath(modelsRoot, modelID string) string {
	return filepath.Join(ModelDir(modelsRoot, modelID), ModelParamFileName)
}

func ModelBinPath(modelsRoot, modelID string) string {
	return filepath.Join(ModelDir(modelsRoot, modelID), ModelBinFileName)
}

func HasModelFiles(modelsRoot, modelID string) bool {
	paramPath := ModelParamPath(modelsRoot, modelID)
	binPath := ModelBinPath(modelsRoot, modelID)
	paramInfo, paramErr := os.Stat(paramPath)
	binInfo, binErr := os.Stat(binPath)
	return paramErr == nil && binErr == nil && !paramInfo.IsDir() && !binInfo.IsDir()
}

func ComputeFileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open file for sha256: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("hash file: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
