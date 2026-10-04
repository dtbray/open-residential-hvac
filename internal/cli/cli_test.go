// SPDX-License-Identifier: AGPL-3.0-only
package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestWorkflow(t *testing.T) {
	p := "../../testdata/designload/buildings/multi-room-ranch.json"
	for _, args := range [][]string{{"validate", p}, {"load", p}, {"load", p, "--room", "Bedroom"}, {"load", p, "--format", "json"}, {"explain", p, "--node", "room/bedroom/heating/envelope/bedroom-west"}} {
		var out, err bytes.Buffer
		if code := Run(args, &out, &err); code != 0 {
			t.Fatal(args, code, err.String())
		}
		if strings.Contains(strings.Join(args, " "), "--format json") && !json.Valid(out.Bytes()) {
			t.Fatal("invalid JSON output")
		}
	}
	var out, err bytes.Buffer
	if code := Run([]string{"load", p, "--room", "absent"}, &out, &err); code == 0 {
		t.Fatal("unknown room accepted")
	}
}

func TestMethodSelectionAndOutputUnits(t *testing.T) {
	p := "../../testdata/designload/buildings/wall-conduction-only.json"
	for _, units := range []string{"si", "ip"} {
		var out, errOut bytes.Buffer
		if code := Run([]string{"load", p, "--model", "designload", "--units", units}, &out, &errOut); code != 0 {
			t.Fatal(errOut.String())
		}
		expected := "Heating design load: 280.0 W"
		if units == "ip" {
			expected = "Heating design load: 955.4 Btu/h"
		}
		if !strings.Contains(out.String(), expected) || !strings.Contains(out.String(), "Open Design Load 0.1") {
			t.Fatal("incorrect units or missing methodology", out.String())
		}
	}
	var out, errOut bytes.Buffer
	if code := Run([]string{"load", p, "--model", "en12831", "--format", "json"}, &out, &errOut); code == 0 || !json.Valid(errOut.Bytes()) {
		t.Fatal("unsupported method not rejected in structured output")
	}
	out.Reset()
	errOut.Reset()
	if code := Run([]string{"explain", p, "--format", "json", "--node", "room/living/heating/envelope/living-west"}, &out, &errOut); code != 0 {
		t.Fatal(errOut.String())
	}
	var v map[string]json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if v["methodology"] == nil || v["node"] == nil {
		t.Fatal("isolated node output lost methodology context")
	}
}
