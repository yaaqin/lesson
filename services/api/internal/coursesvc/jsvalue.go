package coursesvc

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Parser literal JavaScript buat soal write_value: user nulis nilai/data
// sendiri (mis. `{ nama: "Budi", umur: 17 }`), server ngecek bentuknya pakai
// skema -- jawabannya bebas asal tipenya bener, bukan dicocokin ke satu kunci.
//
// Yang didukung cuma literal (gak ada ekspresi / pemanggilan fungsi selain
// Symbol(...)): string '..' ".." `..`, number (termasuk NaN, Infinity, desimal,
// eksponen), bigint (123n), true/false, null, undefined, array, object (key
// identifier / string / angka, trailing comma boleh). Prefix `let x =` /
// `const x =` / `var x =` & titik koma di akhir dibuang dulu.

type JSValue struct {
	Kind   string // string | number | bigint | boolean | null | undefined | symbol | array | object
	Str    string
	Num    float64
	Items  []JSValue
	Keys   []string // urutan key object sesuai tulisan user
	Fields map[string]JSValue
}

var declPrefix = regexp.MustCompile(`^\s*(?:let|const|var)\s+[A-Za-z_$][\w$]*\s*=\s*`)

// ParseJSValue: parse satu literal JS. Error-nya kalimat buat ditampilin ke
// user (bahasa Indonesia).
func ParseJSValue(src string) (JSValue, error) {
	src = declPrefix.ReplaceAllString(src, "")
	src = strings.TrimSpace(src)
	src = strings.TrimSpace(strings.TrimSuffix(src, ";"))
	if src == "" {
		return JSValue{}, fmt.Errorf("jawaban masih kosong")
	}
	p := &jsParser{src: src}
	v, err := p.value(0)
	if err != nil {
		return JSValue{}, err
	}
	p.skipSpace()
	if p.pos < len(p.src) {
		return JSValue{}, fmt.Errorf("ada tulisan lebih setelah nilai: %q", truncate(p.src[p.pos:], 20))
	}
	return v, nil
}

type jsParser struct {
	src string
	pos int
}

const maxDepth = 20

func (p *jsParser) skipSpace() {
	for p.pos < len(p.src) {
		switch p.src[p.pos] {
		case ' ', '\t', '\n', '\r':
			p.pos++
		default:
			return
		}
	}
}

func (p *jsParser) peek() byte {
	if p.pos >= len(p.src) {
		return 0
	}
	return p.src[p.pos]
}

func (p *jsParser) value(depth int) (JSValue, error) {
	if depth > maxDepth {
		return JSValue{}, fmt.Errorf("data terlalu dalam bersarangnya")
	}
	p.skipSpace()
	c := p.peek()
	switch {
	case c == 0:
		return JSValue{}, fmt.Errorf("nilainya kepotong, ada yang belum ditutup")
	case c == '"' || c == '\'' || c == '`':
		s, err := p.str()
		return JSValue{Kind: "string", Str: s}, err
	case c == '[':
		return p.array(depth)
	case c == '{':
		return p.object(depth)
	case c == '-' || c == '+' || c == '.' || (c >= '0' && c <= '9'):
		return p.number()
	case isIdentStart(c):
		return p.keyword()
	default:
		return JSValue{}, fmt.Errorf("karakter %q gak dikenal di sini", string(c))
	}
}

func (p *jsParser) str() (string, error) {
	quote := p.src[p.pos]
	p.pos++
	var b strings.Builder
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		switch {
		case c == '\\' && p.pos+1 < len(p.src):
			next := p.src[p.pos+1]
			switch next {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			default:
				b.WriteByte(next)
			}
			p.pos += 2
		case c == quote:
			p.pos++
			return b.String(), nil
		case c == '\n' && quote != '`':
			return "", fmt.Errorf("string belum ditutup pakai %s", string(quote))
		default:
			b.WriteByte(c)
			p.pos++
		}
	}
	return "", fmt.Errorf("string belum ditutup pakai %s", string(quote))
}

var numberRe = regexp.MustCompile(`^[+-]?(?:\d+\.?\d*|\.\d+)(?:[eE][+-]?\d+)?n?`)

func (p *jsParser) number() (JSValue, error) {
	rest := p.src[p.pos:]
	if strings.HasPrefix(rest, "-Infinity") || strings.HasPrefix(rest, "+Infinity") {
		p.pos += len("-Infinity")
		return JSValue{Kind: "number"}, nil
	}
	m := numberRe.FindString(rest)
	if m == "" {
		return JSValue{}, fmt.Errorf("angka %q gak valid", truncate(rest, 10))
	}
	p.pos += len(m)
	if strings.HasSuffix(m, "n") {
		digits := strings.TrimSuffix(m, "n")
		if strings.ContainsAny(digits, ".eE") {
			return JSValue{}, fmt.Errorf("BigInt (akhiran n) cuma boleh bilangan bulat")
		}
		return JSValue{Kind: "bigint", Str: digits}, nil
	}
	n, err := strconv.ParseFloat(m, 64)
	if err != nil {
		return JSValue{}, fmt.Errorf("angka %q gak valid", m)
	}
	return JSValue{Kind: "number", Num: n}, nil
}

func isIdentStart(c byte) bool {
	return c == '_' || c == '$' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isIdentPart(c byte) bool {
	return isIdentStart(c) || (c >= '0' && c <= '9')
}

func (p *jsParser) ident() string {
	start := p.pos
	for p.pos < len(p.src) && isIdentPart(p.src[p.pos]) {
		p.pos++
	}
	return p.src[start:p.pos]
}

func (p *jsParser) keyword() (JSValue, error) {
	word := p.ident()
	switch word {
	case "true", "false":
		return JSValue{Kind: "boolean", Str: word}, nil
	case "null":
		return JSValue{Kind: "null"}, nil
	case "undefined":
		return JSValue{Kind: "undefined"}, nil
	case "NaN", "Infinity":
		return JSValue{Kind: "number"}, nil
	case "Symbol":
		p.skipSpace()
		if p.peek() != '(' {
			return JSValue{}, fmt.Errorf("Symbol harus dipanggil, mis. Symbol(\"id\")")
		}
		p.pos++
		p.skipSpace()
		desc := ""
		if c := p.peek(); c == '"' || c == '\'' || c == '`' {
			s, err := p.str()
			if err != nil {
				return JSValue{}, err
			}
			desc = s
			p.skipSpace()
		}
		if p.peek() != ')' {
			return JSValue{}, fmt.Errorf("Symbol(...) belum ditutup pakai )")
		}
		p.pos++
		return JSValue{Kind: "symbol", Str: desc}, nil
	case "True", "False", "TRUE", "FALSE":
		return JSValue{}, fmt.Errorf("boolean di JavaScript huruf kecil semua: %s", strings.ToLower(word))
	case "None", "nil", "NULL", "Null":
		return JSValue{}, fmt.Errorf("%s bukan nilai JavaScript, maksudnya null?", word)
	default:
		return JSValue{}, fmt.Errorf("%s bukan nilai -- kalau maksudnya teks, bungkus pakai tanda kutip: \"%s\"", word, word)
	}
}

func (p *jsParser) array(depth int) (JSValue, error) {
	p.pos++ // [
	v := JSValue{Kind: "array", Items: []JSValue{}}
	for {
		p.skipSpace()
		if p.peek() == ']' {
			p.pos++
			return v, nil
		}
		item, err := p.value(depth + 1)
		if err != nil {
			return JSValue{}, err
		}
		v.Items = append(v.Items, item)
		p.skipSpace()
		switch p.peek() {
		case ',':
			p.pos++
		case ']':
			p.pos++
			return v, nil
		default:
			return JSValue{}, fmt.Errorf("isi array harus dipisah koma dan ditutup pakai ]")
		}
	}
}

func (p *jsParser) object(depth int) (JSValue, error) {
	p.pos++ // {
	v := JSValue{Kind: "object", Fields: map[string]JSValue{}}
	for {
		p.skipSpace()
		if p.peek() == '}' {
			p.pos++
			return v, nil
		}
		var key string
		switch c := p.peek(); {
		case c == '"' || c == '\'' || c == '`':
			s, err := p.str()
			if err != nil {
				return JSValue{}, err
			}
			key = s
		case isIdentStart(c):
			key = p.ident()
		case c >= '0' && c <= '9':
			n, err := p.number()
			if err != nil {
				return JSValue{}, err
			}
			key = strconv.FormatFloat(n.Num, 'f', -1, 64)
		default:
			return JSValue{}, fmt.Errorf("key object gak valid -- contoh yang bener: { nama: \"Budi\" }")
		}
		p.skipSpace()
		if p.peek() != ':' {
			return JSValue{}, fmt.Errorf("setelah key %q harus ada titik dua (:)", key)
		}
		p.pos++
		val, err := p.value(depth + 1)
		if err != nil {
			return JSValue{}, err
		}
		if _, dup := v.Fields[key]; !dup {
			v.Keys = append(v.Keys, key)
		}
		v.Fields[key] = val
		p.skipSpace()
		switch p.peek() {
		case ',':
			p.pos++
		case '}':
			p.pos++
			return v, nil
		default:
			return JSValue{}, fmt.Errorf("pasangan key: value harus dipisah koma dan object ditutup pakai }")
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// ---------- Skema ----------

// Schema: aturan bentuk data buat soal write_value. Type "any" = tipe apa aja.
// Type "integer" = number bulat.
type Schema struct {
	Type      string            `json:"type"`
	Required  map[string]Schema `json:"required,omitempty"` // key wajib di object + skemanya
	MinKeys   int               `json:"minKeys,omitempty"`
	Items     *Schema           `json:"items,omitempty"` // skema tiap isi array
	MinItems  int               `json:"minItems,omitempty"`
	NotEmpty  bool              `json:"notEmpty,omitempty"` // string gak boleh ""
	OneOf     []string          `json:"oneOf,omitempty"`    // string harus salah satu dari ini
	Min       *float64          `json:"min,omitempty"`
	AllowNull bool              `json:"allowNull,omitempty"`
}

// typeNameID: nama tipe di pesan error.
func typeNameID(kind string) string {
	switch kind {
	case "integer":
		return "number bulat"
	}
	return kind
}

// Validate: nil kalau v cocok sama skema, selain itu error yang nyebut lokasi
// datanya (mis. `data[1].harga harus number, bukan string`).
func (s Schema) Validate(v JSValue, path string) error {
	where := path
	if where == "" {
		where = "nilainya"
	}
	if s.AllowNull && v.Kind == "null" {
		return nil
	}
	switch s.Type {
	case "any", "":
	case "integer":
		if v.Kind != "number" || v.Num != float64(int64(v.Num)) {
			return fmt.Errorf("%s harus number bulat, bukan %s", where, describe(v))
		}
	default:
		if v.Kind != s.Type {
			return fmt.Errorf("%s harus %s, bukan %s", where, typeNameID(s.Type), describe(v))
		}
	}

	switch v.Kind {
	case "string":
		if s.NotEmpty && strings.TrimSpace(v.Str) == "" {
			return fmt.Errorf("%s gak boleh string kosong", where)
		}
		if len(s.OneOf) > 0 && !contains(s.OneOf, v.Str) {
			return fmt.Errorf("%s harus salah satu dari: %s", where, quoteAll(s.OneOf))
		}
	case "number":
		if s.Min != nil && v.Num < *s.Min {
			return fmt.Errorf("%s minimal %s", where, strconv.FormatFloat(*s.Min, 'f', -1, 64))
		}
	case "array":
		if len(v.Items) < s.MinItems {
			return fmt.Errorf("%s minimal isinya %d item, baru %d", where, s.MinItems, len(v.Items))
		}
		if s.Items != nil {
			for i, item := range v.Items {
				if err := s.Items.Validate(item, fmt.Sprintf("%s[%d]", path, i)); err != nil {
					return err
				}
			}
		}
	case "object":
		if len(v.Keys) < s.MinKeys {
			return fmt.Errorf("%s minimal punya %d property, baru %d", where, s.MinKeys, len(v.Keys))
		}
		keys := make([]string, 0, len(s.Required))
		for k := range s.Required {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			child := k
			if path != "" {
				child = path + "." + k
			}
			fv, ok := v.Fields[k]
			if !ok {
				return fmt.Errorf("property %q belum ada di %s", k, where)
			}
			if err := s.Required[k].Validate(fv, child); err != nil {
				return err
			}
		}
	}
	return nil
}

func describe(v JSValue) string {
	switch v.Kind {
	case "string":
		return fmt.Sprintf("string (%q)", truncate(v.Str, 15))
	case "number":
		return "number"
	}
	return v.Kind
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func quoteAll(list []string) string {
	q := make([]string, len(list))
	for i, s := range list {
		q[i] = strconv.Quote(s)
	}
	return strings.Join(q, ", ")
}

// CheckWriteValue: cek jawaban write_value. strictJSON = harus JSON valid
// (key pakai kutip dua, gak ada undefined / trailing comma / kutip satu).
func CheckWriteValue(answer string, schema Schema, strictJSON bool) error {
	answer = strings.TrimSpace(answer)
	if strictJSON {
		if answer == "" {
			return fmt.Errorf("jawaban masih kosong")
		}
		if !json.Valid([]byte(answer)) {
			return fmt.Errorf("belum JSON valid -- cek: key wajib pakai kutip dua, string pakai kutip dua (bukan kutip satu), gak boleh ada koma di item terakhir, gak ada undefined")
		}
	}
	v, err := ParseJSValue(answer)
	if err != nil {
		return err
	}
	return schema.Validate(v, "")
}
