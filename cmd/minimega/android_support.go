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
	// Configure the host-side Android SDK root directory.
	//
	// This path is interpreted on the minimega host and is not treated as a
	// file served from the minimega files directory.
	AndroidSDKPath string

	// Configure the host-side Android emulator binary.
	//
	// This value may be an absolute path or a binary name resolvable via the
	// host PATH.
	AndroidEmulatorPath string

	// Configure the host-side adb binary.
	//
	// This value may be an absolute path or a binary name resolvable via the
	// host PATH.
	AndroidADBPath string

	// Configure the Android Virtual Device (AVD) name to use when launching
	// Android VMs.
	AndroidAVD string

	// Configure the host-side directory containing Android AVD data.
	//
	// This should generally contain entries such as:
	//   <avd-dir>/<avd-name>.avd
	AndroidAVDDir string

	// Launch Android VMs without creating a local emulator window.
	//
	// Default: true
	AndroidNoWindow bool

	// Configure the base console port for Android emulator instances.
	//
	// Default: 0
	AndroidConsoleBasePort uint64

	// Additional raw arguments to append to the Android emulator command line.
	AndroidExtraArgs []string

	// Require KVM support for Android runtime validation and launch.
	//
	// Default: true
	AndroidRequireKVM bool

	// Request writable system behavior for the Android emulator runtime.
	//
	// Default: false
	AndroidWritableSystem bool
}

func (old AndroidConfig) Copy() AndroidConfig {
	res := old
	res.AndroidExtraArgs = make([]string, len(old.AndroidExtraArgs))
	copy(res.AndroidExtraArgs, old.AndroidExtraArgs)
	return res
}

func (vm *AndroidConfig) String() string {
	var o bytes.Buffer
	w := new(tabwriter.Writer)
	w.Init(&o, 5, 0, 1, ' ', 0)
	fmt.Fprintln(&o, "Android configuration:")
	fmt.Fprintf(w, "SDK Path:\t%v\n", vm.AndroidSDKPath)
	fmt.Fprintf(w, "Emulator Path:\t%v\n", vm.AndroidEmulatorPath)
	fmt.Fprintf(w, "ADB Path:\t%v\n", vm.AndroidADBPath)
	fmt.Fprintf(w, "AVD Name:\t%v\n", vm.AndroidAVD)
	fmt.Fprintf(w, "AVD Dir:\t%v\n", vm.AndroidAVDDir)
	fmt.Fprintf(w, "No Window:\t%v\n", vm.AndroidNoWindow)
	fmt.Fprintf(w, "Console Base Port:\t%v\n", vm.AndroidConsoleBasePort)
	fmt.Fprintf(w, "Extra Args:\t%v\n", vm.AndroidExtraArgs)
	fmt.Fprintf(w, "Require KVM:\t%v\n", vm.AndroidRequireKVM)
	fmt.Fprintf(w, "Writable System:\t%v\n", vm.AndroidWritableSystem)
	w.Flush()
	fmt.Fprintln(&o)
	return o.String()
}

func (v *AndroidConfig) Info(field string) (string, error) {
	switch field {
	case "android-sdk":
		return v.AndroidSDKPath, nil
	case "android-emulator":
		return v.AndroidEmulatorPath, nil
	case "android-adb":
		return v.AndroidADBPath, nil
	case "android-avd":
		return v.AndroidAVD, nil
	case "android-avd-dir":
		return v.AndroidAVDDir, nil
	case "android-no-window":
		return strconv.FormatBool(v.AndroidNoWindow), nil
	case "android-console-base-port":
		return strconv.FormatUint(v.AndroidConsoleBasePort, 10), nil
	case "android-extra-args":
		return fmt.Sprintf("%v", v.AndroidExtraArgs), nil
	case "android-require-kvm":
		return strconv.FormatBool(v.AndroidRequireKVM), nil
	case "android-writable-system":
		return strconv.FormatBool(v.AndroidWritableSystem), nil
	}

	return "", fmt.Errorf("invalid info field: %v", field)
}

func (v *AndroidConfig) Clear(mask string) {
	if mask == Wildcard || mask == "android-sdk" {
		v.AndroidSDKPath = ""
	}
	if mask == Wildcard || mask == "android-emulator" {
		v.AndroidEmulatorPath = ""
	}
	if mask == Wildcard || mask == "android-adb" {
		v.AndroidADBPath = ""
	}
	if mask == Wildcard || mask == "android-avd" {
		v.AndroidAVD = ""
	}
	if mask == Wildcard || mask == "android-avd-dir" {
		v.AndroidAVDDir = ""
	}
	if mask == Wildcard || mask == "android-no-window" {
		v.AndroidNoWindow = true
	}
	if mask == Wildcard || mask == "android-console-base-port" {
		v.AndroidConsoleBasePort = 0
	}
	if mask == Wildcard || mask == "android-extra-args" {
		v.AndroidExtraArgs = nil
	}
	if mask == Wildcard || mask == "android-require-kvm" {
		v.AndroidRequireKVM = true
	}
	if mask == Wildcard || mask == "android-writable-system" {
		v.AndroidWritableSystem = false
	}
}

func (v *AndroidConfig) WriteConfig(w io.Writer) error {
	if v.AndroidSDKPath != "" {
		fmt.Fprintf(w, "vm config android-sdk %v\n", v.AndroidSDKPath)
	}
	if v.AndroidEmulatorPath != "" {
		fmt.Fprintf(w, "vm config android-emulator %v\n", v.AndroidEmulatorPath)
	}
	if v.AndroidADBPath != "" {
		fmt.Fprintf(w, "vm config android-adb %v\n", v.AndroidADBPath)
	}
	if v.AndroidAVD != "" {
		fmt.Fprintf(w, "vm config android-avd %v\n", v.AndroidAVD)
	}
	if v.AndroidAVDDir != "" {
		fmt.Fprintf(w, "vm config android-avd-dir %v\n", v.AndroidAVDDir)
	}
	if v.AndroidNoWindow != true {
		fmt.Fprintf(w, "vm config android-no-window %t\n", v.AndroidNoWindow)
	}
	if v.AndroidConsoleBasePort != 0 {
		fmt.Fprintf(w, "vm config android-console-base-port %v\n", v.AndroidConsoleBasePort)
	}
	if len(v.AndroidExtraArgs) > 0 {
		fmt.Fprintf(w, "vm config android-extra-args %v\n", quoteJoin(v.AndroidExtraArgs, " "))
	}
	if v.AndroidRequireKVM != true {
		fmt.Fprintf(w, "vm config android-require-kvm %t\n", v.AndroidRequireKVM)
	}
	if v.AndroidWritableSystem != false {
		fmt.Fprintf(w, "vm config android-writable-system %t\n", v.AndroidWritableSystem)
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
			v.AndroidSDKPath = config[1]
		case "android-emulator":
			v.AndroidEmulatorPath = config[1]
		case "android-adb":
			v.AndroidADBPath = config[1]
		case "android-avd":
			v.AndroidAVD = config[1]
		case "android-avd-dir":
			v.AndroidAVDDir = config[1]
		case "android-no-window":
			v.AndroidNoWindow, _ = strconv.ParseBool(config[1])
		case "android-console-base-port":
			v.AndroidConsoleBasePort, _ = strconv.ParseUint(config[1], 10, 64)
		case "android-extra-args":
			v.AndroidExtraArgs = fieldsQuoteEscape("\"", strings.Join(config[1:], " "))
		case "android-require-kvm":
			v.AndroidRequireKVM, _ = strconv.ParseBool(config[1])
		case "android-writable-system":
			v.AndroidWritableSystem, _ = strconv.ParseBool(config[1])
		}
	}

	return scanner.Err()
}

func androidConfiguredAny() bool {
	for _, ns := range namespaces {
		if androidConfigured(ns.vmConfig.AndroidConfig) {
			return true
		}
	}
	return false
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
	if cfg.AndroidConsoleBasePort > 0 && cfg.AndroidConsoleBasePort < 1024 {
		return fmt.Errorf("android-console-base-port must be >= 1024")
	}

	return nil
}

func androidAVDExists(cfg AndroidConfig) error {
	if cfg.AndroidAVD == "" || cfg.AndroidAVDDir == "" {
		return nil
	}

	p := filepath.Join(cfg.AndroidAVDDir, cfg.AndroidAVD+".avd")
	if _, err := os.Stat(p); err != nil {
		return err
	}

	return nil
}

func androidConfigured(cfg AndroidConfig) bool {
	return cfg.AndroidSDKPath != "" ||
		cfg.AndroidEmulatorPath != "" ||
		cfg.AndroidADBPath != "" ||
		cfg.AndroidAVD != "" ||
		cfg.AndroidAVDDir != "" ||
		cfg.AndroidConsoleBasePort != 0 ||
		len(cfg.AndroidExtraArgs) > 0 ||
		cfg.AndroidWritableSystem
}

func checkAndroidDependencies(cfg AndroidConfig) error {
	if err := validateAndroidConfig(cfg); err != nil {
		return err
	}

	if cfg.AndroidSDKPath != "" {
		if _, err := os.Stat(cfg.AndroidSDKPath); err != nil {
			return fmt.Errorf("android sdk path invalid: %v", err)
		}
	}

	if _, err := findAndroidEmulator(cfg.AndroidEmulatorPath); err != nil {
		return fmt.Errorf("android emulator not found: %v", err)
	}

	if _, err := findADB(cfg.AndroidADBPath); err != nil {
		return fmt.Errorf("android adb not found: %v", err)
	}

	if cfg.AndroidRequireKVM && !lsModule("kvm") {
		return fmt.Errorf("android requires kvm but kvm kernel module not detected")
	}

	if err := androidAVDExists(cfg); err != nil {
		return fmt.Errorf("android avd invalid: %v", err)
	}

	return nil
}
