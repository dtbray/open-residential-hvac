// SPDX-License-Identifier: AGPL-3.0-only
package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestWorkflow(t *testing.T) {
	p := "../../testdata/buildings/multi-room-ranch.json"
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
