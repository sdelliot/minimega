package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
