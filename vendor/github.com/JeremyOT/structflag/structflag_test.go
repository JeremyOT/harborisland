package structflag

import (
	"flag"
	"testing"
	"time"
)

type TestStruct struct {
	String   string        `json:"string_with_underscores"`
	Int      int           `json:"number"`
	Bool     bool          `json:"yes_no"`
	Duration time.Duration `flag:"interval,Some description,5s"`
}

func TestStructToArgs(t *testing.T) {
	args := StructToArgs("", &TestStruct{String: "some \"string\" with spaces", Int: 42, Bool: true, Duration: time.Minute})
	if len(args) != 4 || args[0] != "-string-with-underscores=\"some \\\"string\\\" with spaces\"" || args[1] != "-number=42" || args[2] != "-yes-no=true" || args[3] != "-interval=1m0s" {
		t.Error("Unexpected args:", args)
	}
	args = StructToArgs("test", &TestStruct{String: "some \"string\" with spaces", Int: 42, Bool: true, Duration: time.Minute})
	if len(args) != 4 || args[0] != "-test-string-with-underscores=\"some \\\"string\\\" with spaces\"" || args[1] != "-test-number=42" || args[2] != "-test-yes-no=true" || args[3] != "-test-interval=1m0s" {
		t.Error("Unexpected args:", args)
	}
}

func TestStructToFlags(t *testing.T) {
	var testStruct TestStruct
	var testStruct2 TestStruct
	StructToFlags("", &testStruct)
	StructToFlags("test", &testStruct2)
	flag.Parse()
	if testStruct.Duration != 5*time.Second {
		t.Error("Failed to parse without prefix:", testStruct)
	}
	if testStruct2.Duration != 5*time.Second {
		t.Error("Failed to parse with prefix:", testStruct2)
	}
}
