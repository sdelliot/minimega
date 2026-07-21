package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
)

type AndroidConfig struct {
	SDKPath          string
	EmulatorPath     string
	ADBPath          string
	AVDName          string
	AVDDir           string
	WritableSystem   bool
	NoWindow         bool
	GPUMode          string
	ExtraArgs        []string
	RequireKVM       bool
	ConsoleBasePort  uint64
	ADBBasePort      uint64
	ReadOnlyTemplate bool
	TemplateDir      string
}

func (old AndroidConfig) Copy() AndroidConfig {
	res := old
	res.ExtraArgs = make([]string, len(old.ExtraArgs))
	copy(res.ExtraArgs, old.ExtraArgs)
	return res
}

func (vm *AndroidConfig) String() string {
	var o bytes.Buffer
	w := new(tabwriter.Writer)
	w.Init(&o, 5, 0, 1, ' ', 0)
	fmt.Fprintln(&o, "Android configuration:")
	fmt.Fprintf(w, "SDK Path:\t%v\n", vm.SDKPath)
	fmt.Fprintf(w, "Emulator Path:\t%v\n", vm.EmulatorPath)
	fmt.Fprintf(w, "ADB Path:\t%v\n", vm.ADBPath)
	fmt.Fprintf(w, "AVD Name:\t%v\n", vm.AVDName)
	fmt.Fprintf(w, "AVD Dir:\t%v\n", vm.AVDDir)
	fmt.Fprintf(w, "Template Dir:\t%v\n", vm.TemplateDir)
	fmt.Fprintf(w, "Read Only Template:\t%v\n", vm.ReadOnlyTemplate)
	fmt.Fprintf(w, "Writable System:\t%v\n", vm.WritableSystem)
	fmt.Fprintf(w, "No Window:\t%v\n", vm.NoWindow)
	fmt.Fprintf(w, "GPU Mode:\t%v\n", vm.GPUMode)
	fmt.Fprintf(w, "Extra Args:\t%v\n", vm.ExtraArgs)
	fmt.Fprintf(w, "Require KVM:\t%v\n", vm.RequireKVM)
	fmt.Fprintf(w, "Console Base Port:\t%v\n", vm.ConsoleBasePort)
	fmt.Fprintf(w, "ADB Base Port:\t%v\n", vm.ADBBasePort)
	w.Flush()
	fmt.Fprintln(&o)
	return o.String()
}

func (v *AndroidConfig) Info(field string) (string, error) {
	switch field {
	case "android-sdk":
		return v.SDKPath, nil
	case "android-emulator":
		return v.EmulatorPath, nil
	case "android-adb":
		return v.ADBPath, nil
	case "android-avd":
		return v.AVDName, nil
	case "android-avd-dir":
		return v.AVDDir, nil
	case "android-template-dir":
		return v.TemplateDir, nil
	case "android-writable-system":
		return strconv.FormatBool(v.WritableSystem), nil
	case "android-no-window":
		return strconv.FormatBool(v.NoWindow), nil
	case "android-gpu":
		return v.GPUMode, nil
	case "android-extra-args":
		return fmt.Sprintf("%v", v.ExtraArgs), nil
	case "android-require-kvm":
		return strconv.FormatBool(v.RequireKVM), nil
	case "android-console-base-port":
		return strconv.FormatUint(v.ConsoleBasePort, 10), nil
	case "android-adb-base-port":
		return strconv.FormatUint(v.ADBBasePort, 10), nil
	case "android-read-only-template":
		return strconv.FormatBool(v.ReadOnlyTemplate), nil
	}

	return "", fmt.Errorf("invalid info field: %v", field)
}

func (v *AndroidConfig) Clear(mask string) {
	if mask == Wildcard || mask == "android-sdk" {
		v.SDKPath = ""
	}
	if mask == Wildcard || mask == "android-emulator" {
		v.EmulatorPath = ""
	}
	if mask == Wildcard || mask == "android-adb" {
		v.ADBPath = ""
	}
	if mask == Wildcard || mask == "android-avd" {
		v.AVDName = ""
	}
	if mask == Wildcard || mask == "android-avd-dir" {
		v.AVDDir = ""
	}
	if mask == Wildcard || mask == "android-template-dir" {
		v.TemplateDir = ""
	}
	if mask == Wildcard || mask == "android-writable-system" {
		v.WritableSystem = false
	}
	if mask == Wildcard || mask == "android-no-window" {
		v.NoWindow = true
	}
	if mask == Wildcard || mask == "android-gpu" {
		v.GPUMode = ""
	}
	if mask == Wildcard || mask == "android-extra-args" {
		v.ExtraArgs = nil
	}
	if mask == Wildcard || mask == "android-require-kvm" {
		v.RequireKVM = true
	}
	if mask == Wildcard || mask == "android-console-base-port" {
		v.ConsoleBasePort = 0
	}
	if mask == Wildcard || mask == "android-adb-base-port" {
		v.ADBBasePort = 0
	}
	if mask == Wildcard || mask == "android-read-only-template" {
		v.ReadOnlyTemplate = true
	}
}

func (v *AndroidConfig) WriteConfig(w io.Writer) error {
	if v.SDKPath != "" {
		fmt.Fprintf(w, "vm config android-sdk %v\n", v.SDKPath)
	}
	if v.EmulatorPath != "" {
		fmt.Fprintf(w, "vm config android-emulator %v\n", v.EmulatorPath)
	}
	if v.ADBPath != "" {
		fmt.Fprintf(w, "vm config android-adb %v\n", v.ADBPath)
	}
	if v.AVDName != "" {
		fmt.Fprintf(w, "vm config android-avd %v\n", v.AVDName)
	}
	if v.AVDDir != "" {
		fmt.Fprintf(w, "vm config android-avd-dir %v\n", v.AVDDir)
	}
	if v.TemplateDir != "" {
		fmt.Fprintf(w, "vm config android-template-dir %v\n", v.TemplateDir)
	}
	if v.WritableSystem {
		fmt.Fprintf(w, "vm config android-writable-system true\n")
	}
	if v.NoWindow != true {
		fmt.Fprintf(w, "vm config android-no-window %t\n", v.NoWindow)
	}
	if v.GPUMode != "" {
		fmt.Fprintf(w, "vm config android-gpu %v\n", v.GPUMode)
	}
	if len(v.ExtraArgs) > 0 {
		fmt.Fprintf(w, "vm config android-extra-args %v\n", quoteJoin(v.ExtraArgs, " "))
	}
	if v.RequireKVM != true {
		fmt.Fprintf(w, "vm config android-require-kvm %t\n", v.RequireKVM)
	}
	if v.ConsoleBasePort != 0 {
		fmt.Fprintf(w, "vm config android-console-base-port %v\n", v.ConsoleBasePort)
	}
	if v.ADBBasePort != 0 {
		fmt.Fprintf(w, "vm config android-adb-base-port %v\n", v.ADBBasePort)
	}
	if v.ReadOnlyTemplate != true {
		fmt.Fprintf(w, "vm config android-read-only-template %t\n", v.ReadOnlyTemplate)
	}
	return nil
}

func (v *AndroidConfig) ReadConfig(r io.Reader, ns string) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "vm config") {
			continue
		}

		config := strings.Fields(line)[2:]
		if len(config) < 2 {
			continue
		}

		field := config[0]

		switch field {
		case "android-sdk":
			v.SDKPath = config[1]
		case "android-emulator":
			v.EmulatorPath = config[1]
		case "android-adb":
			v.ADBPath = config[1]
		case "android-avd":
			v.AVDName = config[1]
		case "android-avd-dir":
			v.AVDDir = config[1]
		case "android-template-dir":
			v.TemplateDir = config[1]
		case "android-writable-system":
			v.WritableSystem, _ = strconv.ParseBool(config[1])
		case "android-no-window":
			v.NoWindow, _ = strconv.ParseBool(config[1])
		case "android-gpu":
			v.GPUMode = config[1]
		case "android-extra-args":
			v.ExtraArgs = strings.Fields(strings.Join(config[1:], " "))
		case "android-require-kvm":
			v.RequireKVM, _ = strconv.ParseBool(config[1])
		case "android-console-base-port":
			v.ConsoleBasePort, _ = strconv.ParseUint(config[1], 10, 64)
		case "android-adb-base-port":
			v.ADBBasePort, _ = strconv.ParseUint(config[1], 10, 64)
		case "android-read-only-template":
			v.ReadOnlyTemplate, _ = strconv.ParseBool(config[1])
		}
	}

	return scanner.Err()
}

func findAndroidEmulator(path string) (string, error) {
	if path != "" {
		if _, err := os.Stat(path); err != nil {
			return "", err
		}
		return path, nil
	}

	return exec.LookPath("emulator")
}

func findADB(path string) (string, error) {
	if path != "" {
		if _, err := os.Stat(path); err != nil {
			return "", err
		}
		return path, nil
	}

	return exec.LookPath("adb")
}

func validateAndroidConfig(cfg AndroidConfig) error {
	if cfg.ConsoleBasePort > 0 && cfg.ConsoleBasePort < 1024 {
		return fmt.Errorf("android-console-base-port must be >= 1024")
	}

	if cfg.ADBBasePort > 0 && cfg.ADBBasePort < 1024 {
		return fmt.Errorf("android-adb-base-port must be >= 1024")
	}

	return nil
}

func androidAVDExists(cfg AndroidConfig) error {
	if cfg.AVDName == "" || cfg.AVDDir == "" {
		return nil
	}

	p := filepath.Join(cfg.AVDDir, cfg.AVDName+".avd")
	if _, err := os.Stat(p); err != nil {
		return err
	}

	return nil
}

func checkAndroidDependencies() error {
	ns := GetNamespace()
	cfg := ns.vmConfig.AndroidConfig

	if err := validateAndroidConfig(cfg); err != nil {
		return err
	}

	if cfg.SDKPath != "" {
		if _, err := os.Stat(cfg.SDKPath); err != nil {
			return fmt.Errorf("android sdk path invalid: %v", err)
		}
	}

	if _, err := findAndroidEmulator(cfg.EmulatorPath); err != nil {
		return fmt.Errorf("android emulator not found: %v", err)
	}

	if _, err := findADB(cfg.ADBPath); err != nil {
		return fmt.Errorf("android adb not found: %v", err)
	}

	if cfg.RequireKVM && !lsModule("kvm") {
		return fmt.Errorf("android requires kvm but kvm kernel module not detected")
	}

	if err := androidAVDExists(cfg); err != nil {
		return fmt.Errorf("android avd invalid: %v", err)
	}

	return nil
}
