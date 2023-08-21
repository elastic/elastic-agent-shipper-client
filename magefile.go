// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

//go:build mage
// +build mage

package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	devtools "github.com/elastic/elastic-agent-libs/dev-tools/mage"
	"github.com/elastic/elastic-agent-libs/dev-tools/mage/gotool"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

const (
	protoDest = "./pkg/proto"

	goProtocGenGo     = "google.golang.org/protobuf/cmd/protoc-gen-go@v1.28"
	goProtocGenGoGRPC = "google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2"
	goLicenserRepo    = "github.com/elastic/go-licenser@v0.4.1"

	goBenchstats = "golang.org/x/perf/cmd/benchstat@v0.0.0-20230227161431-f7320a6d63e8"
)

var (

	// Add here new packages that have to be compiled.
	// Vendor packages are not included since they already have compiled versions.
	// All `.proto` files in the listed directories will be compiled to Go.
	protoPackagesToCompile = []string{
		"api",
		"api/messages",
	}

	// List all the protobuf packages that need to be included
	protoPackages = append(
		protoPackagesToCompile,
		"api/vendor",
	)

	// Add here files that have their own license that must remain untouched
	goLicenserExcluded = []string{
		"api/vendor",
		"api/messages/struct.proto",
		"pkg/proto/messages/struct.pb.go",
		"pkg/helpers/struct.go",
	}

	benchmarkCount int = 8
	// Add the packages to run go benchmark
	goBenchmarkPackages = []string{
		"pkg/helpers",
	}
)

// Update updates all the generated code out of the spec
func Update() {
	mg.SerialDeps(GenerateGo, License)
}

// InstallProtoGo installs required plugins for protoc
func InstallProtoGo() error {
	err := gotool.Install(gotool.Install.Package(goProtocGenGo))
	if err != nil {
		return err
	}
	err = gotool.Install(gotool.Install.Package(goProtocGenGoGRPC))
	if err != nil {
		return err
	}
	return nil
}

// InstallLicenser installs the go-licenser.
// For some reason `devtools.InstallGoLicenser` fails with strange errors, this solution is stable.
func InstallLicenser() error {
	return gotool.Install(gotool.Install.Package(goLicenserRepo))
}

// GenerateGo regenerates the Go files out of .proto files
func GenerateGo() error {
	mg.Deps(InstallProtoGo)

	var (
		importFlags []string
		toCompile   []string
	)

	for _, p := range protoPackages {
		importFlags = append(importFlags, "-I"+p)
	}

	for _, p := range protoPackagesToCompile {
		log.Printf("Listing the %s package...\n", p)

		files, err := ioutil.ReadDir(p)
		if err != nil {
			return fmt.Errorf("failed to read the proto package directory %s: %w", p, err)
		}
		for _, f := range files {
			if path.Ext(f.Name()) != ".proto" {
				continue
			}
			toCompile = append(toCompile, path.Join(p, f.Name()))
		}
	}

	args := append(
		[]string{
			"--go_out=" + protoDest,
			"--go-grpc_out=" + protoDest,
			"--go_opt=paths=source_relative",
			"--go-grpc_opt=paths=source_relative",
		},
		importFlags...,
	)

	args = append(args, toCompile...)

	log.Printf("Compiling %d packages...\n", len(protoPackages))
	err := sh.Run("protoc", args...)
	if err != nil {
		return fmt.Errorf("failed to compile protobuf: %w", err)
	}

	return nil
}

// Check runs all the checks
func Check() {
	mg.Deps(devtools.Deps.CheckModuleTidy, CheckLicenseHeaders)
	mg.Deps(devtools.CheckNoChanges)
}

// License applies the right license header.
func License() error {
	mg.Deps(InstallLicenser)
	log.Println("Adding license headers...")

	return licenser(rewriteHeader)
}

// CheckLicenseHeaders checks ASL2 headers in .go files
func CheckLicenseHeaders() error {
	mg.Deps(InstallLicenser)
	return licenser(checkHeader)
}

type licenserMode int

var (
	rewriteHeader licenserMode = 1
	checkHeader   licenserMode = 2
)

func licenser(mode licenserMode) error {
	var args []string

	switch mode {
	case checkHeader:
		args = append(args, "-d")
	}

	for _, e := range goLicenserExcluded {
		args = append(args, "-exclude", e)
	}

	args = append(args, "-license", "Elastic")

	// go-licenser does not support multiple extensions at the same time,
	// so we have to run it twice

	err := sh.RunV("go-licenser", append(args, "-ext", ".go")...)
	if err != nil {
		return fmt.Errorf("failed to process .go files: %w", err)
	}

	err = sh.RunV("go-licenser", append(args, "-ext", ".proto")...)
	if err != nil {
		return fmt.Errorf("failed to process .proto files: %w", err)
	}

	return nil
}

type Benchmark mg.Namespace

func InstallBenchStat() error {
	err := gotool.Install(gotool.Install.Package(goBenchstats))
	if err != nil {
		return err
	}
	return nil
}

func (Benchmark) Run(ctx context.Context, outputFile string) error {
	mg.Deps(InstallBenchStat)
	fmt.Println(">> go benchmark:", "Testing")
	args := []string{
		"test",
		fmt.Sprintf("-count=%d", benchmarkCount),
		"-bench=.",
		"-run=Bench#",
	}
	for _, pkg := range goBenchmarkPackages {
		args = append(args, filepath.Join("github.com/elastic/elastic-agent-shipper-client", pkg, "..."))
	}
	goTestBench := makeCommand(ctx, nil, "go", args...)

	// Wire up the outputs.
	var outputs []io.Writer
	if outputFile != "" {
		fileOutput, err := os.Create(createDir(outputFile))
		if err != nil {
			return errors.Wrap(err, "failed to create go test output file")
		}
		defer fileOutput.Close()
		outputs = append(outputs, fileOutput)
	}
	output := io.MultiWriter(outputs...)
	goTestBench.Stdout = io.MultiWriter(output, os.Stdout)
	goTestBench.Stderr = io.MultiWriter(output, os.Stderr)

	err := goTestBench.Run()

	var goTestErr *exec.ExitError
	if err != nil {
		// Command ran.
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			return errors.Wrap(err, "failed to execute go")
		}

		// Command ran but failed. Process the output.
		goTestErr = exitErr
	}

	if goTestErr != nil {
		// No packages were tested. Probably the code didn't compile.
		return errors.Wrap(goTestErr, "go test returned a non-zero value")
	}

	return nil
}

func (Benchmark) Diff(ctx context.Context, baseFile string, newFile string, outputFile string) error {
	mg.Deps(InstallBenchStat)
	var args = []string{}
	if outputFile == "" {
		outputFile = "benchmark_stats"
	}
	if baseFile == "" {
		log.Printf("Missing baseline benchmark output")
	} else {
		args = append(args, baseFile)
	}

	if newFile == "" {
		return errors.New("Missing benchmark output file, please run first benchmark:all")
	} else {
		args = append(args, newFile)
	}

	gobench := makeCommand(ctx, nil, "benchstat", args...)

	// Wire up the outputs.
	var outputs []io.Writer
	if outputFile != "" {
		fileOutput, err := os.Create(createDir(outputFile))
		if err != nil {
			return errors.Wrap(err, "failed to create go test output file")
		}
		defer fileOutput.Close()
		outputs = append(outputs, fileOutput)
	}
	output := io.MultiWriter(outputs...)
	gobench.Stdout = io.MultiWriter(output, os.Stdout)
	gobench.Stderr = io.MultiWriter(output, os.Stderr)

	err := gobench.Run()

	var goTestErr *exec.ExitError
	if err != nil {
		// Command ran.
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			return errors.Wrap(err, "failed to execute go")
		}

		// Command ran but failed. Process the output.
		goTestErr = exitErr
	}

	if goTestErr != nil {
		// No packages were tested. Probably the code didn't compile.
		return errors.Wrap(goTestErr, "go test returned a non-zero value")
	}

	return nil
}

func makeCommand(ctx context.Context, env map[string]string, cmd string, args ...string) *exec.Cmd {
	c := exec.CommandContext(ctx, cmd, args...)
	c.Env = os.Environ()
	for k, v := range env {
		c.Env = append(c.Env, k+"="+v)
	}
	c.Stdout = ioutil.Discard
	if mg.Verbose() {
		c.Stdout = os.Stdout
	}
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	log.Println("exec:", cmd, strings.Join(args, " "))
	fmt.Println("exec:", cmd, strings.Join(args, " "))
	return c
}

// CreateDir creates the parent directory for the given file.
func createDir(file string) string {
	// Create the output directory.
	if dir := filepath.Dir(file); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic(errors.Wrapf(err, "failed to create parent dir for %v", file))
		}
	}
	return file
}
