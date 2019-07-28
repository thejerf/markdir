package main

import (
	"os"
	"testing"
)

func TestGetCodeRoot(t *testing.T) {
	path, err := GetCodeRoot()

	if err != nil {
		t.Errorf("err %v != nil", err)
	}

	pwd, _ := os.Getwd()
	if path != pwd {
		t.Errorf("path %v != %v", path, pwd)
	}
}

func TestGetCodeRootWithVariable(t *testing.T) {
	targetPath := "/tmp"

	err := os.Setenv("MARKDIR_CODE_ROOT", targetPath)
	if err != nil {
		t.Error(err)
	}

	path, err := GetCodeRoot()
	if err != nil {
		t.Errorf("err %v != nil", err)
	}

	if path != targetPath {
		t.Errorf("path %v != %v", path, targetPath)
	}
}

func TestExpand(t *testing.T) {
	targetPath, _ := os.Getwd()

	path := Expand(".")

	if path != targetPath {
		t.Errorf("path %v != %v", path, targetPath)
	}
}
