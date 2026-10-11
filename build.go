//go:build ignore
// +build ignore

/*
Copyright (c) 2019 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	VERSION     = "v0.27.0"
	GOPATH      = os.Getenv("GOPATH")
	GIT_COMMIT  = gitCommit()
	BUILD_TIME  = time.Now().UTC().Format(time.RFC3339)
	LD_FLAGS    = fmt.Sprintf("-X \"main.buildTime=%s\" -X main.gitCommit=%s", BUILD_TIME, GIT_COMMIT)
	GO_FLAGS    = fmt.Sprintf("-ldflags=%s", LD_FLAGS)
	IMAGE_FLAGS = "exclude_graphdriver_devicemapper exclude_graphdriver_btrfs containers_image_openpgp"

	// targetOS/targetArch select the platform of the compiled server binary.
	// They default to the host and can be overridden with --target-os/--target-arch
	// (e.g. to ship the macOS desktop sidecar as Apple Silicon while building on
	// any runner). The build.go process itself always runs on the host.
	targetOS   = runtime.GOOS
	targetArch = runtime.GOARCH
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "build.go",
		Short: "Build tools for Octant",
	}

	rootCmd.Name()

	rootCmd.PersistentFlags().StringVar(&targetOS, "target-os", runtime.GOOS,
		"target GOOS for the compiled server binary (default: host OS)")
	rootCmd.PersistentFlags().StringVar(&targetArch, "target-arch", runtime.GOARCH,
		"target GOARCH for the compiled server binary (default: host arch)")

	rootCmd.AddCommand(
		&cobra.Command{
			Use:   "ci",
			Short: "full build, running tests",
			Run: func(cmd *cobra.Command, args []string) {
				test()
				vet()
				webDeps()
				webTest()
				webBuild()
				build()
			},
		},
		&cobra.Command{
			Use:   "ci-quick",
			Short: "full build, skipping tests",
			Run: func(cmd *cobra.Command, args []string) {
				webDeps("--prefer-offline", "--no-audit")
				webBuild()
				build()
			},
		},
		&cobra.Command{
			Use:   "web-deps",
			Short: "install client dependencies",
			Run: func(cmd *cobra.Command, args []string) {
				webDeps()
			},
		},
		&cobra.Command{
			Use:   "web-test",
			Short: "run client tests",
			Run: func(cmd *cobra.Command, args []string) {
				verifyRegistry()
				verifyNpmCache()
				webDeps("--prefer-offline", "--no-audit")
				webTest()
			},
		},
		&cobra.Command{
			Use:   "web-build",
			Short: "client build, skipping tests",
			Run: func(cmd *cobra.Command, args []string) {
				verifyRegistry()
				verifyNpmCache()
				webDeps("--prefer-offline", "--no-audit")
				webBuild()
			},
		},
		&cobra.Command{
			Use:   "web-lint",
			Short: "run client linter",
			Run: func(cmd *cobra.Command, args []string) {
				verifyRegistry()
				webDeps()
				webLint()
			},
		},
		&cobra.Command{
			Use:   "web",
			Short: "client build, running tests",
			Run: func(cmd *cobra.Command, args []string) {
				webDeps()
				webTest()
				webBuild()
			},
		},
		&cobra.Command{
			Use:   "generate",
			Short: "update generated artifacts",
			Run: func(cmd *cobra.Command, args []string) {
				generate()
				goFmt(true)
			},
		},
		&cobra.Command{
			Use:   "vet",
			Short: "lint server code",
			Run: func(cmd *cobra.Command, args []string) {
				vet()
			},
		},
		&cobra.Command{
			Use:   "fmt",
			Short: "format server code",
			Run: func(cmd *cobra.Command, args []string) {
				goFmt(true)
			},
		},
		&cobra.Command{
			Use:   "test",
			Short: "run server tests",
			Run: func(cmd *cobra.Command, args []string) {
				test()
			},
		},
		&cobra.Command{
			Use:   "verify",
			Short: "verify resolving correct registry",
			Run: func(cmd *cobra.Command, args []string) {
				verifyRegistry()
			},
		},
		&cobra.Command{
			Use:   "build",
			Short: "server build, skipping tests",
			Run: func(cmd *cobra.Command, args []string) {
				build()
			},
		},
		&cobra.Command{
			Use:   "build-electron",
			Short: "server build to extraResources, skipping tests",
			Run: func(cmd *cobra.Command, args []string) {
				webDeps("--prefer-offline", "--no-audit")
				buildElectron()
				webBuildElectron()
			},
		},
		&cobra.Command{
			Use:   "dist-electron",
			Short: "build desktop installers (macOS: arm64 + universal)",
			Run: func(cmd *cobra.Command, args []string) {
				distElectron()
			},
		},
		&cobra.Command{
			Use:   "run-dev",
			Short: "run ci produced build",
			Run: func(cmd *cobra.Command, args []string) {
				runDev()
			},
		},
		&cobra.Command{
			Use:   "go-install",
			Short: "install build tools",
			Run: func(cmd *cobra.Command, args []string) {
				goInstall()
			},
		},
		&cobra.Command{
			Use:   "serve",
			Short: "start client and server in development mode",
			Run: func(cmd *cobra.Command, args []string) {
				serve()
			},
		},
		&cobra.Command{
			Use:   "install-test-plugin",
			Short: "build the sample plugin",
			Run: func(cmd *cobra.Command, args []string) {
				installTestPlugin()
			},
		},
		&cobra.Command{
			Use:   "version",
			Short: "",
			Run: func(cmd *cobra.Command, args []string) {
				version()
			},
		},
		&cobra.Command{
			Use:   "release",
			Short: "tag and push a release",
			Run: func(cmd *cobra.Command, args []string) {
				release()
			},
		},
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runCmd(command string, env map[string]string, args ...string) {
	cmd := newCmd(command, env, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	log.Printf("Running: %s\n", cmd.String())
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func runCmdIn(dir, command string, env map[string]string, args ...string) {
	cmd := newCmd(command, env, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	log.Printf("Running in %s: %s\n", dir, cmd.String())
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func newCmd(command string, env map[string]string, args ...string) *exec.Cmd {
	realCommand, err := exec.LookPath(command)
	if err != nil {
		log.Fatalf("unable to find command '%s'", command)
	}

	cmd := exec.Command(realCommand, args...)
	cmd.Stderr = os.Stderr

	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	return cmd
}

func goInstall() {
	pkgs := []string{
		"github.com/golang/mock/gomock",
		"github.com/golang/mock/mockgen",
		"google.golang.org/protobuf/cmd/protoc-gen-go",
		"google.golang.org/grpc/cmd/protoc-gen-go-grpc",
		"golang.org/x/tools/cmd/goimports",
	}
	for _, pkg := range pkgs {
		runCmd("go", map[string]string{"GO111MODULE": "on"}, "install", pkg)
	}
}

func generate() {
	removeFakes()
	runCmd("go", nil, "generate", "-tags", IMAGE_FLAGS, "-v", "./pkg/...", "./internal/...")
}

func build() {
	newPath := filepath.Join(".", "build")
	os.MkdirAll(newPath, 0755)

	artifact := "octant"
	if targetOS == "windows" {
		artifact = "octant.exe"
	}
	buildServerBinary(filepath.Join("build", artifact), "embedded "+IMAGE_FLAGS, targetArch)
}

// buildElectron builds an Octant binary without web assets
func buildElectron() {
	cleanCmd := newCmd("npm", nil, "run", "clean")
	cleanCmd.Stdout = os.Stdout
	cleanCmd.Stderr = os.Stderr
	cleanCmd.Stdin = os.Stdin
	cleanCmd.Dir = "./web"
	if err := cleanCmd.Run(); err != nil {
		log.Fatalf("web-build-electron: create dist/octant/ : %s", err)
	}

	newPath := filepath.Join(".", "build")
	os.MkdirAll(newPath, 0755)

	artifact := "octant"
	if targetOS == "windows" {
		artifact = "octant.exe"
	}
	outPath := filepath.Join("web", "extraResources", artifact)
	log.Printf("Building server sidecar for %s/%s -> %s\n", targetOS, targetArch, outPath)
	buildServerBinary(outPath, IMAGE_FLAGS, targetArch)
}

// distElectron builds the desktop installers.
//
// On macOS it produces two independent artifacts so users can pick the smaller
// native build or the one that also runs on Intel:
//
//	Octant-<version>-arm64.dmg      (Apple Silicon only, native sidecar)
//	Octant-<version>-universal.dmg  (Apple Silicon + Intel, universal sidecar)
//
// electron-builder is invoked once per architecture with the matching Go
// sidecar in web/extraResources, because a single run can only carry one
// sidecar file.
func distElectron() {
	webDeps("--prefer-offline", "--no-audit")

	cleanCmd := newCmd("npm", nil, "run", "clean")
	cleanCmd.Stdout = os.Stdout
	cleanCmd.Stderr = os.Stderr
	cleanCmd.Stdin = os.Stdin
	cleanCmd.Dir = "./web"
	if err := cleanCmd.Run(); err != nil {
		log.Fatalf("dist-electron: clean: %s", err)
	}

	// Compile the main process and build the renderer once; both
	// electron-builder runs reuse dist/ and main.js.
	runCmdIn("./web", "npm", nil, "run", "build-electron")

	artifact := "octant"
	if targetOS == "windows" {
		artifact = "octant.exe"
	}
	sidecar := filepath.Join("web", "extraResources", artifact)

	if targetOS == "darwin" {
		log.Printf("Packaging macOS universal app (Apple Silicon + Intel)")
		buildServerBinary(sidecar, IMAGE_FLAGS, "universal")
		packageElectron("--mac", "--universal")

		log.Printf("Packaging macOS arm64 app (Apple Silicon only)")
		buildServerBinary(sidecar, IMAGE_FLAGS, "arm64")
		packageElectron("--mac", "--arm64")
		return
	}

	log.Printf("Packaging %s/%s app", targetOS, targetArch)
	buildServerBinary(sidecar, IMAGE_FLAGS, targetArch)
	packageElectron()
}

func packageElectron(args ...string) {
	npmArgs := append([]string{"run", "electron:package", "--"}, args...)
	runCmdIn("./web", "npm", nil, npmArgs...)
}

// buildServerBinary compiles cmd/octant for the requested GOARCH.
//
// When goarch is "universal" the amd64 and arm64 binaries are built separately
// and merged into a single fat Mach-O with lipo. Universal is only meaningful
// on darwin.
func buildServerBinary(outPath string, tags string, goarch string) {
	if goarch != "universal" {
		runCmd("go", serverBuildEnv(targetOS, goarch), "build", "-tags", tags, "-mod=vendor", "-o", outPath, GO_FLAGS, "-v", "./cmd/octant")
		return
	}

	if targetOS != "darwin" {
		log.Fatalf("universal target is only supported on darwin, got %s", targetOS)
	}

	amd64Out := outPath + "-amd64"
	arm64Out := outPath + "-arm64"
	runCmd("go", serverBuildEnv(targetOS, "amd64"), "build", "-tags", tags, "-mod=vendor", "-o", amd64Out, GO_FLAGS, "-v", "./cmd/octant")
	runCmd("go", serverBuildEnv(targetOS, "arm64"), "build", "-tags", tags, "-mod=vendor", "-o", arm64Out, GO_FLAGS, "-v", "./cmd/octant")
	runCmd("lipo", nil, "-create", "-output", outPath, amd64Out, arm64Out)

	for _, tmp := range []string{amd64Out, arm64Out} {
		if err := os.Remove(tmp); err != nil {
			log.Printf("warning: unable to remove %s: %s", tmp, err)
		}
	}
}

// serverBuildEnv returns the environment overrides used to compile the server
// binary for the requested target platform. Cross-compiling works because the
// build tags keep cgo out of the dependency tree.
func serverBuildEnv(goos, goarch string) map[string]string {
	return map[string]string{
		"CGO_ENABLED": "0",
		"GOOS":        goos,
		"GOARCH":      goarch,
	}
}

func runDev() {
	env := make(map[string]string)
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		env[parts[0]] = parts[1]
	}
	runCmd("build/octant", env)
}

func test() {
	runCmd("go", nil, "test", "-tags", IMAGE_FLAGS, "-v", "./internal/...", "./pkg/...")
}

func vet() {
	runCmd("go", nil, "vet", "-tags", IMAGE_FLAGS, "./internal/...", "./pkg/...")
	goFmt(false)
}

func goFmt(update bool) {
	if update {
		runCmd("goimports", nil, "--local", "github.com/vmware-tanzu/octant", "-w", "cmd", "internal", "pkg")
	} else {
		out := bytes.NewBufferString("")
		cmd := newCmd("goimports", nil, "--local", "github.com/vmware-tanzu/octant", "-l", "cmd", "internal", "pkg")
		cmd.Stdout = out
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		log.Printf("Running: %s\n", cmd.String())
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
		if out.Len() != 0 {
			os.Stdout.Write(out.Bytes())
			log.Fatal("above files are not formatted correctly. please run `go run build.go fmt`")
		}
	}
}

func webDeps(opts ...string) {
	// The lockfile is resolved with --legacy-peer-deps: @storybook/angular@6.3.8
	// declares a peer of rxjs ^6 while the app is on rxjs 7, so a plain
	// `npm ci` aborts on the peer conflict. Keep this in sync with the flag the
	// sandbox/CI installs already use.
	args := append([]string{"ci", "--legacy-peer-deps"}, opts...)
	cmd := newCmd("npm", nil, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = "./web"
	if err := cmd.Run(); err != nil {
		log.Fatalf("web-deps: %s", err)
	}
}

func webTest() {
	cmd := newCmd("npm", nil, "run", "test:headless")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = "./web"
	if err := cmd.Run(); err != nil {
		log.Fatalf("web-test: %s", err)
	}
}

func webBuild() {
	cmd := newCmd("npm", nil, "run", "build")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = "./web"
	if err := cmd.Run(); err != nil {
		log.Fatalf("web-build: %s", err)
	}
}

func webBuildElectron() {
	cmd := newCmd("npm", nil, "run", "electron:build")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = "./web"
	if err := cmd.Run(); err != nil {
		log.Fatalf("web-build-electron: build : %s", err)
	}
}

func webLint() {
	cmd := newCmd("npm", nil, "run", "lint")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = "./web"
	if err := cmd.Run(); err != nil {
		log.Fatalf("web-lint: %s", err)
	}
}

func serve() {
	var wg sync.WaitGroup

	uiVars := map[string]string{"API_BASE": "http://localhost:7777"}
	uiCmd := newCmd("npm", uiVars, "run", "start")
	uiCmd.Stdout = os.Stdout
	uiCmd.Stderr = os.Stderr
	uiCmd.Stdin = os.Stdin
	uiCmd.Dir = "./web"
	if err := uiCmd.Start(); err != nil {
		log.Fatalf("uiCmd: start: %s", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := uiCmd.Wait(); err != nil {
			log.Fatalf("serve: npm run: %s", err)
		}
	}()

	serverVars := map[string]string{
		"OCTANT_DISABLE_OPEN_BROWSER": "true",
		"OCTANT_LISTENER_ADDR":        "localhost:7777",
		"OCTANT_PROXY_FRONTEND":       "http://localhost:4200",
	}
	serverCmd := newCmd("go", serverVars, "run", "./cmd/octant/main.go")
	serverCmd.Stdout = os.Stdout
	if err := serverCmd.Start(); err != nil {
		log.Fatalf("serveCmd: start: %s", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := serverCmd.Wait(); err != nil {
			log.Fatalf("serve: go run: %s", err)
		}
	}()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-sigc
		uiCmd.Process.Signal(syscall.SIGQUIT)
		serverCmd.Process.Signal(syscall.SIGQUIT)
	}()

	wg.Wait()
}

func installTestPlugin() {
	dir := pluginDir()
	log.Printf("Plugin path: %s", dir)
	os.MkdirAll(dir, 0755)
	filename := "octant-sample-plugin"
	if runtime.GOOS == "windows" {
		filename = "octant-sample-plugin.exe"
	}
	pluginFile := filepath.Join(dir, filename)
	runCmd("go", nil, "build", "-o", pluginFile, "github.com/vmware-tanzu/octant/cmd/octant-sample-plugin")
}

func version() {
	fmt.Println(VERSION)
}

func release() {
	runCmd("git", nil, "tag", "-a", VERSION, "-m", fmt.Sprintf("\"Release %s\"", VERSION))
	runCmd("git", nil, "push", "--follow-tags")
}

func removeFakes() {
	checkDirs := []string{"pkg", "internal"}
	fakePaths := []string{}

	for _, dir := range checkDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				return nil
			}
			if info.Name() == "fake" {
				fakePaths = append(fakePaths, path)
			}
			return nil
		})
		if err != nil {
			log.Fatalf("generate (%s): %s", dir, err)
		}
	}

	log.Print("Removing fakes from pkg/ and internal/")
	for _, p := range fakePaths {
		os.RemoveAll(p)
	}
}

func gitCommit() string {
	cmd := newCmd("git", nil, "rev-parse", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		log.Printf("gitCommit: %s", err)
		return ""
	}
	return fmt.Sprintf("%s", out)
}

func pluginDir() string {
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "octant", "plugins")
	} else if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "octant", "plugins")
	} else {
		return filepath.Join(os.Getenv("HOME"), ".config", "octant", "plugins")
	}
}

func verifyRegistry() {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = newCmd("findstr", nil, "/S", "/C:\"build-artifactory.eng.vmware.com\"", "web/*")
	} else {
		cmd = newCmd("grep", nil, "-R", "build-artifactory.eng.vmware.com", "web")
	}

	out, err := cmd.Output()
	if exitError, ok := err.(*exec.ExitError); ok {
		if exitError.ExitCode() > 1 {
			log.Fatalf("grep: %s", err)
		}
	}
	if len(out) > 0 {
		log.Fatalf("found registry: %s", string(out))
	}
}

func verifyNpmCache() {
	cmd := newCmd("npm", nil, "cache", "verify")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = "./web"
	if err := cmd.Run(); err != nil {
		log.Fatalf("NPM cache verify: %s", err)
	}
}
