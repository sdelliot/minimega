package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParseDiskSpec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		spec    string
		want    *diskSpecParts
		wantErr string
	}{
		{
			name: "path only",
			spec: "foo.qcow2",
			want: &diskSpecParts{
				Path: "foo.qcow2",
			},
		},
		{
			name: "path and interface",
			spec: "foo.qcow2,virtio",
			want: &diskSpecParts{
				Path:      "foo.qcow2",
				Interface: "virtio",
			},
		},
		{
			name: "path and cache",
			spec: "foo.qcow2,writeback",
			want: &diskSpecParts{
				Path:  "foo.qcow2",
				Cache: "writeback",
			},
		},
		{
			name: "path interface cache",
			spec: "foo.qcow2,virtio,writeback",
			want: &diskSpecParts{
				Path:      "foo.qcow2",
				Interface: "virtio",
				Cache:     "writeback",
			},
		},
		{
			name: "ahci interface with cache",
			spec: "foo.qcow2,ahci,directsync",
			want: &diskSpecParts{
				Path:      "foo.qcow2",
				Interface: "ahci",
				Cache:     "directsync",
			},
		},
		{
			name: "file path only",
			spec: "file:images/foo.qcow2",
			want: &diskSpecParts{
				Path: "file:images/foo.qcow2",
			},
		},
		{
			name: "file path and interface",
			spec: "file:images/foo.qcow2,virtio",
			want: &diskSpecParts{
				Path:      "file:images/foo.qcow2",
				Interface: "virtio",
			},
		},
		{
			name: "file path and cache",
			spec: "file:images/foo.qcow2,writeback",
			want: &diskSpecParts{
				Path:  "file:images/foo.qcow2",
				Cache: "writeback",
			},
		},
		{
			name: "file path interface cache",
			spec: "file:images/foo.qcow2,virtio,writeback",
			want: &diskSpecParts{
				Path:      "file:images/foo.qcow2",
				Interface: "virtio",
				Cache:     "writeback",
			},
		},
		{
			name:    "bad second field",
			spec:    "foo.qcow2,bad",
			wantErr: "malformed diskspec",
		},
		{
			name:    "bad third field",
			spec:    "foo.qcow2,virtio,bad",
			wantErr: "malformed diskspec",
		},
		{
			name:    "too many fields",
			spec:    "foo.qcow2,virtio,writeback,extra",
			wantErr: "malformed diskspec",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseDiskSpec(tt.spec)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("unexpected error: got %q want %q", err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("unexpected parsed diskspec:\n got: %#v\nwant: %#v", got, tt.want)
			}
		})
	}
}

func TestPreprocessDiskSpecPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		in        *diskSpecParts
		want      *diskSpecParts
		wantErr   string
		wantCalls []string
	}{
		{
			name: "preprocess path only preserve interface and cache",
			in: &diskSpecParts{
				Path:      "file:foo.qcow2",
				Interface: "virtio",
				Cache:     "writeback",
			},
			want: &diskSpecParts{
				Path:      "/tmp/minimega/files/foo.qcow2",
				Interface: "virtio",
				Cache:     "writeback",
			},
			wantCalls: []string{"file:foo.qcow2"},
		},
		{
			name: "preprocess path only preserve cache",
			in: &diskSpecParts{
				Path:  "file:bar.qcow2",
				Cache: "writeback",
			},
			want: &diskSpecParts{
				Path:  "/tmp/minimega/files/bar.qcow2",
				Cache: "writeback",
			},
			wantCalls: []string{"file:bar.qcow2"},
		},
		{
			name: "preprocess path only no suffixes",
			in: &diskSpecParts{
				Path: "file:baz.qcow2",
			},
			want: &diskSpecParts{
				Path: "/tmp/minimega/files/baz.qcow2",
			},
			wantCalls: []string{"file:baz.qcow2"},
		},
		{
			name: "propagate preprocess error",
			in: &diskSpecParts{
				Path:      "file:nope.qcow2",
				Interface: "virtio",
				Cache:     "writeback",
			},
			wantErr:   "file not found",
			wantCalls: []string{"file:nope.qcow2"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var calls []string
			preprocess := func(s string) (string, error) {
				calls = append(calls, s)

				switch s {
				case "file:foo.qcow2":
					return "/tmp/minimega/files/foo.qcow2", nil
				case "file:bar.qcow2":
					return "/tmp/minimega/files/bar.qcow2", nil
				case "file:baz.qcow2":
					return "/tmp/minimega/files/baz.qcow2", nil
				case "file:nope.qcow2":
					return "", fmt.Errorf("file not found")
				default:
					return "", fmt.Errorf("unexpected preprocess input: %s", s)
				}
			}

			got, err := preprocessDiskSpecPath(tt.in, preprocess)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("unexpected error: got %q want %q", err.Error(), tt.wantErr)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Fatalf("unexpected preprocessed diskspec:\n got: %#v\nwant: %#v", got, tt.want)
				}
			}

			if !reflect.DeepEqual(calls, tt.wantCalls) {
				t.Fatalf("unexpected preprocess calls:\n got: %#v\nwant: %#v", calls, tt.wantCalls)
			}
		})
	}
}

func TestDiskConfigFromParts(t *testing.T) {
	t.Parallel()

	in := &diskSpecParts{
		Path:      "/tmp/minimega/files/foo.qcow2",
		Interface: "virtio",
		Cache:     "writeback",
	}

	got := diskConfigFromParts(in)
	want := &DiskConfig{
		Path:      "/tmp/minimega/files/foo.qcow2",
		Interface: "virtio",
		Cache:     "writeback",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected disk config:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestIsCache(t *testing.T) {
	t.Parallel()

	valid := []string{
		"none",
		"writeback",
		"unsafe",
		"directsync",
		"writethrough",
	}

	for _, v := range valid {
		if !isCache(v) {
			t.Fatalf("expected cache mode %q to be valid", v)
		}
	}

	invalid := []string{
		"",
		"virtio",
		"bad",
		"e1000",
	}

	for _, v := range invalid {
		if isCache(v) {
			t.Fatalf("expected cache mode %q to be invalid", v)
		}
	}
}

func TestIsInterface(t *testing.T) {
	t.Parallel()

	valid := []string{
		"ahci",
		"ide",
		"scsi",
		"sd",
		"mtd",
		"floppy",
		"pflash",
		"virtio",
	}

	for _, v := range valid {
		if !isInterface(v) {
			t.Fatalf("expected interface %q to be valid", v)
		}
	}

	invalid := []string{
		"",
		"writeback",
		"bad",
		"e1000",
	}

	for _, v := range invalid {
		if isInterface(v) {
			t.Fatalf("expected interface %q to be invalid", v)
		}
	}
}
