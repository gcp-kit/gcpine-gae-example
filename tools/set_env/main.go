package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Env ...
type Env struct {
	GcpProject string `json:"GCP_PROJECT"`
	ShortSha   string `json:"SHORT_SHA"`
	BranchName string `json:"BRANCH_NAME"`
	BuildID    string `json:"BUILD_ID"`
}

func main() {
	env := &Env{
		GcpProject: os.Getenv("GCP_PROJECT"),
		ShortSha:   os.Getenv("SHORT_SHA"),
		BranchName: os.Getenv("BRANCH_NAME"),
		BuildID:    os.Getenv("BUILD_ID"),
	}

	base := filepath.Join("backend", "pkg", "shared", "env")
	if err := os.MkdirAll(base, os.ModePerm); err != nil {
		fmt.Println(err)
	}

	base = filepath.Join(base, "env.json")
	w, err := os.OpenFile(base, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatalf("error in os.OpenFile method: %+v", err)
	}
	defer w.Close()

	b, err := json.Marshal(env)
	if err != nil {
		log.Fatalf("error in json.Marshal method: %+v", err)
	}

	var buf bytes.Buffer
	if err := json.Indent(&buf, b, "", "	"); err != nil {
		log.Fatalf("error in json.Indent method: %+v", err)
	}
	buf.WriteByte(0xa)

	fmt.Fprint(w, buf.String())
}
