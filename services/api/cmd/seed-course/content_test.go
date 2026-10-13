package main

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"lesson/api/internal/coursesvc"
)

// Ngecek konten course sebelum di-seed: kunci & contoh jawaban konsisten sama
// skemanya, bank soal cukup, dan (kalau ada Node.js) semua output di materi +
// soal "Apa output kode berikut?" + soal tebak tipe dicocokin ke hasil asli.

func TestCourseContentConsistency(t *testing.T) {
	for _, m := range modules() {
		bank := len(m.examOnly)
		for _, l := range m.levels {
			bank += len(l.questions)
			if len(l.questions) < l.questionCount {
				t.Errorf("%s/%s: bank %d < %d soal per latihan", m.code, l.code, len(l.questions), l.questionCount)
			}
			if len(l.material) == 0 {
				t.Errorf("%s/%s: materi kosong", m.code, l.code)
			}
		}
		if len(m.levels) > 0 && bank < m.examQuestionCount*2 {
			t.Errorf("%s: bank final test %d, minimal 2x %d biar soalnya beragam", m.code, bank, m.examQuestionCount)
		}
		for _, q := range allQuestions(m) {
			p := q.payload
			switch q.kind {
			case coursesvc.KindMultipleChoice:
				seen := map[string]bool{}
				for _, o := range p.Options {
					if seen[o] {
						t.Errorf("%q: opsi dobel %q", q.prompt, o)
					}
					seen[o] = true
				}
				if len(p.Options) < 3 {
					t.Errorf("%q: opsi kurang", q.prompt)
				}
			case coursesvc.KindTypeAnswer:
				if len(p.Accepted) == 0 {
					t.Errorf("%q: accepted kosong", q.code)
				}
			case coursesvc.KindWriteValue:
				if p.Schema == nil {
					t.Fatalf("%q: schema kosong", q.prompt)
				}
				if err := coursesvc.CheckWriteValue(p.Sample, *p.Schema, p.StrictJSON); err != nil {
					t.Errorf("%q: contoh jawaban gak lolos skemanya sendiri: %v", q.prompt, err)
				}
			default:
				t.Errorf("%q: kind %q gak dikenal", q.prompt, q.kind)
			}
		}
	}
}

func allQuestions(m moduleSpec) []questionSpec {
	qs := append([]questionSpec(nil), m.examOnly...)
	for _, l := range m.levels {
		qs = append(qs, l.questions...)
	}
	return qs
}

type nodeCase struct {
	Label  string   `json:"label"`
	Code   string   `json:"code"`
	Mode   string   `json:"mode"` // "output" | "type"
	Expect string   `json:"-"`
	Accept []string `json:"-"`
}

const nodeRunner = `
const util = require("util");
const cases = JSON.parse(require("fs").readFileSync(0, "utf8"));
const out = cases.map((c) => {
  if (c.mode === "type") {
    const v = new Function("const nama = 'x'; return (" + c.code + ");")();
    if (v === null) return "null";
    if (Array.isArray(v)) return "array";
    return typeof v;
  }
  const logs = [];
  const fakeConsole = { log: (...a) => logs.push(util.format(...a)) };
  try {
    new Function("console", c.code)(fakeConsole);
  } catch (e) {
    logs.push(e.name + ": " + e.message);
  }
  return logs.join("\n");
});
process.stdout.write(JSON.stringify(out));
`

func TestCourseOutputsMatchNode(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node gak ada, cek output dilewatin")
	}
	cases := []nodeCase{}
	for _, m := range modules() {
		for _, l := range m.levels {
			for _, b := range l.material {
				if b.Output != "" {
					cases = append(cases, nodeCase{Label: l.code + " materi: " + firstLine(b.Code), Code: b.Code, Mode: "output", Expect: b.Output})
				}
			}
		}
		for _, q := range allQuestions(m) {
			switch {
			case q.kind == coursesvc.KindMultipleChoice && q.prompt == askOutput:
				cases = append(cases, nodeCase{Label: firstLine(q.code), Code: q.code, Mode: "output", Expect: q.payload.Options[q.payload.Correct]})
			case q.kind == coursesvc.KindTypeAnswer:
				cases = append(cases, nodeCase{Label: q.code, Code: q.code, Mode: "type", Accept: q.payload.Accepted})
			}
		}
	}

	input, _ := json.Marshal(cases)
	cmd := exec.Command("node", "-e", nodeRunner)
	cmd.Stdin = strings.NewReader(string(input))
	raw, err := cmd.Output()
	if err != nil {
		t.Fatalf("node: %v", err)
	}
	var got []string
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("parse output node: %v", err)
	}
	for i, c := range cases {
		switch c.Mode {
		case "output":
			if got[i] != c.Expect {
				t.Errorf("%s\n  ditulis: %q\n  node:    %q", c.Label, c.Expect, got[i])
			}
		case "type":
			if !contains(c.Accept, got[i]) {
				t.Errorf("tipe %s: accepted %v, node bilang %q", c.Label, c.Accept, got[i])
			}
		}
	}
	t.Logf("%d output/tipe dicocokin ke node", len(cases))
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
