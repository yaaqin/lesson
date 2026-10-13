package main

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
)

// ---------- Coding Dasar (JavaScript) ----------
//
// Tier subject 'coding' (migrations/0015), dipisah dari picker jenjang
// matematika. Beda dari tier matematika, opsi jawabannya teks (output
// console.log, nama tipe, potongan kode) -- disimpen di option_label, value-nya
// cuma indeks 0..n-1. Prompt yang ada kodenya pakai blok ``` biar userApp
// nge-render-nya monospace (lihat QuestionPrompt di apps/web).
//
// Soal campuran: sebagian ditulis manual (konsep yang gak bisa digenerate,
// mis. == vs ===), sebagian procedural (angka diacak, output dihitung di sini).
// Semua output udah dicocokin ke Node.js -- kalau nambah soal, pastiin output
// yang ditulis di `correct` emang persis yang dicetak console.log.

// codeQ: satu soal coding sebelum opsinya dipotong sesuai optionCount.
type codeQ struct {
	text    string
	code    string
	correct string
	wrong   []string
}

const askOutput = "Apa output kode berikut?"

func (c codeQ) prompt() string {
	if c.code == "" {
		return c.text
	}
	return c.text + "\n```\n" + c.code + "\n```"
}

// toQuestion: 1 benar + (optionCount-1) distraktor acak (dedup, gak boleh
// sama dengan jawaban benar), posisi diacak, value = indeks setelah diacak.
func (c codeQ) toQuestion(optionCount int) question {
	seen := map[string]bool{c.correct: true}
	wrong := make([]string, 0, len(c.wrong))
	for _, w := range c.wrong {
		if !seen[w] {
			seen[w] = true
			wrong = append(wrong, w)
		}
	}
	// Jawaban angka yang distraktornya kebetulan dobel ditambal angka tetangga.
	if n, err := strconv.Atoi(c.correct); err == nil {
		for off := 1; len(wrong) < optionCount-1; off++ {
			for _, v := range []string{itoa(n + off), itoa(n - off)} {
				if !seen[v] && len(wrong) < optionCount-1 {
					seen[v] = true
					wrong = append(wrong, v)
				}
			}
		}
	}
	rand.Shuffle(len(wrong), func(i, j int) { wrong[i], wrong[j] = wrong[j], wrong[i] })
	if len(wrong) > optionCount-1 {
		wrong = wrong[:optionCount-1]
	}

	labels := append([]string{c.correct}, wrong...)
	rand.Shuffle(len(labels), func(i, j int) { labels[i], labels[j] = labels[j], labels[i] })
	options := make([]option, len(labels))
	for i, l := range labels {
		options[i] = option{value: float64(i), label: l, isCorrect: l == c.correct}
	}
	return question{prompt: c.prompt(), options: options}
}

// codingChallenge: bank = soal manual `fixed` + soal procedural dari `gen`
// sampai total `bankSize` (prompt dobel dilewatin, procedural bisa kebetulan
// sama).
func codingChallenge(
	name string, isExam bool, required, bankSize, passPercent, optionCount, timeLimitSeconds int,
	fixed []codeQ, gen func() codeQ,
) challengeSpec {
	seen := map[string]bool{}
	questions := []question{}
	add := func(c codeQ) {
		q := c.toQuestion(optionCount)
		if seen[q.prompt] || len(questions) >= bankSize {
			return
		}
		seen[q.prompt] = true
		questions = append(questions, q)
	}
	for _, c := range fixed {
		add(c)
	}
	for tries := 0; gen != nil && len(questions) < bankSize && tries < bankSize*20; tries++ {
		add(gen())
	}
	return challengeSpec{
		name:                  name,
		isExam:                isExam,
		questionCountRequired: required,
		passThresholdPercent:  passPercent,
		optionCount:           optionCount,
		timeLimitSeconds:      timeLimitSeconds,
		questions:             questions,
	}
}

// mixOfCode: generator yang tiap dipanggil ngambil acak dari beberapa
// generator lain -- buat ujian batch.
func mixOfCode(gens ...func() codeQ) func() codeQ {
	return func() codeQ { return gens[rand.IntN(len(gens))]() }
}

// pickFixed: generator yang ngambil acak dari soal manual -- buat ngisi bank
// ujian dari soal latihan yang udah ada.
func pickFixed(pools ...[]codeQ) func() codeQ {
	all := []codeQ{}
	for _, p := range pools {
		all = append(all, p...)
	}
	return func() codeQ { return all[rand.IntN(len(all))] }
}

func itoa(n int) string { return fmt.Sprint(n) }

func codingTierSpec() tierSpec {
	return tierSpec{
		code: "coding", name: "Coding Dasar", subject: "coding", orderIndex: 10,
		batches: []batchSpec{
			{
				name: "Batch 1 — Output & Variabel",
				challenges: []challengeSpec{
					codingChallenge("console.log & String", false, 5, 10, 70, 4, 40, consoleLogQs, nil),
					codingChallenge("Variabel let & const", false, 5, 12, 70, 4, 40, variabelQs, genReassign),
					codingChallenge("Tipe Data & typeof", false, 5, 10, 70, 4, 40, typeofQs, nil),
					codingChallenge("Operator Aritmatika", false, 5, 12, 70, 4, 40, aritmatikaQs, genArithmetic),
					codingChallenge("Ujian Batch 1 Coding", true, 10, 30, 70, 4, 450, nil,
						mixOfCode(pickFixed(consoleLogQs, variabelQs, typeofQs, aritmatikaQs), genReassign, genArithmetic)),
				},
			},
			{
				name: "Batch 2 — Logika & Percabangan",
				challenges: []challengeSpec{
					codingChallenge("Perbandingan", false, 5, 12, 70, 4, 40, perbandinganQs, genComparison),
					codingChallenge("Operator Logika", false, 5, 12, 70, 4, 40, logikaQs, genLogic),
					codingChallenge("if / else", false, 5, 12, 70, 4, 45, ifElseQs, genIfElse),
					codingChallenge("else if & switch", false, 5, 12, 70, 4, 50, switchQs, genElseIfGrade),
					codingChallenge("Ujian Batch 2 Coding", true, 10, 30, 70, 4, 500, nil,
						mixOfCode(pickFixed(perbandinganQs, logikaQs, ifElseQs, switchQs), genComparison, genLogic, genIfElse, genElseIfGrade)),
				},
			},
			{
				name: "Batch 3 — Perulangan",
				challenges: []challengeSpec{
					codingChallenge("for Loop", false, 5, 12, 70, 4, 50, forQs, genForCount),
					codingChallenge("while Loop", false, 5, 12, 70, 4, 50, whileQs, genWhile),
					codingChallenge("Akumulasi dengan Loop", false, 5, 12, 70, 4, 60, nil, genLoopSum),
					codingChallenge("break & continue", false, 5, 10, 70, 4, 60, breakQs, genBreakContinue),
					codingChallenge("Ujian Batch 3 Coding", true, 10, 30, 70, 4, 600, nil,
						mixOfCode(pickFixed(forQs, whileQs, breakQs), genForCount, genWhile, genLoopSum, genBreakContinue)),
				},
			},
			{
				name: "Batch 4 — Array & Fungsi",
				challenges: []challengeSpec{
					codingChallenge("Array Dasar", false, 5, 12, 70, 4, 45, arrayQs, genArrayIndex),
					codingChallenge("Method Array", false, 5, 12, 70, 4, 50, arrayMethodQs, genArrayPushPop),
					codingChallenge("Fungsi", false, 5, 12, 70, 4, 50, fungsiQs, genFunctionCall),
					codingChallenge("Arrow Function & Scope", false, 5, 10, 70, 4, 50, arrowQs, nil),
					codingChallenge("Ujian Batch 4 Coding", true, 10, 30, 70, 4, 600, nil,
						mixOfCode(pickFixed(arrayQs, arrayMethodQs, fungsiQs, arrowQs), genArrayIndex, genArrayPushPop, genFunctionCall)),
				},
			},
		},
	}
}

// ---------- Batch 1: Output & Variabel ----------

var consoleLogQs = []codeQ{
	{askOutput, `console.log("Halo Dunia");`, "Halo Dunia", []string{`"Halo Dunia"`, "Error", "undefined"}},
	{askOutput, `console.log("Belajar" + "JS");`, "BelajarJS", []string{"Belajar JS", "Belajar+JS", "Error"}},
	{askOutput, `console.log("Belajar " + "JS");`, "Belajar JS", []string{"BelajarJS", "Belajar  JS", "Error"}},
	{askOutput, `console.log("Halo", "Budi");`, "Halo Budi", []string{"HaloBudi", "Halo,Budi", "Halo, Budi"}},
	{askOutput, `console.log("JS".length);`, "2", []string{"1", "3", "undefined"}},
	{askOutput, `console.log("kode".toUpperCase());`, "KODE", []string{"kode", "Kode", "Error"}},
	{askOutput, `console.log("Halo"[0]);`, "H", []string{"a", "Halo", "undefined"}},
	{askOutput, `console.log("ab".repeat(3));`, "ababab", []string{"ab3", "ab ab ab", "aaabbb"}},
	{"Fungsi apa yang dipakai buat nampilin teks ke console di JavaScript?", "", "console.log()", []string{"print()", "echo()", "System.out.println()"}},
	{"Penulisan string yang BENAR di JavaScript adalah…", "", `"Halo" atau 'Halo'`, []string{"Cuma \"Halo\"", "Cuma 'Halo'", "Halo tanpa tanda kutip"}},
	{askOutput, "let nama = \"Sari\";\nconsole.log(`Halo ${nama}`);", "Halo Sari", []string{"Halo ${nama}", "Halo nama", "Error"}},
}

var variabelQs = []codeQ{
	{askOutput, "let x = 5;\nx = 8;\nconsole.log(x);", "8", []string{"5", "13", "Error"}},
	{askOutput, "const pi = 3.14;\npi = 3;\nconsole.log(pi);", "TypeError (gak bisa ubah const)", []string{"3", "3.14", "undefined"}},
	{askOutput, "let a;\nconsole.log(a);", "undefined", []string{"null", "0", "Error"}},
	{askOutput, "let a = 2;\nlet b = a;\na = 10;\nconsole.log(b);", "2", []string{"10", "12", "undefined"}},
	{"Keyword yang dipakai buat variabel yang nilainya GAK boleh diubah adalah…", "", "const", []string{"let", "var", "static"}},
	{"Nama variabel mana yang VALID di JavaScript?", "", "totalNilai", []string{"2nilai", "total-nilai", "let"}},
	{askOutput, "let skor = 10;\nskor += 5;\nconsole.log(skor);", "15", []string{"10", "5", "105"}},
	{askOutput, "let n = 3;\nn++;\nconsole.log(n);", "4", []string{"3", "5", "n++"}},
}

// genReassign: variabel diubah beberapa kali pakai operator assignment.
func genReassign() codeQ {
	x := rand.IntN(10) + 1
	a := rand.IntN(9) + 2
	b := rand.IntN(4) + 2
	c := rand.IntN(5) + 1
	result := (x+a)*b - c
	code := fmt.Sprintf("let x = %d;\nx += %d;\nx *= %d;\nx -= %d;\nconsole.log(x);", x, a, b, c)
	return codeQ{askOutput, code, itoa(result), []string{
		itoa(x + a*b - c), itoa((x + a) * b), itoa(result + c*2), itoa(x), itoa(result - 1),
	}}
}

var typeofQs = []codeQ{
	{askOutput, `console.log(typeof 42);`, "number", []string{"integer", "string", "int"}},
	{askOutput, `console.log(typeof "42");`, "string", []string{"number", "text", "char"}},
	{askOutput, `console.log(typeof true);`, "boolean", []string{"bool", "string", "true"}},
	{askOutput, `console.log(typeof 3.5);`, "number", []string{"float", "double", "decimal"}},
	{askOutput, "let x;\nconsole.log(typeof x);", "undefined", []string{"null", "object", "Error"}},
	{askOutput, `console.log(typeof [1, 2, 3]);`, "object", []string{"array", "list", "number"}},
	{askOutput, `console.log(typeof null);`, "object", []string{"null", "undefined", "boolean"}},
	{askOutput, `console.log(typeof function () {});`, "function", []string{"object", "method", "undefined"}},
	{"Nilai mana yang tipenya boolean?", "", "false", []string{`"false"`, "0", "null"}},
	{askOutput, `console.log(Number("7") + 1);`, "8", []string{"71", "NaN", "Error"}},
	{askOutput, `console.log(String(5) + 5);`, "55", []string{"10", "NaN", "Error"}},
}

var aritmatikaQs = []codeQ{
	{askOutput, `console.log("5" + 3);`, "53", []string{"8", "NaN", "Error"}},
	{askOutput, `console.log("5" - 3);`, "2", []string{"53", "NaN", "Error"}},
	{askOutput, `console.log(10 / 4);`, "2.5", []string{"2", "3", "2.0"}},
	{askOutput, `console.log(2 ** 3);`, "8", []string{"6", "9", "23"}},
	{askOutput, `console.log(10 / 0);`, "Infinity", []string{"0", "Error", "NaN"}},
	{askOutput, `console.log("abc" * 2);`, "NaN", []string{"abcabc", "0", "Error"}},
}

// genArithmetic: urutan operasi, modulo, & pangkat.
func genArithmetic() codeQ {
	switch rand.IntN(3) {
	case 0:
		a, b, c := rand.IntN(10)+1, rand.IntN(8)+2, rand.IntN(8)+2
		r := a + b*c
		return codeQ{askOutput, fmt.Sprintf("console.log(%d + %d * %d);", a, b, c), itoa(r),
			[]string{itoa((a + b) * c), itoa(r + 1), itoa(r - 1), itoa(a + b + c)}}
	case 1:
		a, b := rand.IntN(40)+10, rand.IntN(6)+3
		r := a % b
		return codeQ{askOutput, fmt.Sprintf("console.log(%d %% %d);", a, b), itoa(r),
			[]string{itoa(a / b), itoa(r + 1), itoa(b - r), fmt.Sprintf("%.1f", float64(a)/float64(b))}}
	default:
		a, b, c := rand.IntN(20)+10, rand.IntN(5)+2, rand.IntN(5)+1
		r := (a - c) / b
		a = r*b + c // biar (a - c) / b selalu bulat
		return codeQ{askOutput, fmt.Sprintf("let a = %d;\nconsole.log((a - %d) / %d);", a, c, b), itoa(r),
			[]string{itoa(a - c/b), itoa(r + 1), itoa(r - 1), itoa(a - c)}}
	}
}

// ---------- Batch 2: Logika & Percabangan ----------

var perbandinganQs = []codeQ{
	{askOutput, `console.log(5 == "5");`, "true", []string{"false", "Error", "undefined"}},
	{askOutput, `console.log(5 === "5");`, "false", []string{"true", "Error", "undefined"}},
	{askOutput, `console.log(5 !== "5");`, "true", []string{"false", "Error", "NaN"}},
	{askOutput, `console.log("b" > "a");`, "true", []string{"false", "NaN", "Error"}},
	{askOutput, `console.log(null == undefined);`, "true", []string{"false", "null", "Error"}},
	{"Bedanya == dan === di JavaScript adalah…", "", "=== juga ngecek tipe data, == nggak", []string{"Gak ada bedanya", "== lebih ketat dari ===", "=== cuma buat string"}},
	{askOutput, `console.log(0 == false);`, "true", []string{"false", "0", "Error"}},
}

// genComparison: perbandingan angka biasa.
func genComparison() codeQ {
	a, b := rand.IntN(20)+1, rand.IntN(20)+1
	ops := []struct {
		op string
		ok bool
	}{
		{">", a > b}, {"<", a < b}, {">=", a >= b}, {"<=", a <= b}, {"===", a == b}, {"!==", a != b},
	}
	o := ops[rand.IntN(len(ops))]
	correct, wrong := "true", "false"
	if !o.ok {
		correct, wrong = "false", "true"
	}
	return codeQ{askOutput, fmt.Sprintf("let a = %d;\nlet b = %d;\nconsole.log(a %s b);", a, b, o.op), correct,
		[]string{wrong, "undefined", "Error"}}
}

var logikaQs = []codeQ{
	{askOutput, `console.log(true && false);`, "false", []string{"true", "undefined", "Error"}},
	{askOutput, `console.log(true || false);`, "true", []string{"false", "undefined", "Error"}},
	{askOutput, `console.log(!true);`, "false", []string{"true", "!true", "Error"}},
	{askOutput, `console.log(!!"halo");`, "true", []string{"false", "halo", "Error"}},
	{askOutput, `console.log(Boolean(0));`, "false", []string{"true", "0", "undefined"}},
	{askOutput, `console.log(Boolean(""));`, "false", []string{"true", `""`, "undefined"}},
	{"Operator && menghasilkan true kalau…", "", "Kedua sisi bernilai true", []string{"Salah satu sisi true", "Kedua sisi false", "Sisi kiri true aja"}},
}

// genLogic: kombinasi && / || dari perbandingan angka.
func genLogic() codeQ {
	x := rand.IntN(30) + 1
	lo, hi := rand.IntN(15)+1, rand.IntN(15)+15
	var expr string
	var ok bool
	if rand.IntN(2) == 0 {
		expr = fmt.Sprintf("x > %d && x < %d", lo, hi)
		ok = x > lo && x < hi
	} else {
		expr = fmt.Sprintf("x < %d || x > %d", lo, hi)
		ok = x < lo || x > hi
	}
	correct, wrong := "true", "false"
	if !ok {
		correct, wrong = "false", "true"
	}
	return codeQ{askOutput, fmt.Sprintf("let x = %d;\nconsole.log(%s);", x, expr), correct,
		[]string{wrong, "undefined", itoa(x)}}
}

var ifElseQs = []codeQ{
	{askOutput, "if (\"\") {\n  console.log(\"A\");\n} else {\n  console.log(\"B\");\n}", "B", []string{"A", "AB", "Error"}},
	{askOutput, "if (\"0\") {\n  console.log(\"A\");\n} else {\n  console.log(\"B\");\n}", "A", []string{"B", "AB", "Error"}},
	{askOutput, "let umur = 17;\nlet status = umur >= 17 ? \"dewasa\" : \"anak\";\nconsole.log(status);", "dewasa", []string{"anak", "true", "17"}},
}

// genIfElse: if/else sederhana, cabang mana yang jalan.
func genIfElse() codeQ {
	n := rand.IntN(100) + 1
	batas := []int{50, 60, 70, 75, 80}[rand.IntN(5)]
	correct, wrong := "Lulus", "Remedial"
	if n < batas {
		correct, wrong = "Remedial", "Lulus"
	}
	code := fmt.Sprintf("let nilai = %d;\nif (nilai >= %d) {\n  console.log(\"Lulus\");\n} else {\n  console.log(\"Remedial\");\n}", n, batas)
	return codeQ{askOutput, code, correct, []string{wrong, "Lulus\nRemedial", "Gak ada output"}}
}

var switchQs = []codeQ{
	{askOutput, "let hari = 2;\nswitch (hari) {\n  case 1:\n    console.log(\"Senin\");\n    break;\n  case 2:\n    console.log(\"Selasa\");\n    break;\n  default:\n    console.log(\"Lainnya\");\n}", "Selasa", []string{"Senin", "Lainnya", "Selasa\nLainnya"}},
	{askOutput, "let x = 1;\nswitch (x) {\n  case 1:\n    console.log(\"satu\");\n  case 2:\n    console.log(\"dua\");\n    break;\n}", "satu\ndua", []string{"satu", "dua", "Gak ada output"}},
	{askOutput, "let warna = \"biru\";\nswitch (warna) {\n  case \"merah\":\n    console.log(\"M\");\n    break;\n  default:\n    console.log(\"D\");\n}", "D", []string{"M", "M\nD", "Gak ada output"}},
	{"Fungsi `break` di dalam switch adalah…", "", "Berhenti biar gak lanjut ke case berikutnya", []string{"Ngulang switch dari awal", "Lompat ke default", "Bikin error"}},
}

// genElseIfGrade: else if berantai buat konversi nilai ke grade.
func genElseIfGrade() codeQ {
	n := rand.IntN(61) + 40
	grade := "D"
	switch {
	case n >= 85:
		grade = "A"
	case n >= 70:
		grade = "B"
	case n >= 55:
		grade = "C"
	}
	code := fmt.Sprintf("let nilai = %d;\nif (nilai >= 85) {\n  console.log(\"A\");\n} else if (nilai >= 70) {\n  console.log(\"B\");\n} else if (nilai >= 55) {\n  console.log(\"C\");\n} else {\n  console.log(\"D\");\n}", n)
	return codeQ{askOutput, code, grade, []string{"A", "B", "C", "D"}}
}

// ---------- Batch 3: Perulangan ----------

var forQs = []codeQ{
	{askOutput, "for (let i = 0; i < 3; i++) {\n  console.log(i);\n}", "0\n1\n2", []string{"1\n2\n3", "0\n1\n2\n3", "3"}},
	{askOutput, "for (let i = 3; i > 0; i--) {\n  console.log(i);\n}", "3\n2\n1", []string{"3\n2\n1\n0", "1\n2\n3", "2\n1\n0"}},
	{askOutput, "for (let i = 0; i < 6; i += 2) {\n  console.log(i);\n}", "0\n2\n4", []string{"0\n2\n4\n6", "2\n4\n6", "0\n1\n2"}},
	{"Bagian `i++` di `for (let i = 0; i < 5; i++)` dijalanin kapan?", "", "Tiap kali selesai satu putaran", []string{"Sekali di awal", "Sebelum ngecek kondisi pertama kali", "Cuma di putaran terakhir"}},
}

// genForCount: berapa kali isi loop dijalanin.
func genForCount() codeQ {
	start := rand.IntN(4)
	end := start + rand.IntN(8) + 3
	step := rand.IntN(2) + 1
	count := 0
	for i := start; i < end; i += step {
		count++
	}
	code := fmt.Sprintf("let hitung = 0;\nfor (let i = %d; i < %d; i += %d) {\n  hitung++;\n}\nconsole.log(hitung);", start, end, step)
	return codeQ{askOutput, code, itoa(count), []string{itoa(count + 1), itoa(count - 1), itoa(end - start), itoa(end), itoa(count + 2)}}
}

var whileQs = []codeQ{
	{askOutput, "let i = 0;\nwhile (i < 3) {\n  i++;\n}\nconsole.log(i);", "3", []string{"2", "4", "0"}},
	{askOutput, "let i = 10;\nwhile (i < 3) {\n  i++;\n}\nconsole.log(i);", "10", []string{"3", "11", "Error"}},
	{askOutput, "let i = 5;\ndo {\n  i++;\n} while (i < 3);\nconsole.log(i);", "6", []string{"5", "3", "Error"}},
	{"Apa yang terjadi kalau kondisi while gak pernah jadi false?", "", "Loop jalan terus (infinite loop)", []string{"Loop dilewatin", "Langsung error di awal", "Loop jalan sekali"}},
}

// genWhile: angka dikali/dibagi sampai lewat batas.
func genWhile() codeQ {
	if rand.IntN(2) == 0 {
		n := rand.IntN(3) + 1
		limit := rand.IntN(60) + 20
		steps := 0
		for x := n; x < limit; x *= 2 {
			steps++
		}
		code := fmt.Sprintf("let x = %d;\nlet langkah = 0;\nwhile (x < %d) {\n  x *= 2;\n  langkah++;\n}\nconsole.log(langkah);", n, limit)
		return codeQ{askOutput, code, itoa(steps), []string{itoa(steps + 1), itoa(steps - 1), itoa(limit / 2), itoa(steps + 2)}}
	}
	n := []int{64, 81, 100, 120, 200, 256}[rand.IntN(6)]
	x := n
	for x > 10 {
		x = x / 2
	}
	code := fmt.Sprintf("let x = %d;\nwhile (x > 10) {\n  x = Math.floor(x / 2);\n}\nconsole.log(x);", n)
	return codeQ{askOutput, code, itoa(x), []string{itoa(x * 2), itoa(x + 1), "10", itoa(x - 1)}}
}

// genLoopSum: akumulasi total / perkalian dalam loop.
func genLoopSum() codeQ {
	switch rand.IntN(3) {
	case 0:
		n := rand.IntN(8) + 3
		sum := n * (n + 1) / 2
		code := fmt.Sprintf("let total = 0;\nfor (let i = 1; i <= %d; i++) {\n  total += i;\n}\nconsole.log(total);", n)
		return codeQ{askOutput, code, itoa(sum), []string{itoa(sum - n), itoa(sum + n + 1), itoa(n), itoa(n * n)}}
	case 1:
		n := rand.IntN(4) + 3
		prod := 1
		for i := 1; i <= n; i++ {
			prod *= i
		}
		code := fmt.Sprintf("let hasil = 1;\nfor (let i = 1; i <= %d; i++) {\n  hasil *= i;\n}\nconsole.log(hasil);", n)
		return codeQ{askOutput, code, itoa(prod), []string{itoa(prod / n), itoa(prod * (n + 1)), itoa(n * (n + 1) / 2), "0"}}
	default:
		n := rand.IntN(10) + 6
		sum := 0
		for i := 0; i <= n; i += 2 {
			sum += i
		}
		code := fmt.Sprintf("let total = 0;\nfor (let i = 0; i <= %d; i += 2) {\n  total += i;\n}\nconsole.log(total);", n)
		return codeQ{askOutput, code, itoa(sum), []string{itoa(n * (n + 1) / 2), itoa(sum + 2), itoa(sum - n%2 - 2), itoa(sum + n)}}
	}
}

var breakQs = []codeQ{
	{askOutput, "for (let i = 1; i <= 5; i++) {\n  if (i === 3) break;\n  console.log(i);\n}", "1\n2", []string{"1\n2\n3", "1\n2\n4\n5", "3"}},
	{askOutput, "for (let i = 1; i <= 4; i++) {\n  if (i === 2) continue;\n  console.log(i);\n}", "1\n3\n4", []string{"1", "1\n2\n3\n4", "2"}},
	{"Bedanya `break` dan `continue` di dalam loop adalah…", "", "break keluar dari loop, continue lompat ke putaran berikutnya", []string{"Sama aja", "break lompat ke putaran berikutnya, continue keluar dari loop", "continue ngulang loop dari awal"}},
}

// genBreakContinue: total dengan bilangan yang di-skip / loop yang dihentiin.
func genBreakContinue() codeQ {
	n := rand.IntN(6) + 5
	k := rand.IntN(n-2) + 2
	if rand.IntN(2) == 0 {
		sum := n*(n+1)/2 - k
		code := fmt.Sprintf("let total = 0;\nfor (let i = 1; i <= %d; i++) {\n  if (i === %d) continue;\n  total += i;\n}\nconsole.log(total);", n, k)
		return codeQ{askOutput, code, itoa(sum), []string{itoa(sum + k), itoa((k - 1) * k / 2), itoa(sum - 1), itoa(k)}}
	}
	sum := (k - 1) * k / 2
	code := fmt.Sprintf("let total = 0;\nfor (let i = 1; i <= %d; i++) {\n  if (i === %d) break;\n  total += i;\n}\nconsole.log(total);", n, k)
	return codeQ{askOutput, code, itoa(sum), []string{itoa(sum + k), itoa(n*(n+1)/2 - k), itoa(n * (n + 1) / 2), itoa(k)}}
}

// ---------- Batch 4: Array & Fungsi ----------

var arrayQs = []codeQ{
	{askOutput, "let buah = [\"apel\", \"jeruk\", \"mangga\"];\nconsole.log(buah[1]);", "jeruk", []string{"apel", "mangga", "undefined"}},
	{askOutput, "let buah = [\"apel\", \"jeruk\"];\nconsole.log(buah[2]);", "undefined", []string{"jeruk", "Error", "null"}},
	{askOutput, "let a = [1, 2, 3];\nconsole.log(a.length);", "3", []string{"2", "4", "undefined"}},
	{askOutput, "let a = [1, 2, 3];\na[0] = 9;\nconsole.log(a[0] + a[2]);", "12", []string{"4", "10", "93"}},
	{"Indeks elemen PERTAMA di array JavaScript adalah…", "", "0", []string{"1", "-1", "Tergantung isi array"}},
	{askOutput, "let a = [5, 6, 7];\nconsole.log(a[a.length - 1]);", "7", []string{"6", "undefined", "3"}},
}

// genArrayIndex: akses indeks dari array angka acak.
func genArrayIndex() codeQ {
	n := rand.IntN(3) + 4
	arr := make([]int, n)
	parts := make([]string, n)
	for i := range arr {
		arr[i] = rand.IntN(50) + 1
		parts[i] = itoa(arr[i])
	}
	i, j := rand.IntN(n), rand.IntN(n)
	code := fmt.Sprintf("let nilai = [%s];\nconsole.log(nilai[%d] + nilai[%d]);", strings.Join(parts, ", "), i, j)
	r := arr[i] + arr[j]
	wrong := []string{itoa(r + 1), "undefined"}
	if i+1 < n && j+1 < n {
		wrong = append(wrong, itoa(arr[i+1]+arr[j+1]))
	}
	if i > 0 && j > 0 {
		wrong = append(wrong, itoa(arr[i-1]+arr[j-1]))
	}
	wrong = append(wrong, itoa(i+j), itoa(r-1))
	return codeQ{askOutput, code, itoa(r), wrong}
}

var arrayMethodQs = []codeQ{
	{askOutput, "let a = [1, 2];\na.push(3);\nconsole.log(a.length);", "3", []string{"2", "4", "1"}},
	{askOutput, "let a = [1, 2, 3];\nlet x = a.pop();\nconsole.log(x);", "3", []string{"1", "2", "undefined"}},
	{askOutput, "let a = [\"x\", \"y\", \"z\"];\nconsole.log(a.indexOf(\"z\"));", "2", []string{"3", "-1", "1"}},
	{askOutput, "let a = [\"x\", \"y\"];\nconsole.log(a.indexOf(\"q\"));", "-1", []string{"0", "undefined", "2"}},
	{askOutput, "let a = [1, 2, 3];\nconsole.log(a.includes(2));", "true", []string{"false", "1", "2"}},
	{askOutput, "let a = [\"a\", \"b\", \"c\"];\nconsole.log(a.join(\"-\"));", "a-b-c", []string{"abc", "a,b,c", "-a-b-c-"}},
	{askOutput, "let a = [1, 2, 3];\nlet b = a.map(x => x * 2);\nconsole.log(b.join(\",\"));", "2,4,6", []string{"1,2,3", "6", "2,4,6,8"}},
	{askOutput, "let a = [1, 2, 3, 4];\nlet b = a.filter(x => x > 2);\nconsole.log(b.length);", "2", []string{"3", "4", "1"}},
	{"Method array buat nambah elemen di AKHIR array adalah…", "", "push()", []string{"pop()", "shift()", "unshift()"}},
}

// genArrayPushPop: urutan push/pop terus cek panjang array.
func genArrayPushPop() codeQ {
	start := rand.IntN(4) + 1
	parts := make([]string, start)
	for i := range parts {
		parts[i] = itoa(i + 1)
	}
	pushes, pops := rand.IntN(3)+1, rand.IntN(2)+1
	lines := []string{fmt.Sprintf("let a = [%s];", strings.Join(parts, ", "))}
	for i := 0; i < pushes; i++ {
		lines = append(lines, fmt.Sprintf("a.push(%d);", rand.IntN(9)+1))
	}
	for i := 0; i < pops; i++ {
		lines = append(lines, "a.pop();")
	}
	lines = append(lines, "console.log(a.length);")
	r := start + pushes - pops
	return codeQ{askOutput, strings.Join(lines, "\n"), itoa(r),
		[]string{itoa(r + 1), itoa(r - 1), itoa(start + pushes), itoa(start), itoa(r + 2)}}
}

var fungsiQs = []codeQ{
	{askOutput, "function sapa(nama) {\n  return \"Halo \" + nama;\n}\nconsole.log(sapa(\"Andi\"));", "Halo Andi", []string{"Halo nama", "Andi", "undefined"}},
	{askOutput, "function f() {\n  console.log(\"A\");\n}\nlet x = f();\nconsole.log(x);", "A\nundefined", []string{"A", "undefined", "A\nA"}},
	{askOutput, "function tambah(a, b) {\n  return a + b;\n}\nconsole.log(tambah(2));", "NaN", []string{"2", "Error", "undefined"}},
	{askOutput, "function f() {\n  return 1;\n  console.log(\"X\");\n}\nconsole.log(f());", "1", []string{"X\n1", "1\nX", "X"}},
	{"Keyword buat ngembaliin nilai dari fungsi adalah…", "", "return", []string{"break", "output", "console.log"}},
}

// genFunctionCall: panggil fungsi dengan argumen acak.
func genFunctionCall() codeQ {
	a, b := rand.IntN(9)+1, rand.IntN(9)+1
	switch rand.IntN(3) {
	case 0:
		r := a*2 + b
		code := fmt.Sprintf("function hitung(x, y) {\n  return x * 2 + y;\n}\nconsole.log(hitung(%d, %d));", a, b)
		return codeQ{askOutput, code, itoa(r), []string{itoa(b*2 + a), itoa((a + b) * 2), itoa(r + 1), itoa(a + b)}}
	case 1:
		big := a
		if b > a {
			big = b
		}
		code := fmt.Sprintf("function terbesar(x, y) {\n  if (x > y) return x;\n  return y;\n}\nconsole.log(terbesar(%d, %d));", a, b)
		small := a + b - big
		return codeQ{askOutput, code, itoa(big), []string{itoa(small), itoa(a + b), "undefined", "true"}}
	default:
		r := a * a
		code := fmt.Sprintf("function kuadrat(n) {\n  return n * n;\n}\nconsole.log(kuadrat(%d) + %d);", a, b)
		return codeQ{askOutput, code, itoa(r + b), []string{itoa(r), itoa(a*2 + b), itoa((a + b) * (a + b)), itoa(r + b + 1)}}
	}
}

var arrowQs = []codeQ{
	{askOutput, "const kali = (a, b) => a * b;\nconsole.log(kali(3, 4));", "12", []string{"7", "34", "undefined"}},
	{askOutput, "const f = x => { x * 2 };\nconsole.log(f(5));", "undefined", []string{"10", "5", "Error"}},
	{askOutput, "const f = x => x + 1;\nconsole.log(f(f(1)));", "3", []string{"2", "1", "4"}},
	{askOutput, "let x = 1;\nfunction ubah() {\n  let x = 5;\n}\nubah();\nconsole.log(x);", "1", []string{"5", "6", "undefined"}},
	{askOutput, "let x = 1;\nfunction ubah() {\n  x = 5;\n}\nubah();\nconsole.log(x);", "5", []string{"1", "6", "undefined"}},
	{askOutput, "if (true) {\n  let y = 2;\n}\nconsole.log(y);", "ReferenceError", []string{"2", "undefined", "null"}},
	{"Penulisan arrow function yang BENAR adalah…", "", "const f = (a) => a * 2;", []string{"const f = (a) -> a * 2;", "const f => (a) a * 2;", "function => f(a) { a * 2 }"}},
	{askOutput, "function luar() {\n  let a = 3;\n  function dalam() {\n    return a * 2;\n  }\n  return dalam();\n}\nconsole.log(luar());", "6", []string{"3", "undefined", "ReferenceError"}},
}
