package coursesvc

import "testing"

func TestParseJSValueKinds(t *testing.T) {
	cases := map[string]string{
		`"halo"`:                            "string",
		`'halo'`:                            "string",
		"`halo`":                            "string",
		`42`:                                "number",
		`-3.5`:                              "number",
		`1e3`:                               "number",
		`NaN`:                               "number",
		`-Infinity`:                         "number",
		`10n`:                               "bigint",
		`true`:                              "boolean",
		`null`:                              "null",
		`undefined`:                         "undefined",
		`Symbol("id")`:                      "symbol",
		`Symbol()`:                          "symbol",
		`[1, "a", true,]`:                   "array",
		`{ nama: "Budi", 'x': 1}`:           "object",
		`let umur = 17;`:                    "number",
		`const p = { a: [1, { b: null }] }`: "object",
	}
	for src, want := range cases {
		v, err := ParseJSValue(src)
		if err != nil {
			t.Errorf("%s: unexpected error %v", src, err)
			continue
		}
		if v.Kind != want {
			t.Errorf("%s: kind %s, want %s", src, v.Kind, want)
		}
	}
}

func TestParseJSValueErrors(t *testing.T) {
	for _, src := range []string{``, `"halo`, `Budi`, `True`, `[1, 2`, `{ nama "Budi" }`, `1.5n`, `{a: 1} x`, `None`} {
		if _, err := ParseJSValue(src); err == nil {
			t.Errorf("%q: expected error", src)
		}
	}
}

func TestSchemaValidate(t *testing.T) {
	min0 := 0.0
	produk := Schema{Type: "array", MinItems: 2, Items: &Schema{Type: "object", Required: map[string]Schema{
		"nama":  {Type: "string", NotEmpty: true},
		"harga": {Type: "number", Min: &min0},
	}}}
	ok := `[{ nama: "Kopi", harga: 15000 }, { nama: "Teh", harga: 8000, stok: 3 }]`
	if err := CheckWriteValue(ok, produk, false); err != nil {
		t.Fatalf("valid answer rejected: %v", err)
	}
	bad := map[string]string{
		`[{ nama: "Kopi", harga: 15000 }]`:                              "minimal",
		`[{ nama: "Kopi", harga: "15000" }, { nama: "Teh", harga: 1 }]`: "harga",
		`[{ nama: "Kopi" }, { nama: "Teh", harga: 1 }]`:                 "harga",
	}
	for src := range bad {
		if err := CheckWriteValue(src, produk, false); err == nil {
			t.Errorf("%s: expected schema error", src)
		}
	}

	resp := Schema{Type: "object", Required: map[string]Schema{
		"status": {Type: "string", OneOf: []string{"success", "error"}},
		"data":   {Type: "any"},
	}}
	if err := CheckWriteValue(`{"status": "success", "data": {"id": 1}}`, resp, true); err != nil {
		t.Fatalf("valid JSON rejected: %v", err)
	}
	for _, src := range []string{`{status: "success", data: 1}`, `{"status": 'success', "data": 1}`, `{"status": "success", "data": 1,}`, `{"status": "ok", "data": 1}`} {
		if err := CheckWriteValue(src, resp, true); err == nil {
			t.Errorf("%s: expected strict JSON error", src)
		}
	}

	if err := CheckWriteValue(`3.5`, Schema{Type: "integer"}, false); err == nil {
		t.Error("3.5 should not pass integer")
	}
	if err := CheckWriteValue(`null`, Schema{Type: "string", AllowNull: true}, false); err != nil {
		t.Errorf("allowNull: %v", err)
	}
}

func TestNormalizeTypeAnswer(t *testing.T) {
	for in, want := range map[string]string{` "String" `: "string", "`number`;": "number", "Boolean": "boolean", "tipe: array": "array"} {
		if got := normalizeTypeAnswer(in); got != want {
			t.Errorf("%q -> %q, want %q", in, got, want)
		}
	}
}
