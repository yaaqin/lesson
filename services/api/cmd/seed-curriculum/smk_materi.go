package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"sort"
	"strings"
)

// ---------- SMK/SMA - per materi (batch 2 & 3) + campuran semua materi (batch 4) ----------
//
// Batch 2 = materi Kelas X–XI, batch 3 = materi Kelas XII + matematika SMK.
// Satu level = satu materi (materi yang deket digabung biar pas 10 level), ujian
// batch = campuran materi batch itu. Batch 4 = campuran dari 20 materi di atas.
// Semua level: bank 35 soal, wajib 20, opsi 5. Ujian: bank 50, wajib 30.
//
// Jawaban soal pilihan ganda disimpan NUMERIC, jadi tiap soal dirancang supaya
// jawabannya satu angka (bulat, atau desimal pas maksimal 2 angka di belakang
// koma) -- angka-angka soalnya dipilih biar hasil akhirnya "bersih".

type materiSpec struct {
	name string
	gen  func(optionCount int) question
}

const (
	smkLevelRequired     = 20
	smkLevelBank         = 35
	smkLevelTimeSeconds  = 60 // per soal
	smkExamRequired      = 30
	smkExamBank          = 50
	smkExamTimeSeconds   = 1800 // total sesi ujian
	smkOptionCount       = 5
	smkPassThresholdPcnt = 70
)

func smkMateriKelasXXI() []materiSpec {
	return []materiSpec{
		{"Eksponen & Logaritma", mixOf(genExponentRule, genExponentEquation, genLogCombo, genLogQuotient)},
		{"Barisan & Deret", mixOf(genArithmeticTerm, genArithmeticSum, genGeometricTerm, genGeometricInfinite)},
		{"Sistem Persamaan & Pertidaksamaan", mixOf(genLinearSystemHard, genLinearInequalityMax, genQuadraticInequalityCount)},
		{"Fungsi & Polinomial", mixOf(genFunctionComposition, genFunctionInverse, genQuadraticExtreme, genQuadraticSumProduct, genPolynomialRemainder)},
		{"Trigonometri", mixOf(genTrigSideLength, genTrigIdentity, genTrigCombo, genCosineRule, genTriangleAreaSine)},
		{"Vektor", mixOf(genVectorDot, genVectorMagnitude, genVectorLinearCombo, genVectorFromPoints)},
		{"Statistika & Peluang", mixOf(genStatisticsMean, genStatsMedianHard, genVariance, genCombinationCount, genArrangementWord, genExpectedFrequency, genSimpleProbability)},
		{"Program Linear", genLinearProgramMax},
		{"Matriks", mixOf(genMatrixDet2, genMatrixDet3, genMatrixProductEntry, genMatrixInverseEntry)},
		{"Lingkaran & Transformasi Geometri", mixOf(genCircleRadius, genCircleCenter, genCircleTangentGradient, genTransformation)},
	}
}

func smkMateriKelasXIIDanSMK() []materiSpec {
	return []materiSpec{
		{"Limit Fungsi", mixOf(genLimitFactor, genLimitInfinity, genLimitTrig, genLimitSubstitution)},
		{"Turunan", mixOf(genDerivativeAtPoint, genTangentSlope, genStationaryPoint, genVelocity)},
		{"Integral", mixOf(genDefiniteIntegral, genIndefiniteCoeff, genAreaUnderCurve, genAreaBetweenCurves)},
		{"Statistika Lanjut & Terapan", mixOf(genFrequencyMean, genInterquartileRange, genStdDeviation, genRegressionPredict, genNewMemberAverage)},
		{"Dimensi Tiga", mixOf(genCuboidSpaceDiagonal, genCuboidFaceDiagonal, genCuboidPointToLine, genCubeAngle)},
		{"Irisan Kerucut", mixOf(genParabola, genEllipse, genHyperbola)},
		{"Matematika Keuangan", mixOf(genSimpleInterest, genCompoundInterest, genAnnuity)},
		{"Matriks & Program Linear Terapan", mixOf(genMatrixSalesRevenue, genMatrixPriceSystem, genLinearProgramProduction)},
		{"Trigonometri Terapan", mixOf(genSlopeHeight, genElevation45, genShipDistance, genLadder)},
		{"Logika Matematika", genTruthTableCount},
	}
}

func smkMateriChallenges(materi []materiSpec, examName string) []challengeSpec {
	challenges := make([]challengeSpec, 0, len(materi)+1)
	gens := make([]func(int) question, 0, len(materi))
	for _, m := range materi {
		challenges = append(challenges, buildChallenge(m.name, false, smkLevelRequired, smkLevelBank, smkPassThresholdPcnt, smkOptionCount, smkLevelTimeSeconds, m.gen))
		gens = append(gens, m.gen)
	}
	challenges = append(challenges, buildChallenge(examName, true, smkExamRequired, smkExamBank, smkPassThresholdPcnt, smkOptionCount, smkExamTimeSeconds, mixOf(gens...)))
	return challenges
}

func smkCampuranSemuaMateriChallenges() []challengeSpec {
	var gens []func(int) question
	for _, m := range append(smkMateriKelasXXI(), smkMateriKelasXIIDanSMK()...) {
		gens = append(gens, m.gen)
	}
	all := mixOf(gens...)

	challenges := make([]challengeSpec, 0, 11)
	for level := 1; level <= 10; level++ {
		name := fmt.Sprintf("Campuran Level %d", level)
		challenges = append(challenges, buildChallenge(name, false, smkLevelRequired, smkLevelBank, smkPassThresholdPcnt, smkOptionCount, smkLevelTimeSeconds, all))
	}
	challenges = append(challenges, buildChallenge("Ujian Campuran Semua Materi SMK/SMA", true, smkExamRequired, smkExamBank, smkPassThresholdPcnt, smkOptionCount, smkExamTimeSeconds, all))
	return challenges
}

// ---------- helper ----------

// mixOf: generator yang tiap dipanggil milih salah satu sub-generator secara acak.
func mixOf(gens ...func(int) question) func(int) question {
	return func(optionCount int) question {
		return gens[rand.IntN(len(gens))](optionCount)
	}
}

func between(lo, hi int) int { return rand.IntN(hi-lo+1) + lo }

func nonZeroBetween(lo, hi int) int {
	for {
		if v := between(lo, hi); v != 0 {
			return v
		}
	}
}

func pickOne[T any](xs []T) T { return xs[rand.IntN(len(xs))] }

func randSign() int {
	if rand.IntN(2) == 0 {
		return -1
	}
	return 1
}

func intPow(base, exp int) int {
	result := 1
	for range exp {
		result *= base
	}
	return result
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

func factorial(n int) int {
	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return result
}

func combination(n, r int) int {
	if r < 0 || r > n {
		return 0
	}
	return factorial(n) / (factorial(r) * factorial(n-r))
}

func fmtNum(v float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", v), "0"), ".")
}

var superscript = map[int]string{2: "²", 3: "³"}

// formatPolyVar: polinomial dari koefisien pangkat tertinggi -> "2x³ - x + 4".
func formatPolyVar(v string, coeffs ...int) string {
	deg := len(coeffs) - 1
	var b strings.Builder
	for i, c := range coeffs {
		if c == 0 {
			continue
		}
		p := deg - i
		switch {
		case b.Len() == 0 && c < 0:
			b.WriteString("-")
		case b.Len() > 0 && c < 0:
			b.WriteString(" - ")
		case b.Len() > 0:
			b.WriteString(" + ")
		}
		if absInt(c) != 1 || p == 0 {
			fmt.Fprintf(&b, "%d", absInt(c))
		}
		if p >= 1 {
			b.WriteString(v)
		}
		if p >= 2 {
			b.WriteString(superscript[p])
		}
	}
	if b.Len() == 0 {
		return "0"
	}
	return b.String()
}

func formatPoly(coeffs ...int) string { return formatPolyVar("x", coeffs...) }

// linTerm: suku lanjutan " + 3x" / " - x" / "" (buat nyambung ke ekspresi).
func linTerm(coef int, v string) string {
	if coef == 0 {
		return ""
	}
	sign := "+"
	if coef < 0 {
		sign = "-"
	}
	if absInt(coef) == 1 && v != "" {
		return fmt.Sprintf(" %s %s", sign, v)
	}
	return fmt.Sprintf(" %s %d%s", sign, absInt(coef), v)
}

// formatLinearXY: "x + 2y" (koefisien positif).
func formatLinearXY(c, d int) string {
	x := "x"
	if c != 1 {
		x = fmt.Sprintf("%dx", c)
	}
	return x + linTerm(d, "y")
}

func joinInts(xs []int) string {
	strs := make([]string, len(xs))
	for i, x := range xs {
		strs[i] = fmt.Sprintf("%d", x)
	}
	return strings.Join(strs, ", ")
}

func formatMatrix(rows [][]int) string {
	parts := make([]string, len(rows))
	for i, row := range rows {
		cells := make([]string, len(row))
		for j, v := range row {
			cells[j] = fmt.Sprintf("%d", v)
		}
		parts[i] = strings.Join(cells, " ")
	}
	return "[" + strings.Join(parts, "; ") + "]"
}

type pyTriple struct{ a, b, c int } // a² + b² = c²

var pyTriples = []pyTriple{{3, 4, 5}, {5, 12, 13}, {8, 15, 17}, {7, 24, 25}}

type pyQuadruple struct{ a, b, c, d int } // a² + b² + c² = d²

var pyQuadruples = []pyQuadruple{
	{1, 2, 2, 3}, {2, 3, 6, 7}, {1, 4, 8, 9}, {4, 4, 7, 9}, {2, 6, 9, 11},
	{6, 6, 7, 11}, {3, 4, 12, 13}, {2, 10, 11, 15}, {8, 9, 12, 17},
}

// Segitiga sisi bulat dengan sudut C = 60° (c² = a² + b² - ab) / 120° (c² = a² + b² + ab).
type angleTriangle struct{ a, b, c, angle int }

var angleTriangles = []angleTriangle{
	{3, 8, 7, 60}, {5, 8, 7, 60}, {7, 15, 13, 60}, {8, 15, 13, 60}, {5, 21, 19, 60}, {16, 21, 19, 60},
	{3, 5, 7, 120}, {7, 8, 13, 120}, {5, 16, 19, 120},
}

// randomDeviations: n simpangan bulat dari rata-rata (jumlahnya 0), diulang
// sampai jumlah kuadratnya lolos `accept` (dipake biar ragam/simpangan baku bulat).
func randomDeviations(n, spread int, accept func(sumSq int) bool) []int {
	for {
		devs := make([]int, n)
		sum := 0
		for i := range n - 1 {
			devs[i] = between(-spread, spread)
			sum += devs[i]
		}
		devs[n-1] = -sum
		if absInt(devs[n-1]) > spread {
			continue
		}
		sumSq := 0
		for _, d := range devs {
			sumSq += d * d
		}
		if sumSq > 0 && accept(sumSq) {
			rand.Shuffle(n, func(i, j int) { devs[i], devs[j] = devs[j], devs[i] })
			return devs
		}
	}
}

// ---------- Kelas X–XI: Eksponen & Logaritma ----------

func genExponentRule(optionCount int) question {
	base := pickOne([]int{2, 3, 5, 7})
	m, n := between(3, 9), between(2, 8)
	p := between(1, m+n-1)
	correct := float64(m + n - p)
	return question{
		prompt:  fmt.Sprintf("Jika %d^%d × %d^%d ÷ %d^%d = %d^k, nilai k = ?", base, m, base, n, base, p, base),
		options: buildOptions(correct, []float64{float64(m*n - p), float64(m + n + p), correct - 1, correct + 1, correct + 2}, optionCount, false),
	}
}

func genExponentEquation(optionCount int) question {
	base := pickOne([]int{2, 3, 5})
	maxPow := map[int]int{2: 10, 3: 6, 5: 4}[base]
	n := between(2, maxPow)
	c := between(1, n-1)
	correct := float64(n - c)
	return question{
		prompt:  fmt.Sprintf("Jika %d^(x + %d) = %d, nilai x = ?", base, c, intPow(base, n)),
		options: buildOptions(correct, []float64{float64(n), float64(n + c), correct - 1, correct + 1, correct + 2}, optionCount, true),
	}
}

func genLogQuotient(optionCount int) question {
	base := pickOne([]int{2, 3, 5})
	k := between(1, 4)
	m := between(2, 12)
	correct := float64(k)
	return question{
		prompt:  fmt.Sprintf("Log basis %d dari %d - Log basis %d dari %d = ?", base, intPow(base, k)*m, base, m),
		options: buildOptions(correct, []float64{correct + 1, correct - 1, correct + 2, float64(m), float64(k * m)}, optionCount, false),
	}
}

// ---------- Kelas X–XI: Barisan & Deret ----------

func genArithmeticTerm(optionCount int) question {
	a := between(-10, 20)
	d := nonZeroBetween(-6, 9)
	n := between(10, 40)
	correct := float64(a + (n-1)*d)
	return question{
		prompt:  fmt.Sprintf("Suku ke-%d dari barisan aritmetika %d, %d, %d, %d, ... adalah?", n, a, a+d, a+2*d, a+3*d),
		options: buildOptions(correct, []float64{float64(a + n*d), float64(a + (n-2)*d), float64(n * d), correct + 1, correct - 1}, optionCount, true),
	}
}

func genArithmeticSum(optionCount int) question {
	a := between(1, 20)
	d := between(2, 8)
	n := between(8, 30)
	un := a + (n-1)*d
	correct := float64(n * (a + un) / 2)
	return question{
		prompt:  fmt.Sprintf("Jumlah %d suku pertama deret aritmetika %d + %d + %d + ... adalah?", n, a, a+d, a+2*d),
		options: buildOptions(correct, []float64{float64(un), 2 * correct, correct + float64(d), correct - float64(a), correct + float64(un)}, optionCount, false),
	}
}

func genGeometricTerm(optionCount int) question {
	r := pickOne([]int{2, 3, -2})
	a := between(1, 5)
	maxN := 9
	if r == 3 {
		maxN = 7
	}
	n := between(4, maxN)
	correct := float64(a * intPow(r, n-1))
	return question{
		prompt:  fmt.Sprintf("Suku ke-%d dari barisan geometri %d, %d, %d, ... adalah?", n, a, a*r, a*r*r),
		options: buildOptions(correct, []float64{float64(a * intPow(r, n)), float64(a * intPow(r, n-2)), -correct, float64(a * (n - 1) * r), correct + float64(a)}, optionCount, true),
	}
}

func genGeometricInfinite(optionCount int) question {
	// rasio 1/k, suku pertama t(k-1)k² -> 3 suku pertama bulat & S∞ = t·k³ bulat.
	k := between(2, 5)
	t := between(1, map[int]int{2: 10, 3: 5, 4: 3, 5: 2}[k])
	a := t * (k - 1) * k * k
	correct := float64(t * k * k * k)
	partial := a + a/k + a/(k*k)
	return question{
		prompt:  fmt.Sprintf("Jumlah deret geometri tak hingga %d + %d + %d + ... adalah?", a, a/k, a/(k*k)),
		options: buildOptions(correct, []float64{float64(partial), float64(2 * a), correct + float64(a/k), correct - float64(a/(k*k)), correct + 1}, optionCount, false),
	}
}

// ---------- Kelas X–XI: Sistem Persamaan & Pertidaksamaan ----------

func genLinearInequalityMax(optionCount int) question {
	a := between(2, 9)
	b := between(-20, 20)
	c := between(-10, 60)
	// ax + b < c  <=>  ax <= c - b - 1  <=>  x <= floor((c-b-1)/a)
	correct := float64(floorDiv(c-b-1, a))
	return question{
		prompt:  fmt.Sprintf("Bilangan bulat terbesar x yang memenuhi %s < %d adalah?", formatPoly(a, b), c),
		options: buildOptions(correct, []float64{correct + 1, correct - 1, correct + 2, float64(c - b)}, optionCount, true),
	}
}

func genQuadraticInequalityCount(optionCount int) question {
	p := between(-6, 4)
	q := p + between(3, 9)
	poly := formatPoly(1, -(p + q), p*q)
	strict := q - p - 1
	inclusive := q - p + 1
	if rand.IntN(2) == 0 {
		return question{
			prompt:  fmt.Sprintf("Banyak bilangan bulat x yang memenuhi %s < 0 adalah?", poly),
			options: buildOptions(float64(strict), []float64{float64(inclusive), float64(q - p), float64(strict - 1), float64(strict + 2)}, optionCount, false),
		}
	}
	return question{
		prompt:  fmt.Sprintf("Banyak bilangan bulat x yang memenuhi %s ≤ 0 adalah?", poly),
		options: buildOptions(float64(inclusive), []float64{float64(strict), float64(q - p), float64(inclusive + 1), float64(inclusive - 3)}, optionCount, false),
	}
}

// ---------- Kelas X–XI: Fungsi & Polinomial ----------

func genFunctionComposition(optionCount int) question {
	a, b := between(2, 6), between(-9, 9)
	c, d := between(2, 6), between(-9, 9)
	k := between(-3, 5)
	fk, gk := a*k+b, c*k+d
	correct := float64(a*gk + b)
	return question{
		prompt:  fmt.Sprintf("Diketahui f(x) = %s dan g(x) = %s. Nilai (f ∘ g)(%d) = ?", formatPoly(a, b), formatPoly(c, d), k),
		options: buildOptions(correct, []float64{float64(c*fk + d), float64(fk * gk), float64(fk + gk), correct + 1, correct - 1}, optionCount, true),
	}
}

func genFunctionInverse(optionCount int) question {
	a := between(2, 9)
	b := nonZeroBetween(-15, 15)
	t := between(-8, 12)
	v := a*t + b
	correct := float64(t)
	return question{
		prompt:  fmt.Sprintf("Diketahui f(x) = %s. Nilai f⁻¹(%d) = ?", formatPoly(a, b), v),
		options: buildOptions(correct, []float64{float64(a*v + b), round2(float64(v+b) / float64(a)), correct + 1, correct - 1, -correct}, optionCount, true),
	}
}

func genQuadraticExtreme(optionCount int) question {
	a := pickOne([]int{1, 2, 3, -1, -2, -3})
	h := between(-5, 5)
	k := between(-20, 20)
	poly := formatPoly(a, -2*a*h, a*h*h+k)
	if rand.IntN(2) == 0 {
		word := "minimum"
		if a < 0 {
			word = "maksimum"
		}
		correct := float64(k)
		return question{
			prompt:  fmt.Sprintf("Nilai %s dari fungsi f(x) = %s adalah?", word, poly),
			options: buildOptions(correct, []float64{float64(a*h*h + k), -correct, float64(h), correct + float64(a), correct - 1}, optionCount, true),
		}
	}
	correct := float64(h)
	return question{
		prompt:  fmt.Sprintf("Persamaan sumbu simetri grafik f(x) = %s adalah x = ?", poly),
		options: buildOptions(correct, []float64{-correct, correct + 1, correct - 1, 2 * correct, float64(k)}, optionCount, true),
	}
}

func genPolynomialRemainder(optionCount int) question {
	a, b := between(1, 3), between(-6, 6)
	c, d := between(-9, 9), between(-12, 12)
	k := nonZeroBetween(-3, 3)
	eval := func(x int) int { return a*x*x*x + b*x*x + c*x + d }
	correct := float64(eval(k))
	return question{
		prompt:  fmt.Sprintf("Sisa pembagian P(x) = %s oleh (%s) adalah?", formatPoly(a, b, c, d), formatPoly(1, -k)),
		options: buildOptions(correct, []float64{float64(eval(-k)), float64(d), correct + 1, correct - 1}, optionCount, true),
	}
}

// ---------- Kelas X–XI: Trigonometri ----------

func genTrigSideLength(optionCount int) question {
	t := pickOne(pyTriples)
	k := between(1, 5)
	sides := []float64{float64(t.a * k), float64(t.b * k), float64(t.c * k)}
	var prompt string
	var correct float64
	switch rand.IntN(3) {
	case 0:
		correct = float64(t.a * k)
		prompt = fmt.Sprintf("Pada segitiga siku-siku, sin A = %d/%d dan panjang sisi miringnya %d cm. Panjang sisi di depan sudut A = ? (cm)", t.a, t.c, t.c*k)
	case 1:
		correct = float64(t.b * k)
		prompt = fmt.Sprintf("Pada segitiga siku-siku, cos A = %d/%d dan panjang sisi miringnya %d cm. Panjang sisi di samping sudut A (selain sisi miring) = ? (cm)", t.b, t.c, t.c*k)
	default:
		correct = float64(t.a * k)
		prompt = fmt.Sprintf("Pada segitiga siku-siku, tan A = %d/%d dan panjang sisi di samping sudut A adalah %d cm. Panjang sisi di depan sudut A = ? (cm)", t.a, t.b, t.b*k)
	}
	return question{
		prompt:  prompt,
		options: buildOptions(correct, append(sides, correct+float64(k), correct-float64(k)), optionCount, false),
	}
}

func genTrigIdentity(optionCount int) question {
	// cuma tripel yang rasionya desimal pas: 3-4-5 (0,6/0,8) & 7-24-25 (0,28/0,96)
	t := pickOne([]pyTriple{{3, 4, 5}, {7, 24, 25}})
	sin := float64(t.a) / float64(t.c)
	cos := float64(t.b) / float64(t.c)
	if rand.IntN(2) == 0 {
		return question{
			prompt:  fmt.Sprintf("Jika sin A = %d/%d dan A sudut lancip, nilai cos A = ?", t.a, t.c),
			options: buildOptions(round2(cos), []float64{round2(sin), round2(1 - sin), round2(cos + 0.1), round2(cos - 0.2), 1}, optionCount, false),
		}
	}
	return question{
		prompt:  fmt.Sprintf("Jika cos A = %d/%d dan A sudut lancip, nilai sin A = ?", t.b, t.c),
		options: buildOptions(round2(sin), []float64{round2(cos), round2(1 - cos), round2(sin + 0.1), round2(sin + 0.2), 1}, optionCount, false),
	}
}

func genCosineRule(optionCount int) question {
	t := pickOne(angleTriangles)
	k := between(1, 4)
	a, b := t.a*k, t.b*k
	if rand.IntN(2) == 0 {
		a, b = b, a
	}
	correct := float64(t.c * k)
	otherAngle := math.Sqrt(float64(a*a + b*b + a*b))
	if t.angle == 120 {
		otherAngle = math.Sqrt(float64(a*a + b*b - a*b))
	}
	return question{
		prompt:  fmt.Sprintf("Pada segitiga ABC, panjang sisi a = %d cm, b = %d cm, dan besar sudut C = %d°. Panjang sisi c = ? (cm)", a, b, t.angle),
		options: buildOptions(correct, []float64{float64(a + b), round2(otherAngle), round2(math.Sqrt(float64(a*a + b*b))), correct + float64(k), correct - float64(k)}, optionCount, false),
	}
}

func genTriangleAreaSine(optionCount int) question {
	angle := pickOne([]int{30, 90, 150})
	sin := map[int]float64{30: 0.5, 90: 1, 150: 0.5}[angle]
	a, b := 2*between(2, 10), 2*between(2, 10)
	correct := float64(a*b) * sin / 2
	return question{
		prompt:  fmt.Sprintf("Luas segitiga ABC dengan panjang sisi a = %d cm, b = %d cm, dan besar sudut C = %d° adalah? (cm²)", a, b, angle),
		options: buildOptions(correct, []float64{float64(a*b) * sin, float64(a*b) / 2, float64(a * b), correct + 2, correct - 2}, optionCount, false),
	}
}

// ---------- Kelas X–XI: Vektor ----------

func genVectorDot(optionCount int) question {
	dim := between(2, 3)
	u, v := make([]int, dim), make([]int, dim)
	dot, sum := 0, 0
	for i := range dim {
		u[i], v[i] = between(-6, 9), between(-6, 9)
		dot += u[i] * v[i]
		sum += u[i] + v[i]
	}
	correct := float64(dot)
	return question{
		prompt:  fmt.Sprintf("Diketahui vektor u = (%s) dan v = (%s). Nilai u · v = ?", joinInts(u), joinInts(v)),
		options: buildOptions(correct, []float64{-correct, correct + 1, correct - 1, float64(sum)}, optionCount, true),
	}
}

func genVectorMagnitude(optionCount int) question {
	var comps []int
	var length int
	if rand.IntN(2) == 0 {
		t := pickOne(pyTriples)
		comps, length = []int{t.a, t.b}, t.c
	} else {
		q := pickOne(pyQuadruples)
		comps, length = []int{q.a, q.b, q.c}, q.d
	}
	sumAbs := 0
	for i := range comps {
		sumAbs += comps[i]
		comps[i] *= randSign()
	}
	rand.Shuffle(len(comps), func(i, j int) { comps[i], comps[j] = comps[j], comps[i] })
	correct := float64(length)
	return question{
		prompt:  fmt.Sprintf("Panjang (besar) vektor u = (%s) adalah?", joinInts(comps)),
		options: buildOptions(correct, []float64{float64(sumAbs), float64(length * length), correct + 1, correct - 1}, optionCount, false),
	}
}

func genVectorLinearCombo(optionCount int) question {
	u := []int{between(-9, 9), between(-9, 9)}
	v := []int{between(-9, 9), between(-9, 9)}
	p, q := between(2, 4), between(2, 4)
	i := rand.IntN(2)
	axis := []string{"x", "y"}[i]
	correct := float64(p*u[i] - q*v[i])
	other := float64(p*u[1-i] - q*v[1-i])
	return question{
		prompt:  fmt.Sprintf("Diketahui u = (%d, %d) dan v = (%d, %d). Komponen %s dari vektor %du - %dv adalah?", u[0], u[1], v[0], v[1], axis, p, q),
		options: buildOptions(correct, []float64{float64(p*u[i] + q*v[i]), float64(u[i] - v[i]), other, correct + 1, correct - 1}, optionCount, true),
	}
}

func genVectorFromPoints(optionCount int) question {
	t := pickOne(pyTriples)
	k := between(1, 3)
	dx, dy := t.a*k*randSign(), t.b*k*randSign()
	if rand.IntN(2) == 0 {
		dx, dy = dy, dx
	}
	x1, y1 := between(-5, 5), between(-5, 5)
	correct := float64(t.c * k)
	return question{
		prompt:  fmt.Sprintf("Diketahui titik A(%d, %d) dan B(%d, %d). Panjang vektor AB adalah?", x1, y1, x1+dx, y1+dy),
		options: buildOptions(correct, []float64{float64((t.a + t.b) * k), correct * correct, correct + 1, correct - 1}, optionCount, false),
	}
}

// ---------- Kelas X–XI: Statistika & Peluang ----------

func genVariance(optionCount int) question {
	n := pickOne([]int{4, 5, 6})
	mean := between(10, 80)
	devs := randomDeviations(n, 6, func(sumSq int) bool { return sumSq%n == 0 })
	data := make([]int, n)
	sumSq := 0
	for i, d := range devs {
		data[i] = mean + d
		sumSq += d * d
	}
	correct := float64(sumSq / n)
	return question{
		prompt:  fmt.Sprintf("Ragam (varians) dari data %s adalah?", joinInts(data)),
		options: buildOptions(correct, []float64{float64(sumSq), round2(math.Sqrt(correct)), round2(float64(sumSq) / float64(n-1)), correct + 1, correct - 1}, optionCount, false),
	}
}

func genCombinationCount(optionCount int) question {
	n := between(5, 12)
	r := between(2, 4)
	correct := float64(combination(n, r))
	perm := float64(factorial(n) / factorial(n-r))
	return question{
		prompt:  fmt.Sprintf("Dari %d siswa akan dipilih %d orang untuk menjadi panitia. Banyak cara memilih panitia tersebut adalah?", n, r),
		options: buildOptions(correct, []float64{perm, float64(combination(n, r-1)), float64(combination(n, r+1)), float64(n * r)}, optionCount, false),
	}
}

func genArrangementWord(optionCount int) question {
	word := pickOne([]string{"BUKU", "KAKAK", "MATA", "SEPEDA", "KERETA", "SAMPAN", "MALAM", "PAPAN", "KELAS", "BELAJAR"})
	counts := map[rune]int{}
	for _, ch := range word {
		counts[ch]++
	}
	n := len(word)
	denom := 1
	for _, c := range counts {
		denom *= factorial(c)
	}
	correct := float64(factorial(n) / denom)
	return question{
		prompt:  fmt.Sprintf("Banyak susunan huruf berbeda yang dapat dibentuk dari huruf-huruf pada kata \"%s\" adalah?", word),
		options: buildOptions(correct, []float64{float64(factorial(n)), float64(factorial(n) / 2), 2 * correct, correct + 1, float64(factorial(n - 1))}, optionCount, false),
	}
}

func genExpectedFrequency(optionCount int) question {
	type event struct {
		desc string
		fav  int
	}
	e := pickOne([]event{
		{"mata dadu genap", 3}, {"mata dadu prima", 3}, {"mata dadu kelipatan 3", 2},
		{"mata dadu kurang dari 5", 4}, {"mata dadu 6", 1},
	})
	n := 6 * between(5, 50)
	correct := float64(n * e.fav / 6)
	return question{
		prompt:  fmt.Sprintf("Sebuah dadu dilempar %d kali. Frekuensi harapan muncul %s adalah?", n, e.desc),
		options: buildOptions(correct, []float64{float64(n / 6), float64(n * (6 - e.fav) / 6), float64(n * e.fav), correct + 1, correct - 1}, optionCount, false),
	}
}

func genSimpleProbability(optionCount int) question {
	total := pickOne([]int{4, 5, 10, 20, 25, 50}) // penyebut yang bikin desimal pas 2 angka
	red := between(1, total-1)
	blue := total - red
	correct := round2(float64(red) / float64(total))
	return question{
		prompt:  fmt.Sprintf("Sebuah kantong berisi %d kelereng merah dan %d kelereng biru. Diambil 1 kelereng secara acak. Peluang terambil kelereng merah adalah?", red, blue),
		options: buildOptions(correct, []float64{round2(float64(blue) / float64(total)), round2(float64(red) / float64(blue)), round2(correct + 0.1), round2(correct - 0.05), round2(correct + 0.05)}, optionCount, false),
	}
}

// ---------- Kelas X–XI: Program Linear ----------

type lpConstraint struct{ c, d, e int } // cx + dy ≤ e

// randomLPConstraints: 2 kendala yang titik potong & titik potong sumbunya
// semuanya bulat (dibangun mundur dari titik potong (x0, y0)).
func randomLPConstraints() []lpConstraint {
	k := between(2, 3)
	if rand.IntN(2) == 0 {
		x0, y0 := k*between(1, 6), between(1, 8)
		return []lpConstraint{{1, 1, x0 + y0}, {1, k, x0 + k*y0}}
	}
	x0, y0 := between(1, 8), k*between(1, 6)
	return []lpConstraint{{1, 1, x0 + y0}, {k, 1, k*x0 + y0}}
}

// lpVertices: titik-titik pojok daerah layak (x ≥ 0, y ≥ 0, plus kendala
// cx + dy ≤ e, atau ≥ e kalau atLeast).
func lpVertices(cons []lpConstraint, atLeast bool) [][2]float64 {
	lines := append([]lpConstraint{{1, 0, 0}, {0, 1, 0}}, cons...)
	var vertices [][2]float64
	seen := map[[2]float64]bool{}
	for i := range lines {
		for j := i + 1; j < len(lines); j++ {
			l1, l2 := lines[i], lines[j]
			det := l1.c*l2.d - l2.c*l1.d
			if det == 0 {
				continue
			}
			x := float64(l1.e*l2.d-l2.e*l1.d) / float64(det)
			y := float64(l1.c*l2.e-l2.c*l1.e) / float64(det)
			if x < -1e-9 || y < -1e-9 {
				continue
			}
			feasible := true
			for _, c := range cons {
				lhs, rhs := float64(c.c)*x+float64(c.d)*y, float64(c.e)
				if (!atLeast && lhs > rhs+1e-9) || (atLeast && lhs < rhs-1e-9) {
					feasible = false
					break
				}
			}
			key := [2]float64{round2(x), round2(y)}
			if !feasible || seen[key] {
				continue
			}
			seen[key] = true
			vertices = append(vertices, key)
		}
	}
	return vertices
}

// lpObjectiveValues: nilai f(x, y) = px + qy di tiap titik pojok, urut naik.
func lpObjectiveValues(cons []lpConstraint, atLeast bool, p, q int) []float64 {
	var values []float64
	for _, v := range lpVertices(cons, atLeast) {
		values = append(values, round2(float64(p)*v[0]+float64(q)*v[1]))
	}
	sort.Float64s(values)
	return values
}

// lpVertexValues: nilai maksimum (kendala ≤) + nilai titik pojok lain buat distraktor.
func lpVertexValues(cons []lpConstraint, p, q int) (best float64, others []float64) {
	values := lpObjectiveValues(cons, false, p, q)
	return values[len(values)-1], values[:len(values)-1]
}

func genLinearProgramMax(optionCount int) question {
	cons := randomLPConstraints()
	p, q := between(2, 9), between(2, 9)
	best, others := lpVertexValues(cons, p, q)
	return question{
		prompt: fmt.Sprintf(
			"Nilai maksimum f(x, y) = %dx + %dy yang memenuhi %s ≤ %d, %s ≤ %d, x ≥ 0, y ≥ 0 adalah?",
			p, q, formatLinearXY(cons[0].c, cons[0].d), cons[0].e, formatLinearXY(cons[1].c, cons[1].d), cons[1].e,
		),
		options: buildOptions(best, append(others, best+1, best-1, best+float64(p)), optionCount, false),
	}
}

// ---------- Kelas X–XI: Matriks ----------

func genMatrixDet2(optionCount int) question {
	a, b, c, d := between(-9, 9), between(-9, 9), between(-9, 9), between(-9, 9)
	correct := float64(a*d - b*c)
	return question{
		prompt:  fmt.Sprintf("Determinan matriks A = %s (baris dipisah titik koma) adalah?", formatMatrix([][]int{{a, b}, {c, d}})),
		options: buildOptions(correct, []float64{float64(a*d + b*c), -correct, float64(a * d), correct + 1, correct - 1}, optionCount, true),
	}
}

func genMatrixDet3(optionCount int) question {
	m := make([][]int, 3)
	for i := range m {
		m[i] = []int{between(-3, 5), between(-3, 5), between(-3, 5)}
	}
	plus := m[0][0]*m[1][1]*m[2][2] + m[0][1]*m[1][2]*m[2][0] + m[0][2]*m[1][0]*m[2][1]
	minus := m[0][2]*m[1][1]*m[2][0] + m[0][0]*m[1][2]*m[2][1] + m[0][1]*m[1][0]*m[2][2]
	correct := float64(plus - minus)
	return question{
		prompt:  fmt.Sprintf("Determinan matriks A = %s (baris dipisah titik koma) adalah?", formatMatrix(m)),
		options: buildOptions(correct, []float64{float64(plus + minus), -correct, float64(plus), correct + 1, correct - 1}, optionCount, true),
	}
}

func genMatrixProductEntry(optionCount int) question {
	a := [][]int{{between(-5, 6), between(-5, 6)}, {between(-5, 6), between(-5, 6)}}
	b := [][]int{{between(-5, 6), between(-5, 6)}, {between(-5, 6), between(-5, 6)}}
	i, j := rand.IntN(2), rand.IntN(2)
	correct := float64(a[i][0]*b[0][j] + a[i][1]*b[1][j])
	ba := float64(b[i][0]*a[0][j] + b[i][1]*a[1][j])
	return question{
		prompt: fmt.Sprintf(
			"Diketahui A = %s dan B = %s (baris dipisah titik koma). Elemen baris ke-%d kolom ke-%d dari A × B adalah?",
			formatMatrix(a), formatMatrix(b), i+1, j+1,
		),
		options: buildOptions(correct, []float64{ba, float64(a[i][j] * b[i][j]), correct + 1, correct - 1, -correct}, optionCount, true),
	}
}

func genMatrixInverseEntry(optionCount int) question {
	// A = [1+mn m; n 1] -> det = 1 -> A⁻¹ = [1 -m; -n 1+mn] (semua bulat)
	m, n := nonZeroBetween(-4, 4), nonZeroBetween(-4, 4)
	a := [][]int{{1 + m*n, m}, {n, 1}}
	inv := [][]int{{1, -m}, {-n, 1 + m*n}}
	if rand.IntN(2) == 0 { // variasi: A = [1 m; n 1+mn] -> A⁻¹ = [1+mn -m; -n 1]
		a = [][]int{{1, m}, {n, 1 + m*n}}
		inv = [][]int{{1 + m*n, -m}, {-n, 1}}
	}
	i, j := rand.IntN(2), rand.IntN(2)
	correct := float64(inv[i][j])
	return question{
		prompt:  fmt.Sprintf("Diketahui A = %s (baris dipisah titik koma). Elemen baris ke-%d kolom ke-%d dari invers matriks A (A⁻¹) adalah?", formatMatrix(a), i+1, j+1),
		options: buildOptions(correct, []float64{float64(a[i][j]), -correct, correct + 1, correct - 1, float64(a[1-i][1-j])}, optionCount, true),
	}
}

// ---------- Kelas X–XI: Lingkaran & Transformasi Geometri ----------

func circleEquation(a, b, r int) string {
	return fmt.Sprintf("x² + y²%s%s%s = 0", linTerm(-2*a, "x"), linTerm(-2*b, "y"), linTerm(a*a+b*b-r*r, ""))
}

func genCircleRadius(optionCount int) question {
	a, b, r := between(-6, 6), between(-6, 6), between(2, 10)
	correct := float64(r)
	return question{
		prompt:  fmt.Sprintf("Jari-jari lingkaran %s adalah?", circleEquation(a, b, r)),
		options: buildOptions(correct, []float64{float64(r * r), correct + 1, correct - 1, float64(absInt(a*a + b*b - r*r))}, optionCount, false),
	}
}

func genCircleCenter(optionCount int) question {
	a, b, r := nonZeroBetween(-6, 6), nonZeroBetween(-6, 6), between(2, 10)
	axis, correct := "x", float64(a)
	if rand.IntN(2) == 0 {
		axis, correct = "y", float64(b)
	}
	return question{
		prompt:  fmt.Sprintf("Koordinat %s titik pusat lingkaran %s adalah?", axis, circleEquation(a, b, r)),
		options: buildOptions(correct, []float64{-correct, 2 * correct, -2 * correct, correct + 1}, optionCount, true),
	}
}

func genCircleTangentGradient(optionCount int) question {
	// titik (x1, y1) dipilih biar gradien -x1/y1 desimal pas: (3,4), (12,5), (15,8)
	pt := pickOne([][2]int{{3, 4}, {12, 5}, {15, 8}})
	k := between(1, 2)
	x1, y1 := pt[0]*k*randSign(), pt[1]*k*randSign()
	r2 := x1*x1 + y1*y1
	correct := round2(-float64(x1) / float64(y1))
	return question{
		prompt:  fmt.Sprintf("Gradien garis singgung lingkaran x² + y² = %d di titik (%d, %d) adalah?", r2, x1, y1),
		options: buildOptions(correct, []float64{-correct, round2(float64(y1) / float64(x1)), round2(-float64(y1) / float64(x1)), correct + 1}, optionCount, true),
	}
}

type transform struct {
	desc  string
	apply func(x, y int) (int, int)
}

func randomTransform() transform {
	switch rand.IntN(7) {
	case 0:
		a, b := nonZeroBetween(-6, 6), nonZeroBetween(-6, 6)
		return transform{fmt.Sprintf("ditranslasi oleh T(%d, %d)", a, b), func(x, y int) (int, int) { return x + a, y + b }}
	case 1:
		return transform{"dicerminkan terhadap sumbu x", func(x, y int) (int, int) { return x, -y }}
	case 2:
		return transform{"dicerminkan terhadap sumbu y", func(x, y int) (int, int) { return -x, y }}
	case 3:
		return transform{"dicerminkan terhadap garis y = x", func(x, y int) (int, int) { return y, x }}
	case 4:
		return transform{"dirotasi 90° berlawanan arah jarum jam terhadap titik O", func(x, y int) (int, int) { return -y, x }}
	case 5:
		return transform{"dirotasi 180° terhadap titik O", func(x, y int) (int, int) { return -x, -y }}
	default:
		k := between(2, 3)
		return transform{fmt.Sprintf("didilatasi dengan pusat O dan faktor skala %d", k), func(x, y int) (int, int) { return k * x, k * y }}
	}
}

func genTransformation(optionCount int) question {
	x, y := nonZeroBetween(-8, 8), nonZeroBetween(-8, 8)
	t1, t2 := randomTransform(), randomTransform()
	x1, y1 := t1.apply(x, y)
	x2, y2 := t2.apply(x1, y1)
	rx, ry := t1.apply(t2.apply(x, y)) // urutan kebalik
	axis, correct, other, reversed, firstOnly := "x", x2, y2, rx, x1
	if rand.IntN(2) == 0 {
		axis, correct, other, reversed, firstOnly = "y", y2, x2, ry, y1
	}
	return question{
		prompt:  fmt.Sprintf("Titik P(%d, %d) %s, lalu %s. Koordinat %s bayangan titik P adalah?", x, y, t1.desc, t2.desc, axis),
		options: buildOptions(float64(correct), []float64{float64(other), float64(-correct), float64(reversed), float64(firstOnly), float64(correct + 1)}, optionCount, true),
	}
}

// ---------- Kelas XII: Limit Fungsi ----------

func genLimitFactor(optionCount int) question {
	// (x² + (k-a)x - ka) / (x - a) = x + k  ->  limit x→a = a + k
	a := nonZeroBetween(-5, 5)
	k := between(-6, 6)
	correct := float64(a + k)
	return question{
		prompt:  fmt.Sprintf("lim (x→%d) (%s) / (%s) = ?", a, formatPoly(1, k-a, -k*a), formatPoly(1, -a)),
		options: buildOptions(correct, []float64{float64(k), float64(a - k), float64(a * k), correct + 1, 0}, optionCount, true),
	}
}

func genLimitInfinity(optionCount int) question {
	q := between(1, 5)
	d, e := between(-9, 9), between(-9, 9)
	denom := formatPoly(q, d, e)
	if rand.IntN(4) == 0 { // pangkat pembilang < penyebut -> 0
		b, c := nonZeroBetween(-9, 9), between(-9, 9)
		return question{
			prompt:  fmt.Sprintf("lim (x→∞) (%s) / (%s) = ?", formatPoly(b, c), denom),
			options: buildOptions(0, []float64{round2(float64(b) / float64(q)), float64(b), 1, -1}, optionCount, true),
		}
	}
	m := nonZeroBetween(-4, 6)
	p := q * m
	b, c := between(-9, 9), between(-9, 9)
	correct := float64(m)
	return question{
		prompt:  fmt.Sprintf("lim (x→∞) (%s) / (%s) = ?", formatPoly(p, b, c), denom),
		options: buildOptions(correct, []float64{float64(p), float64(q), 0, correct + 1, -correct}, optionCount, true),
	}
}

func genLimitTrig(optionCount int) question {
	b := pickOne([]int{2, 4, 5}) // a/b pasti desimal pas
	a := between(2, 12)
	for a == b {
		a = between(2, 12)
	}
	correct := round2(float64(a) / float64(b))
	// semua bentuk ini limitnya a/b
	form := pickOne([]string{"sin(%dx) / (%dx)", "tan(%dx) / (%dx)", "sin(%dx) / tan(%dx)", "%dx / sin(%dx)"})
	return question{
		prompt:  "lim (x→0) " + fmt.Sprintf(form, a, b) + " = ?",
		options: buildOptions(correct, []float64{round2(float64(b) / float64(a)), float64(a), 0, 1}, optionCount, false),
	}
}

func genLimitSubstitution(optionCount int) question {
	a, b, c, d := between(1, 3), between(-5, 5), between(-8, 8), between(-10, 10)
	k := nonZeroBetween(-3, 4)
	eval := func(x int) int { return a*x*x*x + b*x*x + c*x + d }
	correct := float64(eval(k))
	return question{
		prompt:  fmt.Sprintf("lim (x→%d) (%s) = ?", k, formatPoly(a, b, c, d)),
		options: buildOptions(correct, []float64{float64(eval(-k)), float64(a + b + c + d), correct + 1, correct - 1, float64(d)}, optionCount, true),
	}
}

// ---------- Kelas XII: Turunan ----------

func genDerivativeAtPoint(optionCount int) question {
	a, b, c, d := between(1, 4), between(-8, 8), between(-8, 8), between(-10, 10)
	k := between(-3, 4)
	correct := float64(3*a*k*k + 2*b*k + c)
	fk := float64(a*k*k*k + b*k*k + c*k + d)
	return question{
		prompt:  fmt.Sprintf("Jika f(x) = %s, nilai f'(%d) = ?", formatPoly(a, b, c, d), k),
		options: buildOptions(correct, []float64{fk, float64(a*k*k + b*k + c), -correct, correct + 1, correct - 1}, optionCount, true),
	}
}

func genTangentSlope(optionCount int) question {
	a, b, c := nonZeroBetween(-4, 5), between(-9, 9), between(-9, 9)
	k := between(-4, 5)
	correct := float64(2*a*k + b)
	return question{
		prompt:  fmt.Sprintf("Gradien garis singgung kurva y = %s di titik berabsis x = %d adalah?", formatPoly(a, b, c), k),
		options: buildOptions(correct, []float64{float64(a*k*k + b*k + c), float64(a*k + b), correct + 2, correct - 2, -correct}, optionCount, true),
	}
}

func genStationaryPoint(optionCount int) question {
	if rand.IntN(2) == 0 {
		// f(x) = -a x² + 2ah x + c -> maksimum di x = h, nilai a h² + c
		a, h, c := between(1, 3), nonZeroBetween(-5, 6), between(-10, 10)
		correct := float64(a*h*h + c)
		return question{
			prompt:  fmt.Sprintf("Nilai maksimum fungsi f(x) = %s adalah?", formatPoly(-a, 2*a*h, c)),
			options: buildOptions(correct, []float64{float64(h), float64(c), float64(a*h*h - c), correct + 1, correct - 1}, optionCount, true),
		}
	}
	// f(x) = x³ - 3p²x + c -> f'(x) = 3(x - p)(x + p), maksimum lokal di x = -p
	p, c := between(1, 5), between(-10, 10)
	correct := float64(-p)
	return question{
		prompt:  fmt.Sprintf("Fungsi f(x) = %s mencapai maksimum lokal saat x = ?", formatPoly(1, 0, -3*p*p, c)),
		options: buildOptions(correct, []float64{float64(p), 0, float64(3 * p), float64(-3 * p), float64(p * p)}, optionCount, true),
	}
}

func genVelocity(optionCount int) question {
	a, b, c, d := between(1, 3), between(-6, 6), between(0, 10), between(0, 20)
	t := between(1, 5)
	pos := float64(a*t*t*t + b*t*t + c*t + d)
	vel := float64(3*a*t*t + 2*b*t + c)
	acc := float64(6*a*t + 2*b)
	s := formatPolyVar("t", a, b, c, d)
	if rand.IntN(2) == 0 {
		return question{
			prompt:  fmt.Sprintf("Posisi benda s(t) = %s meter (t dalam detik). Kecepatan benda saat t = %d detik adalah? (m/s)", s, t),
			options: buildOptions(vel, []float64{pos, acc, vel + 1, vel - 1}, optionCount, true),
		}
	}
	return question{
		prompt:  fmt.Sprintf("Posisi benda s(t) = %s meter (t dalam detik). Percepatan benda saat t = %d detik adalah? (m/s²)", s, t),
		options: buildOptions(acc, []float64{vel, pos, acc + 2, acc - 2}, optionCount, true),
	}
}

// ---------- Kelas XII: Integral ----------

func genDefiniteIntegral(optionCount int) question {
	m, n, c := between(1, 3), between(-4, 4), between(-5, 6)
	a := between(-2, 2)
	b := a + between(1, 3)
	antideriv := func(x int) int { return m*x*x*x + n*x*x + c*x }
	integrand := func(x int) int { return 3*m*x*x + 2*n*x + c }
	correct := float64(antideriv(b) - antideriv(a))
	return question{
		prompt:  fmt.Sprintf("∫ dari %d sampai %d (%s) dx = ?", a, b, formatPoly(3*m, 2*n, c)),
		options: buildOptions(correct, []float64{float64(antideriv(b)), float64(antideriv(b) + antideriv(a)), float64(integrand(b) - integrand(a)), correct + 1, correct - 1}, optionCount, true),
	}
}

func genIndefiniteCoeff(optionCount int) question {
	p, q, r := between(1, 4), nonZeroBetween(-5, 5), nonZeroBetween(-6, 6)
	correct := float64(p + q + r)
	return question{
		prompt:  fmt.Sprintf("Jika ∫ (%s) dx = ax³ + bx² + cx + C, nilai a + b + c = ?", formatPoly(3*p, 2*q, r)),
		options: buildOptions(correct, []float64{float64(3*p + 2*q + r), float64(p + q), correct + 1, correct - 1, float64(9*p + 4*q + r)}, optionCount, true),
	}
}

func genAreaUnderCurve(optionCount int) question {
	m, n, b := between(1, 3), between(0, 3), between(1, 4)
	correct := float64(m*b*b*b + n*b*b)
	return question{
		prompt:  fmt.Sprintf("Luas daerah yang dibatasi kurva y = %s, sumbu x, garis x = 0, dan garis x = %d adalah? (satuan luas)", formatPoly(3*m, 2*n, 0), b),
		options: buildOptions(correct, []float64{float64(3*m*b*b + 2*n*b), float64(3*m*b*b*b + 2*n*b*b), correct + float64(b), correct - float64(b)}, optionCount, false),
	}
}

func genAreaBetweenCurves(optionCount int) question {
	if rand.IntN(2) == 0 {
		// y = x² dan y = cx -> luas c³/6
		c := pickOne([]int{3, 6, 9, 12})
		c3 := float64(c * c * c)
		correct := round2(c3 / 6)
		return question{
			prompt:  fmt.Sprintf("Luas daerah yang dibatasi kurva y = x² dan garis y = %dx adalah? (satuan luas)", c),
			options: buildOptions(correct, []float64{round2(c3 / 2), round2(c3 / 3), round2(float64(c*c) / 2), correct + 1}, optionCount, false),
		}
	}
	// y = x² dan y = s² -> luas 4s³/3
	s := pickOne([]int{3, 6})
	s3 := float64(s * s * s)
	correct := 4 * s3 / 3
	return question{
		prompt:  fmt.Sprintf("Luas daerah yang dibatasi kurva y = x² dan garis y = %d adalah? (satuan luas)", s*s),
		options: buildOptions(correct, []float64{2 * s3 / 3, s3 / 3, 2 * s3, correct + 1}, optionCount, false),
	}
}

// ---------- Kelas XII: Statistika Lanjut & Terapan ----------

func genFrequencyMean(optionCount int) question {
	for {
		k := between(3, 4)
		start := between(5, 7) * 10
		values, freqs := make([]int, k), make([]int, k)
		sum, total, plain := 0, 0, 0
		parts := make([]string, k)
		for i := range k {
			values[i] = start + 10*i
			freqs[i] = between(1, 8)
			sum += values[i] * freqs[i]
			total += freqs[i]
			plain += values[i]
			parts[i] = fmt.Sprintf("nilai %d ada %d siswa", values[i], freqs[i])
		}
		if sum%total != 0 {
			continue
		}
		correct := float64(sum / total)
		return question{
			prompt:  fmt.Sprintf("Hasil ulangan satu kelas: %s. Rata-rata nilai kelas tersebut adalah?", strings.Join(parts, ", ")),
			options: buildOptions(correct, []float64{round2(float64(plain) / float64(k)), correct + 1, correct - 1, correct + 5, correct - 5}, optionCount, false),
		}
	}
}

func genInterquartileRange(optionCount int) question {
	n := pickOne([]int{7, 11})
	data := make([]int, n)
	for i := range data {
		data[i] = between(10, 90)
	}
	sorted := append([]int{}, data...)
	sort.Ints(sorted)
	q1 := sorted[(n+1)/4-1]
	q3 := sorted[3*(n+1)/4-1]
	median := sorted[n/2]
	correct := float64(q3 - q1)
	return question{
		prompt:  fmt.Sprintf("Jangkauan antarkuartil (Q3 - Q1) dari data %s adalah?", joinInts(data)),
		options: buildOptions(correct, []float64{float64(sorted[n-1] - sorted[0]), float64(q3), float64(q1), float64(median), correct + 1}, optionCount, false),
	}
}

func genStdDeviation(optionCount int) question {
	n := pickOne([]int{4, 5, 6})
	mean := between(20, 80)
	devs := randomDeviations(n, 6, func(sumSq int) bool {
		if sumSq%n != 0 {
			return false
		}
		root := int(math.Sqrt(float64(sumSq / n)))
		return root*root == sumSq/n
	})
	data := make([]int, n)
	sumSq := 0
	for i, d := range devs {
		data[i] = mean + d
		sumSq += d * d
	}
	variance := float64(sumSq / n)
	correct := math.Sqrt(variance)
	return question{
		prompt:  fmt.Sprintf("Simpangan baku dari data %s adalah?", joinInts(data)),
		options: buildOptions(correct, []float64{variance, round2(math.Sqrt(float64(sumSq) / float64(n-1))), correct + 1, correct - 1, float64(sumSq)}, optionCount, false),
	}
}

func genRegressionPredict(optionCount int) question {
	a := between(5, 50)
	b := pickOne([]float64{0.5, 1.5, 2, 2.5, 3, 4})
	x := between(2, 20)
	correct := round2(float64(a) + b*float64(x))
	return question{
		prompt: fmt.Sprintf(
			"Hasil regresi linear antara biaya promosi x (juta rupiah) dan omzet y (juta rupiah) adalah y = %d + %sx. Jika biaya promosi %d juta rupiah, perkiraan omzetnya adalah? (juta rupiah)",
			a, fmtNum(b), x,
		),
		options: buildOptions(correct, []float64{round2((float64(a) + b) * float64(x)), round2(b * float64(x)), round2(float64(a)*float64(x) + b), round2(correct + b)}, optionCount, false),
	}
}

func genNewMemberAverage(optionCount int) question {
	n := between(4, 9)
	old := between(20, 60) * 100 // ribu rupiah
	d := nonZeroBetween(-3, 6) * 50
	newAvg := old + d
	correct := float64(old + (n+1)*d)
	return question{
		prompt: fmt.Sprintf(
			"Rata-rata gaji %d karyawan sebuah toko adalah Rp%d ribu. Setelah 1 karyawan baru bergabung, rata-ratanya menjadi Rp%d ribu. Gaji karyawan baru tersebut adalah? (ribu rupiah)",
			n, old, newAvg,
		),
		options: buildOptions(correct, []float64{float64(newAvg), float64(old + n*d), float64((n + 1) * newAvg), correct + 50, correct - 50}, optionCount, false),
	}
}

// ---------- Kelas XII: Dimensi Tiga ----------

func genCuboidSpaceDiagonal(optionCount int) question {
	q := pickOne(pyQuadruples)
	k := between(1, 3)
	dims := []int{q.a * k, q.b * k, q.c * k}
	rand.Shuffle(3, func(i, j int) { dims[i], dims[j] = dims[j], dims[i] })
	correct := float64(q.d * k)
	return question{
		prompt:  fmt.Sprintf("Balok ABCD.EFGH berukuran panjang %d cm, lebar %d cm, dan tinggi %d cm. Panjang diagonal ruang AG adalah? (cm)", dims[0], dims[1], dims[2]),
		options: buildOptions(correct, []float64{float64(dims[0] + dims[1] + dims[2]), round2(math.Sqrt(float64(dims[0]*dims[0] + dims[1]*dims[1]))), correct + 1, correct - 1}, optionCount, false),
	}
}

func genCuboidFaceDiagonal(optionCount int) question {
	t := pickOne(pyTriples)
	k := between(1, 3)
	p, l := t.a*k, t.b*k
	if rand.IntN(2) == 0 {
		p, l = l, p
	}
	h := between(3, 20)
	correct := float64(t.c * k)
	return question{
		prompt:  fmt.Sprintf("Balok ABCD.EFGH dengan AB = %d cm, BC = %d cm, dan AE = %d cm. Panjang diagonal bidang AC adalah? (cm)", p, l, h),
		options: buildOptions(correct, []float64{float64(p + l), round2(math.Sqrt(float64(p*p + l*l + h*h))), correct + 1, correct - 1}, optionCount, false),
	}
}

func genCuboidPointToLine(optionCount int) question {
	// jarak E ke garis CG = panjang EG = diagonal alas
	t := pickOne(pyTriples)
	k := between(1, 3)
	p, l := t.a*k, t.b*k
	if rand.IntN(2) == 0 {
		p, l = l, p
	}
	h := between(3, 20)
	correct := float64(t.c * k)
	return question{
		prompt:  fmt.Sprintf("Balok ABCD.EFGH dengan AB = %d cm, BC = %d cm, dan AE = %d cm. Jarak titik E ke garis CG adalah? (cm)", p, l, h),
		options: buildOptions(correct, []float64{float64(h), round2(math.Sqrt(float64(p*p + l*l + h*h))), float64(p + l), correct + 1}, optionCount, false),
	}
}

func genCubeAngle(optionCount int) question {
	type pair struct {
		lines string
		angle int
	}
	p := pickOne([]pair{
		{"AE dan AF", 45}, {"AC dan BD", 90}, {"AF dan AH", 60}, {"AC dan AF", 60},
		{"AB dan FG", 90}, {"AC dan EG", 0}, {"EG dan BD", 90}, {"AF dan BE", 90}, {"AH dan BG", 0},
	})
	return question{
		prompt:  fmt.Sprintf("Pada kubus ABCD.EFGH dengan rusuk %d cm, besar sudut antara garis %s adalah? (derajat)", between(4, 12), p.lines),
		options: buildOptions(float64(p.angle), []float64{0, 30, 45, 60, 90}, optionCount, false),
	}
}

// ---------- Kelas XII: Irisan Kerucut ----------

func genParabola(optionCount int) question {
	p := nonZeroBetween(-6, 6)
	switch rand.IntN(4) {
	case 0:
		correct := float64(p)
		return question{
			prompt:  fmt.Sprintf("Koordinat x titik fokus parabola y² = %dx adalah?", 4*p),
			options: buildOptions(correct, []float64{float64(4 * p), -correct, float64(2 * p), correct + 1}, optionCount, true),
		}
	case 1:
		correct := float64(p)
		return question{
			prompt:  fmt.Sprintf("Koordinat y titik fokus parabola x² = %dy adalah?", 4*p),
			options: buildOptions(correct, []float64{float64(4 * p), -correct, float64(2 * p), correct + 1}, optionCount, true),
		}
	case 2:
		correct := float64(-p)
		return question{
			prompt:  fmt.Sprintf("Persamaan garis direktris parabola y² = %dx adalah x = ?", 4*p),
			options: buildOptions(correct, []float64{float64(p), float64(-4 * p), float64(4 * p), correct - 1}, optionCount, true),
		}
	default:
		a, b := between(-5, 5), between(-5, 5)
		correct := float64(a + p)
		return question{
			prompt:  fmt.Sprintf("Koordinat x titik fokus parabola (y%s)² = %d(x%s) adalah?", linTerm(-b, ""), 4*p, linTerm(-a, "")),
			options: buildOptions(correct, []float64{float64(p), float64(a - p), float64(a + 4*p), float64(-a + p)}, optionCount, true),
		}
	}
}

func genEllipse(optionCount int) question {
	// a² = b² + c² -> pakai tripel: a = sisi miring, b & c = sisi siku
	t := pickOne(pyTriples)
	a, b, c := t.c, t.a, t.b
	if rand.IntN(2) == 0 {
		b, c = c, b
	}
	eq := fmt.Sprintf("x²/%d + y²/%d = 1", a*a, b*b)
	all := []float64{float64(a), float64(b), float64(c), float64(2 * a), float64(2 * b), float64(2 * c)}
	var prompt string
	var correct float64
	switch rand.IntN(3) {
	case 0:
		prompt, correct = fmt.Sprintf("Koordinat x titik fokus elips %s yang bernilai positif adalah?", eq), float64(c)
	case 1:
		prompt, correct = fmt.Sprintf("Panjang sumbu mayor elips %s adalah?", eq), float64(2*a)
	default:
		prompt, correct = fmt.Sprintf("Jarak antara kedua titik fokus elips %s adalah?", eq), float64(2*c)
	}
	return question{prompt: prompt, options: buildOptions(correct, all, optionCount, false)}
}

func genHyperbola(optionCount int) question {
	// c² = a² + b²
	t := pickOne(pyTriples)
	a, b, c := t.a, t.b, t.c
	if rand.IntN(2) == 0 {
		a, b = b, a
	}
	eq := fmt.Sprintf("x²/%d - y²/%d = 1", a*a, b*b)
	all := []float64{float64(a), float64(b), float64(c), float64(2 * a), float64(2 * c)}
	if rand.IntN(2) == 0 {
		return question{
			prompt:  fmt.Sprintf("Koordinat x titik fokus hiperbola %s yang bernilai positif adalah?", eq),
			options: buildOptions(float64(c), all, optionCount, false),
		}
	}
	return question{
		prompt:  fmt.Sprintf("Jarak antara kedua titik puncak hiperbola %s adalah?", eq),
		options: buildOptions(float64(2*a), all, optionCount, false),
	}
}

// ---------- SMK: Matematika Keuangan ----------

func genSimpleInterest(optionCount int) question {
	m := between(10, 100) * 100 // ribu rupiah
	r := between(4, 12)
	n := between(2, 5)
	interest := m * r * n / 100
	if rand.IntN(2) == 0 {
		return question{
			prompt:  fmt.Sprintf("Modal Rp%d ribu disimpan di bank dengan bunga tunggal %d%% per tahun. Besar bunga setelah %d tahun adalah? (ribu rupiah)", m, r, n),
			options: buildOptions(float64(interest), []float64{float64(m * r / 100), float64(m + interest), float64(interest + m*r/100), float64(interest - m*r/100)}, optionCount, false),
		}
	}
	correct := float64(m + interest)
	return question{
		prompt:  fmt.Sprintf("Modal Rp%d ribu disimpan di bank dengan bunga tunggal %d%% per tahun. Jumlah tabungan setelah %d tahun adalah? (ribu rupiah)", m, r, n),
		options: buildOptions(correct, []float64{float64(interest), float64(m + m*r/100), correct + float64(m*r/100), correct - float64(m*r/100)}, optionCount, false),
	}
}

func genCompoundInterest(optionCount int) question {
	r := pickOne([]int{10, 20})
	n := pickOne([]int{2, 3})
	m := between(1, 20) * 1000 // ribu rupiah, kelipatan 1000 biar hasil bulat
	correct := float64(m * intPow(100+r, n) / intPow(100, n))
	simple := float64(m + m*r*n/100)
	return question{
		prompt:  fmt.Sprintf("Modal Rp%d ribu disimpan dengan bunga majemuk %d%% per tahun. Nilai akhir modal setelah %d tahun adalah? (ribu rupiah)", m, r, n),
		options: buildOptions(correct, []float64{simple, float64(m + m*r/100), correct + float64(m/100), correct - float64(m/100)}, optionCount, false),
	}
}

func genAnnuity(optionCount int) question {
	loan := between(10, 100) * 1000 // ribu rupiah
	i := between(1, 3)              // % per bulan
	firstInterest := loan * i / 100
	firstPrincipal := between(5, 40) * 100
	annuity := firstInterest + firstPrincipal
	base := fmt.Sprintf("Pinjaman Rp%d ribu dilunasi dengan anuitas Rp%d ribu per bulan dan suku bunga %d%% per bulan.", loan, annuity, i)
	switch rand.IntN(3) {
	case 0:
		return question{
			prompt:  base + " Besar angsuran pokok pada bulan pertama adalah? (ribu rupiah)",
			options: buildOptions(float64(firstPrincipal), []float64{float64(firstInterest), float64(annuity), float64(annuity + firstInterest), float64(firstPrincipal + 100)}, optionCount, false),
		}
	case 1:
		return question{
			prompt:  base + " Besar bunga pada bulan pertama adalah? (ribu rupiah)",
			options: buildOptions(float64(firstInterest), []float64{float64(firstPrincipal), float64(annuity), float64(loan * i / 10), float64(firstInterest + 100)}, optionCount, false),
		}
	default:
		return question{
			prompt:  base + " Sisa pinjaman setelah pembayaran bulan pertama adalah? (ribu rupiah)",
			options: buildOptions(float64(loan-firstPrincipal), []float64{float64(loan - annuity), float64(loan - firstInterest), float64(loan - firstPrincipal + firstInterest), float64(loan)}, optionCount, false),
		}
	}
}

// ---------- SMK: Matriks & Program Linear Terapan ----------

func genMatrixSalesRevenue(optionCount int) question {
	s := [][]int{{between(5, 30), between(5, 30)}, {between(5, 30), between(5, 30)}}
	p1, p2 := between(10, 40), between(10, 40)
	shop := rand.IntN(2)
	name := []string{"A", "B"}[shop]
	correct := float64(s[shop][0]*p1 + s[shop][1]*p2)
	return question{
		prompt: fmt.Sprintf(
			"Toko A menjual %d kg apel dan %d kg jeruk, toko B menjual %d kg apel dan %d kg jeruk. Harga apel Rp%d ribu/kg dan jeruk Rp%d ribu/kg. Dengan perkalian matriks, total pendapatan toko %s adalah? (ribu rupiah)",
			s[0][0], s[0][1], s[1][0], s[1][1], p1, p2, name,
		),
		options: buildOptions(correct, []float64{
			float64(s[1-shop][0]*p1 + s[1-shop][1]*p2),
			float64(s[shop][0]*p2 + s[shop][1]*p1),
			float64((s[shop][0] + s[shop][1]) * (p1 + p2)),
			correct + float64(p1),
		}, optionCount, false),
	}
}

func genMatrixPriceSystem(optionCount int) question {
	x, y := between(10, 30), between(10, 30) // harga gula & beras (ribu/kg)
	a, b, c, d := between(1, 6), between(1, 6), between(1, 6), between(1, 6)
	for a*d == b*c {
		d = between(1, 6)
	}
	item, correct, other := "gula", float64(x), float64(y)
	if rand.IntN(2) == 0 {
		item, correct, other = "beras", float64(y), float64(x)
	}
	return question{
		prompt: fmt.Sprintf(
			"Harga %d kg gula dan %d kg beras adalah Rp%d ribu, sedangkan harga %d kg gula dan %d kg beras adalah Rp%d ribu. Harga 1 kg %s adalah? (ribu rupiah)",
			a, b, a*x+b*y, c, d, c*x+d*y, item,
		),
		options: buildOptions(correct, []float64{other, float64(x + y), correct + 1, correct - 1, correct + 2}, optionCount, false),
	}
}

func genLinearProgramProduction(optionCount int) question {
	cons := randomLPConstraints()
	p, q := between(2, 9)*5, between(2, 9)*5 // keuntungan per unit (ribu rupiah)
	best, others := lpVertexValues(cons, p, q)
	return question{
		prompt: fmt.Sprintf(
			"Sebuah pabrik membuat produk A dan B. Setiap unit A butuh %d jam mesin dan %d jam perakitan, setiap unit B butuh %d jam mesin dan %d jam perakitan. Tersedia %d jam mesin dan %d jam perakitan. Keuntungan per unit A Rp%d ribu dan B Rp%d ribu. Keuntungan maksimum adalah? (ribu rupiah)",
			cons[0].c, cons[1].c, cons[0].d, cons[1].d, cons[0].e, cons[1].e, p, q,
		),
		options: buildOptions(best, append(others, best+float64(p), best-float64(q), best+5), optionCount, false),
	}
}

// ---------- SMK: Trigonometri Terapan ----------

func genSlopeHeight(optionCount int) question {
	l := 2 * between(10, 100)
	half := float64(l / 2)
	switch rand.IntN(3) {
	case 0:
		return question{
			prompt:  fmt.Sprintf("Seutas benang layang-layang sepanjang %d m membentuk sudut 30° dengan tanah (benang dianggap lurus). Tinggi layang-layang dari tanah adalah? (m)", l),
			options: buildOptions(half, []float64{float64(l), round2(float64(l) * 0.87), float64(l / 4), half + 10}, optionCount, false),
		}
	case 1:
		return question{
			prompt:  fmt.Sprintf("Sebuah jalan menanjak dengan sudut kemiringan 30°. Setelah menempuh %d m di sepanjang jalan itu, kenaikan tingginya adalah? (m)", l),
			options: buildOptions(half, []float64{float64(l), round2(float64(l) * 0.87), float64(l / 4), half + 10}, optionCount, false),
		}
	default:
		return question{
			prompt:  fmt.Sprintf("Sebuah tangga sepanjang %d dm bersandar pada tembok dan membentuk sudut 60° dengan lantai. Jarak kaki tangga ke tembok adalah? (dm)", l),
			options: buildOptions(half, []float64{float64(l), round2(float64(l) * 0.87), float64(l / 4), half + 10}, optionCount, false),
		}
	}
}

func genElevation45(optionCount int) question {
	d := between(5, 60)
	eye := pickOne([]float64{1.5, 1.6, 1.7})
	correct := round2(float64(d) + eye)
	return question{
		prompt:  fmt.Sprintf("Seorang pengamat dengan tinggi mata %s m berdiri %d m dari sebuah tiang dan melihat puncak tiang dengan sudut elevasi 45°. Tinggi tiang adalah? (m)", fmtNum(eye), d),
		options: buildOptions(correct, []float64{float64(d), round2(float64(d) - eye), round2(2*float64(d) + eye), round2(correct + 1)}, optionCount, false),
	}
}

func genShipDistance(optionCount int) question {
	t := pickOne(angleTriangles)
	k := between(1, 5)
	a, b := t.a*k, t.b*k
	correct := float64(t.c * k)
	return question{
		prompt: fmt.Sprintf(
			"Dua kapal berangkat dari pelabuhan yang sama. Kapal A berlayar sejauh %d km dan kapal B sejauh %d km, dengan sudut antara arah keduanya %d°. Jarak kedua kapal sekarang adalah? (km)",
			a, b, t.angle,
		),
		options: buildOptions(correct, []float64{float64(a + b), float64(absInt(a - b)), round2(math.Sqrt(float64(a*a + b*b))), correct + float64(k)}, optionCount, false),
	}
}

func genLadder(optionCount int) question {
	t := pickOne(pyTriples)
	k := between(1, 2)
	foot, top := t.a*k, t.b*k
	if rand.IntN(2) == 0 {
		foot, top = top, foot
	}
	correct := float64(top)
	return question{
		prompt:  fmt.Sprintf("Sebuah tangga sepanjang %d m bersandar pada tembok tegak. Kaki tangga berjarak %d m dari tembok. Tinggi ujung atas tangga pada tembok adalah? (m)", t.c*k, foot),
		options: buildOptions(correct, []float64{float64(foot), float64(t.c*k - foot), float64(t.c*k + foot), correct + 1}, optionCount, false),
	}
}

// ---------- SMK: Logika Matematika ----------

type logicExpr struct {
	text string
	eval func(vals []bool) bool
}

func randomLiteral(idx int) logicExpr {
	name := []string{"p", "q", "r", "s"}[idx]
	if rand.IntN(3) == 0 {
		return logicExpr{"~" + name, func(v []bool) bool { return !v[idx] }}
	}
	return logicExpr{name, func(v []bool) bool { return v[idx] }}
}

func combineLogic(l, r logicExpr) logicExpr {
	switch rand.IntN(4) {
	case 0:
		return logicExpr{l.text + " ∧ " + r.text, func(v []bool) bool { return l.eval(v) && r.eval(v) }}
	case 1:
		return logicExpr{l.text + " ∨ " + r.text, func(v []bool) bool { return l.eval(v) || r.eval(v) }}
	case 2:
		return logicExpr{l.text + " → " + r.text, func(v []bool) bool { return !l.eval(v) || r.eval(v) }}
	default:
		return logicExpr{l.text + " ↔ " + r.text, func(v []bool) bool { return l.eval(v) == r.eval(v) }}
	}
}

func parenthesize(e logicExpr) logicExpr { return logicExpr{"(" + e.text + ")", e.eval} }

func genTruthTableCount(optionCount int) question {
	vars := between(2, 3)
	var expr logicExpr
	if vars == 2 {
		expr = combineLogic(randomLiteral(0), randomLiteral(1))
	} else if rand.IntN(2) == 0 {
		expr = combineLogic(parenthesize(combineLogic(randomLiteral(0), randomLiteral(1))), randomLiteral(2))
	} else {
		expr = combineLogic(randomLiteral(0), parenthesize(combineLogic(randomLiteral(1), randomLiteral(2))))
	}
	rows := 1 << vars
	trueCount := 0
	for mask := range rows {
		vals := make([]bool, vars)
		for i := range vars {
			vals[i] = mask&(1<<i) != 0
		}
		if expr.eval(vals) {
			trueCount++
		}
	}
	word, correct := "BENAR", trueCount
	if rand.IntN(2) == 0 {
		word, correct = "SALAH", rows-trueCount
	}
	return question{
		prompt:  fmt.Sprintf("Banyak baris bernilai %s pada tabel kebenaran pernyataan %s adalah?", word, expr.text),
		options: buildOptions(float64(correct), []float64{float64(rows - correct), float64(correct + 1), float64(correct - 1), float64(rows)}, optionCount, false),
	}
}
