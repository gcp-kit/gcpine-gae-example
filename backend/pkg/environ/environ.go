package environ

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/gcp-kit/gcpen"
)

// IsTest - variable to determine if it is a test.
var IsTest = func() bool {
	return true
}()

// IsLocal - Variable to determine if the environment is local or not.
var IsLocal = func() bool {
	return gcpen.ProjectID == ""
}()

var local = "local"

var (
	BranchName = func() string {
		name, ok := os.LookupEnv("BRANCH_NAME")
		if !ok || name == "main" {
			return ""
		}
		return name
	}()
	CommitHash = func() string {
		hash, ok := os.LookupEnv("SHORT_SHA")
		if !ok {
			return local
		}
		return hash
	}()
	TimeZoneJST = time.FixedZone("JST", 9*60*60)
)

func init() {
	type Env struct {
		BranchName string `json:"BRANCH_NAME"`
		ShortSHA   string `json:"SHORT_SHA"`
	}

	if !IsLocal {
		filePath := filepath.Join("backend", "pkg", "shared", "env")
		// NOTE: Simple path operation
		{
			if f, err := os.Stat(filePath); err != nil || !f.IsDir() {
				if f, err := os.Stat("../" + filePath); err == nil && f.IsDir() {
					filePath = "../" + filePath
				} else if f, err := os.Stat("/workspace/serverless_function_source_code/" + filePath); err == nil && f.IsDir() {
					// NOTE: for GCF
					filePath = "/workspace/serverless_function_source_code/" + filePath
				}
			}
			filePath = filepath.Join(filePath, "env.json")
		}

		raw, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Println(err.Error())
			fmt.Println(`try: export BRANCH_NAME="XXX" SHORT_SHA="XXX"`)
			fmt.Println("run: go run tools/set_env/main.go")
			os.Exit(1)
		}
		env := new(Env)
		if err = json.Unmarshal(raw, env); err != nil {
			panic(err)
		}
		BranchName = env.BranchName
		CommitHash = env.ShortSHA
	}
}
