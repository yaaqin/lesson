package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strings"
)

// ---------- Kampus ----------
//
// 25 materi kuliah dibagi 2 batch per-materi (satu level = satu materi, ujian =
// campuran materi batch itu), lalu 3 batch campuran:
//   Batch 1 (12 materi) & 2 (13 materi): level wajib 20 dari bank 35; ujian wajib 30 dari 50
//   Batch 3: 5 level campuran wajib 20 dari 35; ujian wajib 20 dari 35
//   Batch 4: 10 level campuran wajib 25 dari 40; ujian wajib 40 dari 60
//   Batch 5: 10 level campuran sulit wajib 25 dari 50; ujian wajib 40 dari 75
// Semua 5 opsi & jawaban numerik (lihat catatan di smk_materi.go). File ini:
// spec tier + helper + materi batch 1; kampus_2.go: materi batch 2;
// kampus_sulit.go: generator batch 5.

const (
	kampusOptionCount  = 5
	kampusPassPercent  = 70
	kampusSecondsLevel = 90  // per soal, batch 1-4
	kampusSecondsHard  = 120 // per soal, batch 5
)

func kampusTierSpec() tierSpec {
	b1, b2 := kampusMateriBatch1(), kampusMateriBatch2()
	var gens []func(int) question
	for _, m := range append(append([]materiSpec{}, b1...), b2...) {
		gens = append(gens, m.gen)
	}
	all := mixOf(gens...)

	return tierSpec{
		code: "kampus", name: "Kampus",
		batches: []batchSpec{
			{name: "Batch 1 — Kalkulus, Aljabar, Analisis & PD", challenges: kampusMateriChallenges(b1, "Ujian Batch 1 Kampus")},
			{name: "Batch 2 — Diskrit, Statistika, Geometri & Terapan", challenges: kampusMateriChallenges(b2, "Ujian Batch 2 Kampus")},
			{name: "Batch 3 — Campuran Dasar", challenges: kampusCampuranChallenges("Campuran Dasar Level %d", 5, 20, 35, kampusSecondsLevel, "Ujian Campuran Dasar Kampus", 20, 35, all)},
			{name: "Batch 4 — Campuran", challenges: kampusCampuranChallenges("Campuran Level %d", 10, 25, 40, kampusSecondsLevel, "Ujian Campuran Kampus", 40, 60, all)},
			{name: "Batch 5 — Campuran Sulit", challenges: kampusCampuranChallenges("Campuran Sulit Level %d", 10, 25, 50, kampusSecondsHard, "Ujian Campuran Sulit Kampus", 40, 75, kampusHardMix())},
		},
	}
}

func kampusMateriBatch1() []materiSpec {
	return []materiSpec{
		{"Kalkulus I", mixOf(genKLimitExp, genKLimitSinCubic, genKImplicitDerivative, genKRelatedRates, genKExtremeClosed)},
		{"Kalkulus II", mixOf(genKIntegralSubstitution, genKPartialFraction, genKImproperIntegral, genKSeriesSum, genKRadiusConvergence, genKByPartsPi)},
		{"Kalkulus III (Multivariabel)", mixOf(genKPartialDerivative, genKDoubleIntegralRect, genKDirectionalDerivative, genKCriticalValue, genKTripleIntegralBox)},
		{"Kalkulus Vektor", mixOf(genKDivergence, genKCurlZ, genKGreenRectangle, genKFluxBox, genKConservativeLine)},
		{"Aljabar Linear", mixOf(genKDet4, genKEigen2, genKRank, genKLinearTransform)},
		{"Aljabar Abstrak", mixOf(genKOrderZn, genKOrderMultiplicative, genKGroupCounting, genKInverseModP, genKPermutationOrder)},
		{"Aljabar Boolean", mixOf(genKBooleanMinterms, genKTwosComplement, genKBaseConversion, genKBooleanCount)},
		{"Analisis Real", mixOf(genKSequenceLimit, genKSupInf, genKTelescoping, genKEpsilonDelta, genKContinuityParam, genKRecursiveLimit)},
		{"Analisis Kompleks", mixOf(genKComplexModulus, genKComplexProduct, genKComplexPower, genKResidueSimple, genKContourIntegral, genKArgument, genKHarmonicParam)},
		{"Analisis Numerik", mixOf(genKNewtonStep, genKBisection, genKTrapezoid, genKSimpson, genKLagrange, genKEulerStep, genKRelativeError)},
		{"PD Biasa", mixOf(genKCharacteristicRoots, genKSeparable, genKLinearFirstOrder, genKExactEquation, genKLaplaceEval, genKSystemEigen)},
		{"PD Parsial", mixOf(genKPDEClassify, genKHeatDecay, genKWaveFreq, genKLaplaceHarmonicPDE, genKSeparationEigen)},
	}
}

func kampusMateriChallenges(materi []materiSpec, examName string) []challengeSpec {
	challenges := make([]challengeSpec, 0, len(materi)+1)
	gens := make([]func(int) question, 0, len(materi))
	for _, m := range materi {
		challenges = append(challenges, buildChallenge(m.name, false, 20, 35, kampusPassPercent, kampusOptionCount, kampusSecondsLevel, m.gen))
		gens = append(gens, m.gen)
	}
	challenges = append(challenges, buildChallenge(examName, true, 30, 50, kampusPassPercent, kampusOptionCount, 30*kampusSecondsLevel, mixOf(gens...)))
	return challenges
}

func kampusCampuranChallenges(levelFmt string, levels, required, bank, secondsPerQuestion int, examName string, examRequired, examBank int, gen func(int) question) []challengeSpec {
	challenges := make([]challengeSpec, 0, levels+1)
	for level := 1; level <= levels; level++ {
		challenges = append(challenges, buildChallenge(fmt.Sprintf(levelFmt, level), false, required, bank, kampusPassPercent, kampusOptionCount, secondsPerQuestion, gen))
	}
	challenges = append(challenges, buildChallenge(examName, true, examRequired, examBank, kampusPassPercent, kampusOptionCount, examRequired*secondsPerQuestion, gen))
	return challenges
}

// ---------- helper ----------

func lcm(a, b int) int { return absInt(a / gcd(a, b) * b) }

func modPow(base, exp, mod int) int {
	result, b := 1%mod, ((base%mod)+mod)%mod
	for exp > 0 {
		if exp&1 == 1 {
			result = result * b % mod
		}
		b = b * b % mod
		exp >>= 1
	}
	return result
}

func modInverse(a, m int) int {
	for x := 1; x < m; x++ {
		if a*x%m == 1 {
			return x
		}
	}
	return -1
}

func eulerPhi(n int) int {
	count := 0
	for k := 1; k <= n; k++ {
		if gcd(n, k) == 1 {
			count++
		}
	}
	return count
}

func divisorCount(n int) int {
	count := 0
	for d := 1; d <= n; d++ {
		if n%d == 0 {
			count++
		}
	}
	return count
}

func divisorSum(n int) int {
	sum := 0
	for d := 1; d <= n; d++ {
		if n%d == 0 {
			sum += d
		}
	}
	return sum
}

// cleanFloat: buang noise floating point (0.30000000000000004 -> 0.3).
func cleanFloat(v float64) float64 { return math.Round(v*1e9) / 1e9 }

func matMulInt(a, b [][]int) [][]int {
	out := make([][]int, len(a))
	for i := range a {
		out[i] = make([]int, len(b[0]))
		for j := range b[0] {
			for k := range b {
				out[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	return out
}

func diagInt(vals ...int) [][]int {
	m := make([][]int, len(vals))
	for i, v := range vals {
		m[i] = make([]int, len(vals))
		m[i][i] = v
	}
	return m
}

// unimodular2: P dengan det 1 + inversnya (semua bulat).
func unimodular2() (p, pinv [][]int) {
	m, n := nonZeroBetween(-2, 2), nonZeroBetween(-2, 2)
	return [][]int{{1 + m*n, m}, {n, 1}}, [][]int{{1, -m}, {-n, 1 + m*n}}
}

func det4(m [][]int) int {
	total := 0
	for col := range 4 {
		minor := make([][]int, 0, 3)
		for r := 1; r < 4; r++ {
			row := make([]int, 0, 3)
			for c := range 4 {
				if c != col {
					row = append(row, m[r][c])
				}
			}
			minor = append(minor, row)
		}
		sign := 1
		if col%2 == 1 {
			sign = -1
		}
		total += sign * m[0][col] * det3(minor)
	}
	return total
}

func matrixRank(m [][]int) int {
	rows, cols := len(m), len(m[0])
	a := make([][]float64, rows)
	for i := range m {
		a[i] = make([]float64, cols)
		for j := range m[i] {
			a[i][j] = float64(m[i][j])
		}
	}
	rank := 0
	for c := 0; c < cols && rank < rows; c++ {
		pivot := -1
		for r := rank; r < rows; r++ {
			if math.Abs(a[r][c]) > 1e-9 {
				pivot = r
				break
			}
		}
		if pivot < 0 {
			continue
		}
		a[rank], a[pivot] = a[pivot], a[rank]
		for r := range rows {
			if r != rank {
				f := a[r][c] / a[rank][c]
				for k := c; k < cols; k++ {
					a[r][k] -= f * a[rank][k]
				}
			}
		}
		rank++
	}
	return rank
}

// fmtComplex: "3 - 2i", "-i", "4", "5i".
func fmtComplex(re, im int) string {
	imPart := func(v int) string {
		switch v {
		case 1:
			return "i"
		case -1:
			return "-i"
		}
		return fmt.Sprintf("%di", v)
	}
	switch {
	case im == 0:
		return fmt.Sprintf("%d", re)
	case re == 0:
		return imPart(im)
	case im < 0:
		return fmt.Sprintf("%d - %s", re, strings.TrimPrefix(imPart(-im), "-"))
	}
	return fmt.Sprintf("%d + %s", re, imPart(im))
}

func complexMul(a, b, c, d int) (int, int) { return a*c - b*d, a*d + b*c }

func complexPow(a, b, n int) (int, int) {
	re, im := 1, 0
	for range n {
		re, im = complexMul(re, im, a, b)
	}
	return re, im
}

// ---------- Kalkulus I ----------

func genKLimitExp(optionCount int) question {
	a := nonZeroBetween(-9, 9)
	b := pickOne([]int{1, 2, 4, 5})
	correct := round2(float64(a) / float64(b))
	return question{
		prompt:  fmt.Sprintf("lim (x→0) (e^(%s) - 1) / (%s) = ?", formatPoly(a, 0), formatPoly(b, 0)),
		options: buildOptions(correct, []float64{round2(float64(b) / float64(a)), float64(a), 0, 1, -correct}, optionCount, true),
	}
}

func genKLimitSinCubic(optionCount int) question {
	// sin(ax) - ax ≈ -(ax)³/6  ->  limit = -a³/(6b)
	a := pickOne([]int{3, 6})
	b := between(1, 3)
	correct := -float64(a*a*a) / float64(6*b)
	denom := "x³"
	if b != 1 {
		denom = fmt.Sprintf("%dx³", b)
	}
	return question{
		prompt:  fmt.Sprintf("lim (x→0) (sin(%dx) - %dx) / (%s) = ?", a, a, denom),
		options: buildOptions(correct, []float64{-correct, round2(-float64(a) / float64(6*b)), 0, 2 * correct}, optionCount, true),
	}
}

func genKImplicitDerivative(optionCount int) question {
	for {
		a, b := between(1, 3), between(-3, 3)
		x0, y0 := between(-4, 4), between(-4, 4)
		num, den := -(2*a*x0 + b*y0), b*x0+2*y0 // ax² + bxy + y² = c
		if den == 0 || num == 0 || (x0 == 0 && y0 == 0) {
			continue
		}
		val := float64(num) / float64(den)
		if !isExact2(val) {
			continue
		}
		c := a*x0*x0 + b*x0*y0 + y0*y0
		return question{
			prompt:  fmt.Sprintf("Kurva %s = %d melalui titik (%d, %d). Nilai dy/dx di titik tersebut adalah?", formatTerms([]int{a, b, 1}, []string{"x²", "xy", "y²"}), c, x0, y0),
			options: buildOptions(round2(val), []float64{-round2(val), round2(float64(den) / float64(num)), 0, round2(val + 1)}, optionCount, true),
		}
	}
}

func genKRelatedRates(optionCount int) question {
	r := between(2, 15)
	v := between(1, 5)
	switch rand.IntN(3) {
	case 0:
		correct := float64(2 * r * v)
		return question{
			prompt:  fmt.Sprintf("Jari-jari sebuah lingkaran bertambah %d cm/detik. Saat jari-jarinya %d cm, laju pertambahan luasnya adalah kπ cm²/detik. Nilai k = ?", v, r),
			options: buildOptions(correct, []float64{float64(r * v), float64(r * r * v), float64(2 * v), float64(r * r)}, optionCount, false),
		}
	case 1:
		correct := float64(4 * r * r * v)
		return question{
			prompt:  fmt.Sprintf("Jari-jari sebuah bola bertambah %d cm/detik. Saat jari-jarinya %d cm, laju pertambahan volumenya adalah kπ cm³/detik. Nilai k = ?", v, r),
			options: buildOptions(correct, []float64{float64(4 * r * v), float64(2 * r * r * v), round2(4 * float64(r*r*r*v) / 3), float64(r * r * v)}, optionCount, false),
		}
	default:
		correct := float64(3 * r * r * v)
		return question{
			prompt:  fmt.Sprintf("Rusuk sebuah kubus bertambah %d cm/detik. Saat rusuknya %d cm, laju pertambahan volume kubus adalah? (cm³/detik)", v, r),
			options: buildOptions(correct, []float64{float64(r * r * r * v), float64(3 * r * v), float64(r * r * v), correct + float64(v)}, optionCount, false),
		}
	}
}

func genKExtremeClosed(optionCount int) question {
	// f(x) = x³ - 3p²x + c di [0, 2p]: min f(p) = c - 2p³, max f(2p) = c + 2p³
	p := between(1, 4)
	c := between(-10, 20)
	f := formatPoly(1, 0, -3*p*p, c)
	if rand.IntN(2) == 0 {
		correct := float64(c - 2*p*p*p)
		return question{
			prompt:  fmt.Sprintf("Nilai minimum f(x) = %s pada selang [0, %d] adalah?", f, 2*p),
			options: buildOptions(correct, []float64{float64(c), float64(c + 2*p*p*p), float64(-2 * p * p * p), float64(p)}, optionCount, true),
		}
	}
	correct := float64(c + 2*p*p*p)
	return question{
		prompt:  fmt.Sprintf("Nilai maksimum f(x) = %s pada selang [0, %d] adalah?", f, 2*p),
		options: buildOptions(correct, []float64{float64(c), float64(c - 2*p*p*p), float64(2 * p * p * p), float64(2 * p)}, optionCount, true),
	}
}

// ---------- Kalkulus II ----------

func genKIntegralSubstitution(optionCount int) question {
	for {
		n, c, b := between(1, 2), between(0, 3), between(1, 3)
		num := intPow(b*b+c, n+1) - intPow(c, n+1)
		val := float64(num) / float64(n+1)
		if !isExact2(val) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("∫ dari 0 sampai %d 2x(%s)^%d dx = ?", b, formatPoly(1, 0, c), n),
			options: buildOptions(round2(val), []float64{float64(num), round2(float64(intPow(b*b+c, n+1)) / float64(n+1)), round2(val + 1), round2(2 * val)}, optionCount, false),
		}
	}
}

func genKPartialFraction(optionCount int) question {
	for {
		a, b := between(-5, 5), between(-5, 5)
		p, q := between(-6, 6), between(-9, 9)
		if a == b || (p == 0 && q == 0) {
			continue
		}
		A := float64(p*a+q) / float64(a-b)
		B := float64(p*b+q) / float64(b-a)
		if !isExact2(A) || !isExact2(B) {
			continue
		}
		target, correct, other := "A", round2(A), round2(B)
		if rand.IntN(2) == 0 {
			target, correct, other = "B", round2(B), round2(A)
		}
		fa, fb := formatPoly(1, -a), formatPoly(1, -b)
		return question{
			prompt:  fmt.Sprintf("Jika (%s) / ((%s)(%s)) = A/(%s) + B/(%s), nilai %s = ?", formatPoly(p, q), fa, fb, fa, fb, target),
			options: buildOptions(correct, []float64{other, -correct, float64(p), round2(correct + 1)}, optionCount, true),
		}
	}
}

func genKImproperIntegral(optionCount int) question {
	for {
		if rand.IntN(2) == 0 {
			p, k := between(2, 6), between(1, 12)
			val := float64(k) / float64(p-1)
			if !isExact2(val) {
				continue
			}
			return question{
				prompt:  fmt.Sprintf("∫ dari 1 sampai ∞ %d / x^%d dx = ?", k, p),
				options: buildOptions(round2(val), []float64{round2(float64(k) / float64(p)), float64(k), 0, round2(val + 1)}, optionCount, false),
			}
		}
		a, k := between(1, 8), between(1, 12)
		val := float64(k) / float64(a)
		if !isExact2(val) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("∫ dari 0 sampai ∞ %de^(%s) dx = ?", k, formatPoly(-a, 0)),
			options: buildOptions(round2(val), []float64{float64(k * a), round2(float64(a) / float64(k)), 0, round2(val + 1)}, optionCount, false),
		}
	}
}

func genKSeriesSum(optionCount int) question {
	switch rand.IntN(3) {
	case 0:
		c := between(1, 20)
		correct := float64(c)
		return question{
			prompt:  fmt.Sprintf("Σ (n=1 sampai ∞) %d / (n(n + 1)) = ?", c),
			options: buildOptions(correct, []float64{round2(correct / 2), 2 * correct, correct + 1, 0}, optionCount, false),
		}
	case 1:
		c := 4 * between(1, 5)
		correct := float64(3 * c / 4)
		return question{
			prompt:  fmt.Sprintf("Σ (n=1 sampai ∞) %d / (n(n + 2)) = ?", c),
			options: buildOptions(correct, []float64{float64(c / 2), float64(c), float64(3 * c / 2), correct + 1}, optionCount, false),
		}
	default:
		for {
			a, k := between(1, 12), between(2, 5)
			val := float64(a*k) / float64(k-1)
			if !isExact2(val) {
				continue
			}
			return question{
				prompt:  fmt.Sprintf("Σ (n=0 sampai ∞) %d / %d^n = ?", a, k),
				options: buildOptions(round2(val), []float64{round2(float64(a) / float64(k-1)), float64(a * k), float64(a), round2(val + 1)}, optionCount, false),
			}
		}
	}
}

func genKRadiusConvergence(optionCount int) question {
	k := between(2, 9)
	a := between(-5, 5)
	base := "x"
	if a != 0 {
		base = "(x" + linTerm(-a, "") + ")"
	}
	if rand.IntN(2) == 0 {
		correct := float64(k)
		return question{
			prompt:  fmt.Sprintf("Jari-jari kekonvergenan deret Σ %s^n / (%d^n · n) adalah?", base, k),
			options: buildOptions(correct, []float64{round2(1 / float64(k)), float64(k * k), float64(2 * k), float64(absInt(a) + k)}, optionCount, false),
		}
	}
	correct := float64(2 * k)
	return question{
		prompt:  fmt.Sprintf("Panjang selang kekonvergenan deret Σ n%s^n / %d^n adalah?", base, k),
		options: buildOptions(correct, []float64{float64(k), float64(k * k), round2(2 / float64(k)), correct + 1}, optionCount, false),
	}
}

func genKByPartsPi(optionCount int) question {
	// ∫₀^{mπ} a·x·sin x dx = a·m·π·(-1)^{m+1}
	a := nonZeroBetween(-5, 6)
	m := between(1, 3)
	sign := 1
	if m%2 == 0 {
		sign = -1
	}
	correct := float64(a * m * sign)
	upper := "π"
	if m != 1 {
		upper = fmt.Sprintf("%dπ", m)
	}
	return question{
		prompt:  fmt.Sprintf("∫ dari 0 sampai %s %s sin(x) dx = kπ. Nilai k = ?", upper, formatPoly(a, 0)),
		options: buildOptions(correct, []float64{-correct, float64(a), float64(2 * a * m), 0}, optionCount, true),
	}
}

// ---------- Kalkulus III ----------

func genKPartialDerivative(optionCount int) question {
	a, b := nonZeroBetween(-4, 4), nonZeroBetween(-3, 3)
	c, d := between(-6, 6), between(-6, 6)
	x0, y0 := between(-3, 3), between(-3, 3)
	fx := 2*a*x0*y0 + b*y0*y0*y0 + d
	fy := a*x0*x0 + 3*b*x0*y0*y0 + c
	f := formatTerms([]int{a, b, c, d}, []string{"x²y", "xy³", "y", "x"})
	wrt, correct, other := "x", float64(fx), float64(fy)
	if rand.IntN(2) == 0 {
		wrt, correct, other = "y", float64(fy), float64(fx)
	}
	return question{
		prompt:  fmt.Sprintf("Jika f(x, y) = %s, nilai ∂f/∂%s di titik (%d, %d) adalah?", f, wrt, x0, y0),
		options: buildOptions(correct, []float64{other, correct + other, correct + 1, -correct}, optionCount, true),
	}
}

func genKDoubleIntegralRect(optionCount int) question {
	p, q := nonZeroBetween(-5, 6), nonZeroBetween(-5, 6)
	a, b := between(1, 4), between(1, 4)
	val := func(a, b int) float64 { return float64(p*a*a*b)/2 + float64(q*a*b*b)/2 }
	correct := val(a, b)
	return question{
		prompt:  fmt.Sprintf("∫ (x: 0→%d) ∫ (y: 0→%d) (%s) dy dx = ?", a, b, formatTerms([]int{p, q}, []string{"x", "y"})),
		options: buildOptions(correct, []float64{val(b, a), float64(p*a*b + q*a*b), correct + 1, 2 * correct}, optionCount, true),
	}
}

func genKDirectionalDerivative(optionCount int) question {
	for {
		a, b, c := nonZeroBetween(-3, 3), nonZeroBetween(-3, 3), between(-3, 3)
		x0, y0 := between(-3, 3), between(-3, 3)
		gx, gy := 2*a*x0+c*y0, 2*b*y0+c*x0
		t := pickOne(pyTriples)
		ux, uy := t.a*randSign(), t.b*randSign()
		if rand.IntN(2) == 0 {
			ux, uy = uy, ux
		}
		raw := gx*ux + gy*uy
		val := float64(raw) / float64(t.c)
		if raw == 0 || !isExact2(val) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Turunan berarah f(x, y) = %s di titik (%d, %d) dalam arah vektor (%d, %d) adalah?", formatTerms([]int{a, b, c}, []string{"x²", "y²", "xy"}), x0, y0, ux, uy),
			options: buildOptions(round2(val), []float64{float64(raw), float64(gx + gy), -round2(val), round2(math.Hypot(float64(gx), float64(gy)))}, optionCount, true),
		}
	}
}

func genKCriticalValue(optionCount int) question {
	if rand.IntN(2) == 0 {
		a, b, c := 2*between(-4, 4), 2*between(-4, 4), between(-10, 10)
		correct := float64(c - a*a/4 - b*b/4)
		return question{
			prompt:  fmt.Sprintf("Nilai minimum f(x, y) = %s adalah?", formatTerms([]int{1, 1, a, b, c}, []string{"x²", "y²", "x", "y", ""})),
			options: buildOptions(correct, []float64{float64(c), float64(c + a*a/4 + b*b/4), float64(-a / 2), correct - 1}, optionCount, true),
		}
	}
	k := between(-5, 5)
	correct := float64(4 - k*k)
	return question{
		prompt:  fmt.Sprintf("Untuk f(x, y) = %s, nilai determinan Hessian D = f_xx·f_yy - (f_xy)² di titik kritis (0, 0) adalah?", formatTerms([]int{1, k, 1}, []string{"x²", "xy", "y²"})),
		options: buildOptions(correct, []float64{float64(4 + k*k), -correct, float64(k * k), 4}, optionCount, true),
	}
}

func genKTripleIntegralBox(optionCount int) question {
	a, b, c := between(1, 4), between(1, 4), between(1, 4)
	abc := float64(a * b * c)
	correct := abc * float64(a+b+c) / 2
	return question{
		prompt:  fmt.Sprintf("∭ (x + y + z) dV pada balok 0 ≤ x ≤ %d, 0 ≤ y ≤ %d, 0 ≤ z ≤ %d adalah?", a, b, c),
		options: buildOptions(correct, []float64{abc, abc * float64(a+b+c), float64(a+b+c) / 2, correct + 1}, optionCount, false),
	}
}

// ---------- Kalkulus Vektor ----------

func genKDivergence(optionCount int) question {
	a, b, c := nonZeroBetween(-4, 4), nonZeroBetween(-4, 4), nonZeroBetween(-4, 4)
	x, y, z := between(-3, 3), between(-3, 3), between(-3, 3)
	correct := float64(2*a*x*y + b*z + 2*c*x*z)
	sumF := float64(a*x*x*y + b*y*z + c*x*z*z)
	return question{
		prompt: fmt.Sprintf(
			"Divergensi medan vektor F = (%s, %s, %s) di titik (%d, %d, %d) adalah?",
			formatTerms([]int{a}, []string{"x²y"}), formatTerms([]int{b}, []string{"yz"}), formatTerms([]int{c}, []string{"xz²"}), x, y, z,
		),
		options: buildOptions(correct, []float64{sumF, -correct, correct + 1, float64(a + b + c)}, optionCount, true),
	}
}

func genKCurlZ(optionCount int) question {
	a, b, c, d, e := nonZeroBetween(-4, 4), between(-3, 3), nonZeroBetween(-3, 3), between(-3, 3), nonZeroBetween(-3, 3)
	x, y, z := between(-3, 3), between(-3, 3), between(-3, 3)
	qx := 2*c*x + d*y
	py := a*x + 2*b*y
	correct := float64(qx - py)
	return question{
		prompt: fmt.Sprintf(
			"F = (%s, %s, %s). Komponen z dari curl F di titik (%d, %d, %d) adalah?",
			formatTerms([]int{a, b}, []string{"xy", "y²"}), formatTerms([]int{c, d}, []string{"x²", "xy"}), formatTerms([]int{e}, []string{"z"}), x, y, z,
		),
		options: buildOptions(correct, []float64{-correct, float64(qx), float64(qx + py), correct + 1}, optionCount, true),
	}
}

func genKGreenRectangle(optionCount int) question {
	m, n := between(-4, 4), between(-4, 4)
	p, q := nonZeroBetween(-5, 5), nonZeroBetween(-5, 5)
	for p+q == 0 {
		q = nonZeroBetween(-5, 5)
	}
	a, b := between(1, 5), between(1, 5)
	// P = mx - py, Q = qx + ny  ->  Q_x - P_y = q + p
	correct := float64((q + p) * a * b)
	return question{
		prompt: fmt.Sprintf(
			"Dengan teorema Green, nilai ∮ (%s) dx + (%s) dy mengelilingi persegi panjang 0 ≤ x ≤ %d, 0 ≤ y ≤ %d berlawanan arah jarum jam adalah?",
			formatTerms([]int{m, -p}, []string{"x", "y"}), formatTerms([]int{q, n}, []string{"x", "y"}), a, b,
		),
		options: buildOptions(correct, []float64{float64((q - p) * a * b), -correct, float64(a * b), float64((q + p) * (a + b))}, optionCount, true),
	}
}

func genKFluxBox(optionCount int) question {
	a, b, c := nonZeroBetween(-3, 4), nonZeroBetween(-3, 4), nonZeroBetween(-3, 4)
	for a+b+c == 0 {
		c = nonZeroBetween(-3, 4)
	}
	p, q, r := between(1, 4), between(1, 4), between(1, 4)
	vol := p * q * r
	correct := float64((a + b + c) * vol)
	return question{
		prompt: fmt.Sprintf(
			"Dengan teorema divergensi, fluks F = (%s, %s, %s) yang keluar dari permukaan balok 0 ≤ x ≤ %d, 0 ≤ y ≤ %d, 0 ≤ z ≤ %d adalah?",
			formatTerms([]int{a}, []string{"x"}), formatTerms([]int{b}, []string{"y"}), formatTerms([]int{c}, []string{"z"}), p, q, r,
		),
		options: buildOptions(correct, []float64{float64(a + b + c), float64(a * b * c * vol), float64(vol), correct + 1}, optionCount, true),
	}
}

func genKConservativeLine(optionCount int) question {
	// φ = a x²y + b yz + c z  ->  F = ∇φ = (2axy, ax² + bz, by + c)
	a, b, c := nonZeroBetween(-3, 3), nonZeroBetween(-3, 3), between(-5, 5)
	phi := func(x, y, z int) int { return a*x*x*y + b*y*z + c*z }
	A := []int{between(-2, 3), between(-2, 3), between(-2, 3)}
	B := []int{between(-2, 3), between(-2, 3), between(-2, 3)}
	correct := float64(phi(B[0], B[1], B[2]) - phi(A[0], A[1], A[2]))
	return question{
		prompt: fmt.Sprintf(
			"Medan F = (%s, %s, %s) konservatif. Nilai ∫ F · dr sepanjang lintasan dari (%s) ke (%s) adalah?",
			formatTerms([]int{2 * a}, []string{"xy"}), formatTerms([]int{a, b}, []string{"x²", "z"}), formatTerms([]int{b, c}, []string{"y", ""}),
			joinInts(A), joinInts(B),
		),
		options: buildOptions(correct, []float64{-correct, float64(phi(B[0], B[1], B[2])), correct + 1, correct - 1}, optionCount, true),
	}
}

// ---------- Aljabar Linear ----------

func genKDet4(optionCount int) question {
	m := make([][]int, 4)
	for i := range m {
		m[i] = []int{between(-2, 3), between(-2, 3), between(-2, 3), between(-2, 3)}
	}
	correct := float64(det4(m))
	diag := float64(m[0][0] * m[1][1] * m[2][2] * m[3][3])
	return question{
		prompt:  fmt.Sprintf("Determinan matriks A = %s (baris dipisah titik koma) adalah?", formatMatrix(m)),
		options: buildOptions(correct, []float64{-correct, diag, correct + 1, correct - 1}, optionCount, true),
	}
}

func genKEigen2(optionCount int) question {
	l1 := between(-5, 7)
	l2 := between(-5, 7)
	for l2 == l1 {
		l2 = between(-5, 7)
	}
	p, pinv := unimodular2()
	a := matMulInt(matMulInt(p, diagInt(l1, l2)), pinv)
	hi, lo := max(l1, l2), min(l1, l2)
	type variant struct {
		label   string
		correct int
	}
	v := pickOne([]variant{{"Nilai eigen terbesar", hi}, {"Nilai eigen terkecil", lo}, {"Hasil kali semua nilai eigen", l1 * l2}, {"Jumlah semua nilai eigen", l1 + l2}})
	correct := float64(v.correct)
	return question{
		prompt:  fmt.Sprintf("Diketahui A = %s (baris dipisah titik koma). %s dari A adalah?", formatMatrix(a), v.label),
		options: buildOptions(correct, []float64{float64(hi), float64(lo), float64(l1 * l2), float64(l1 + l2), -correct, correct + 1}, optionCount, true),
	}
}

func genKRank(optionCount int) question {
	for {
		target := between(1, 3)
		var base [][]int
		for len(base) < target {
			row := []int{between(-3, 3), between(-3, 3), between(-3, 3), between(-3, 3)}
			if matrixRank(append(append([][]int{}, base...), row)) == len(base)+1 {
				base = append(base, row)
			}
		}
		rows := append([][]int{}, base...)
		for len(rows) < 3 {
			c1, c2 := between(-2, 2), between(-2, 2)
			i, j := rand.IntN(len(base)), rand.IntN(len(base))
			row := make([]int, 4)
			for k := range row {
				row[k] = c1*base[i][k] + c2*base[j][k]
			}
			rows = append(rows, row)
		}
		rand.Shuffle(3, func(i, j int) { rows[i], rows[j] = rows[j], rows[i] })
		if matrixRank(rows) != target {
			continue
		}
		correct := float64(target)
		if rand.IntN(2) == 0 {
			return question{
				prompt:  fmt.Sprintf("Rank matriks A = %s (baris dipisah titik koma) adalah?", formatMatrix(rows)),
				options: buildOptions(correct, []float64{0, 1, 2, 3, 4}, optionCount, false),
			}
		}
		return question{
			prompt:  fmt.Sprintf("Dimensi ruang nol (nullity) matriks A = %s (berukuran 3×4, baris dipisah titik koma) adalah?", formatMatrix(rows)),
			options: buildOptions(float64(4-target), []float64{0, 1, 2, 3, 4}, optionCount, false),
		}
	}
}

func genKLinearTransform(optionCount int) question {
	e1 := []int{between(-5, 5), between(-5, 5)} // T(1, 0)
	e2 := []int{between(-5, 5), between(-5, 5)} // T(0, 1)
	u := []int{e1[0] + e2[0], e1[1] + e2[1]}    // T(1, 1)
	v := []int{e1[0] - e2[0], e1[1] - e2[1]}    // T(1, -1)
	x, y := nonZeroBetween(-4, 4), nonZeroBetween(-4, 4)
	i := rand.IntN(2)
	correct := float64(x*e1[i] + y*e2[i])
	return question{
		prompt: fmt.Sprintf(
			"Transformasi linear T: R² → R² memenuhi T(1, 1) = (%d, %d) dan T(1, -1) = (%d, %d). Komponen %s dari T(%d, %d) adalah?",
			u[0], u[1], v[0], v[1], []string{"pertama", "kedua"}[i], x, y,
		),
		options: buildOptions(correct, []float64{float64(x*e1[1-i] + y*e2[1-i]), float64(x*u[i] + y*v[i]), -correct, correct + 1}, optionCount, true),
	}
}

// ---------- Aljabar Abstrak ----------

func genKOrderZn(optionCount int) question {
	n := between(6, 30)
	k := between(1, n-1)
	correct := float64(n / gcd(n, k))
	return question{
		prompt:  fmt.Sprintf("Orde unsur %d di grup (Z_%d, +) adalah?", k, n),
		options: buildOptions(correct, []float64{float64(n), float64(gcd(n, k)), float64(k), float64(n - k)}, optionCount, false),
	}
}

func genKOrderMultiplicative(optionCount int) question {
	p := pickOne([]int{7, 11, 13, 17, 19, 23})
	a := between(2, p-1)
	order := 1
	for x := a % p; x != 1; x = x * a % p {
		order++
	}
	correct := float64(order)
	return question{
		prompt:  fmt.Sprintf("Orde unsur %d di grup perkalian Z_%d* adalah?", a, p),
		options: buildOptions(correct, []float64{float64(p - 1), float64((p - 1) / 2), float64(a), correct + 1}, optionCount, false),
	}
}

func genKGroupCounting(optionCount int) question {
	n := between(6, 40)
	switch rand.IntN(3) {
	case 0:
		return question{
			prompt:  fmt.Sprintf("Banyak generator grup siklik Z_%d adalah?", n),
			options: buildOptions(float64(eulerPhi(n)), []float64{float64(n), float64(n - 1), float64(divisorCount(n))}, optionCount, false),
		}
	case 1:
		return question{
			prompt:  fmt.Sprintf("Banyak subgrup dari grup siklik Z_%d adalah?", n),
			options: buildOptions(float64(divisorCount(n)), []float64{float64(eulerPhi(n)), float64(n), float64(divisorSum(n))}, optionCount, false),
		}
	default:
		k := between(3, 6)
		return question{
			prompt:  fmt.Sprintf("Orde grup simetri S_%d adalah?", k),
			options: buildOptions(float64(factorial(k)), []float64{float64(k), float64(k * k), float64(factorial(k - 1)), float64(factorial(k) / 2)}, optionCount, false),
		}
	}
}

func genKInverseModP(optionCount int) question {
	p := pickOne([]int{7, 11, 13, 17, 19, 23, 29, 31})
	a := between(2, p-1)
	inv := modInverse(a, p)
	return question{
		prompt:  fmt.Sprintf("Invers perkalian %d di field Z_%d adalah?", a, p),
		options: buildOptions(float64(inv), []float64{float64(p - inv), float64(a), float64(p - a), float64(inv%(p-1) + 1)}, optionCount, false),
	}
}

func genKPermutationOrder(optionCount int) question {
	for {
		n := between(5, 9)
		elems := rand.Perm(n)
		var cycles []string
		var lengths []int
		for i := 0; i < n; {
			l := between(1, min(4, n-i))
			if l > 1 {
				parts := make([]string, l)
				for j := range l {
					parts[j] = fmt.Sprintf("%d", elems[i+j]+1)
				}
				cycles = append(cycles, "("+strings.Join(parts, " ")+")")
				lengths = append(lengths, l)
			}
			i += l
		}
		if len(lengths) == 0 {
			continue
		}
		order, sum, prod, longest := 1, 0, 1, 0
		for _, l := range lengths {
			order = lcm(order, l)
			sum += l
			prod *= l
			longest = max(longest, l)
		}
		return question{
			prompt:  fmt.Sprintf("Orde permutasi σ = %s di S_%d adalah?", strings.Join(cycles, ""), n),
			options: buildOptions(float64(order), []float64{float64(sum), float64(prod), float64(longest), float64(2 * order)}, optionCount, false),
		}
	}
}

// ---------- Aljabar Boolean ----------

func boolLiteral(idx int) logicExpr {
	name := string("ABCD"[idx])
	if rand.IntN(3) == 0 {
		return logicExpr{name + "'", func(v []bool) bool { return !v[idx] }}
	}
	return logicExpr{name, func(v []bool) bool { return v[idx] }}
}

func combineBool(l, r logicExpr) logicExpr {
	switch rand.IntN(3) {
	case 0:
		return logicExpr{l.text + "·" + r.text, func(v []bool) bool { return l.eval(v) && r.eval(v) }}
	case 1:
		return logicExpr{l.text + " + " + r.text, func(v []bool) bool { return l.eval(v) || r.eval(v) }}
	default:
		return logicExpr{l.text + " ⊕ " + r.text, func(v []bool) bool { return l.eval(v) != r.eval(v) }}
	}
}

func countTrueRows(expr logicExpr, vars int) int {
	count := 0
	for mask := range 1 << vars {
		vals := make([]bool, vars)
		for i := range vars {
			vals[i] = mask&(1<<i) != 0
		}
		if expr.eval(vals) {
			count++
		}
	}
	return count
}

func genKBooleanMinterms(optionCount int) question {
	var expr logicExpr
	if rand.IntN(2) == 0 {
		expr = combineBool(parenthesize(combineBool(boolLiteral(0), boolLiteral(1))), boolLiteral(2))
	} else {
		expr = combineBool(boolLiteral(0), parenthesize(combineBool(boolLiteral(1), boolLiteral(2))))
	}
	correct := countTrueRows(expr, 3)
	return question{
		prompt:  fmt.Sprintf("Banyak minterm (kombinasi input yang menghasilkan 1) dari fungsi Boolean F(A, B, C) = %s adalah?", expr.text),
		options: buildOptions(float64(correct), []float64{float64(8 - correct), float64(correct + 1), float64(correct - 1), 8}, optionCount, false),
	}
}

func genKTwosComplement(optionCount int) question {
	v := between(-128, 127)
	for v == 0 {
		v = between(-128, 127)
	}
	bits := fmt.Sprintf("%08b", uint8(int8(v)))
	unsigned := float64(uint8(int8(v)))
	return question{
		prompt:  fmt.Sprintf("Bilangan biner 8-bit %s dalam representasi komplemen dua bernilai (desimal)?", bits),
		options: buildOptions(float64(v), []float64{unsigned, float64(-v), float64(v + 1), float64(v - 1)}, optionCount, true),
	}
}

func genKBaseConversion(optionCount int) question {
	switch rand.IntN(3) {
	case 0:
		v := between(64, 1023)
		return question{
			prompt:  fmt.Sprintf("Nilai desimal dari bilangan biner %b adalah?", v),
			options: buildOptions(float64(v), []float64{float64(v + 1), float64(v - 1), float64(v + 2), float64(v ^ 1<<2)}, optionCount, false),
		}
	case 1:
		v := between(256, 4095)
		return question{
			prompt:  fmt.Sprintf("Nilai desimal dari bilangan heksadesimal %X adalah?", v),
			options: buildOptions(float64(v), []float64{float64(v + 16), float64(v - 16), float64(v + 1), float64(v + 256)}, optionCount, false),
		}
	default:
		v := between(64, 511)
		return question{
			prompt:  fmt.Sprintf("Nilai desimal dari bilangan oktal %o adalah?", v),
			options: buildOptions(float64(v), []float64{float64(v + 8), float64(v - 8), float64(v + 1), float64(v + 64)}, optionCount, false),
		}
	}
}

func genKBooleanCount(optionCount int) question {
	n := between(2, 3)
	if rand.IntN(2) == 0 {
		correct := float64(intPow(2, intPow(2, n)))
		return question{
			prompt:  fmt.Sprintf("Banyak fungsi Boolean berbeda dengan %d variabel input adalah?", n),
			options: buildOptions(correct, []float64{float64(intPow(2, n)), float64(intPow(2, n) * 2), float64(n * n), correct / 2}, optionCount, false),
		}
	}
	k := between(1, intPow(2, n)-1)
	correct := float64(intPow(2, n) - k)
	return question{
		prompt:  fmt.Sprintf("Fungsi Boolean %d variabel memiliki %d minterm. Banyak maxterm-nya adalah?", n, k),
		options: buildOptions(correct, []float64{float64(k), float64(intPow(2, n)), float64(intPow(2, n) + k), correct + 1}, optionCount, false),
	}
}

// ---------- Analisis Real ----------

func genKSequenceLimit(optionCount int) question {
	if rand.IntN(4) == 0 {
		a, b := nonZeroBetween(-9, 9), between(-9, 9)
		d, e := between(1, 5), between(1, 9)
		return question{
			prompt:  fmt.Sprintf("lim (n→∞) (%s) / (%s) = ?", formatPolyVar("n", a, b), formatPolyVar("n", d, 0, e)),
			options: buildOptions(0, []float64{round2(float64(a) / float64(d)), float64(a), 1, -1}, optionCount, true),
		}
	}
	a, d := nonZeroBetween(-9, 9), pickOne([]int{1, 2, 4, 5}) // a/d pasti desimal pas
	b, c, e := between(-9, 9), between(-9, 9), between(-9, 9)
	val := float64(a) / float64(d)
	return question{
		prompt:  fmt.Sprintf("lim (n→∞) (%s) / (%s) = ?", formatPolyVar("n", a, b, c), formatPolyVar("n", d, 0, e)),
		options: buildOptions(round2(val), []float64{round2(float64(d) / float64(a)), float64(a), 0, round2(val + 1)}, optionCount, true),
	}
}

func genKSupInf(optionCount int) question {
	a := between(-5, 5)
	b := 2 * between(1, 5)
	sup, inf := float64(a)+float64(b)/2, float64(a-b)
	set := fmt.Sprintf("%d(-1)^n / n", b)
	if a != 0 {
		set = fmt.Sprintf("%d + %s", a, set)
	}
	if rand.IntN(2) == 0 {
		return question{
			prompt:  fmt.Sprintf("Diketahui S = {%s : n ∈ ℕ}. Nilai sup S adalah?", set),
			options: buildOptions(sup, []float64{inf, float64(a), float64(a + b), sup + 1}, optionCount, true),
		}
	}
	return question{
		prompt:  fmt.Sprintf("Diketahui S = {%s : n ∈ ℕ}. Nilai inf S adalah?", set),
		options: buildOptions(inf, []float64{sup, float64(a), float64(a) - float64(b)/2, inf - 1}, optionCount, true),
	}
}

func genKTelescoping(optionCount int) question {
	for {
		k := pickOne([]int{0, 1, 3, 4, 9})
		c := between(1, 10)
		val := float64(c) / float64(k+1)
		if !isExact2(val) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Σ (n=1 sampai ∞) %d / ((%s)(%s)) = ?", c, formatPolyVar("n", 1, k), formatPolyVar("n", 1, k+1)),
			options: buildOptions(round2(val), []float64{round2(float64(c) / float64(k+2)), float64(c), round2(val + 1), 0}, optionCount, false),
		}
	}
}

func genKEpsilonDelta(optionCount int) question {
	m := nonZeroBetween(-9, 9)
	for absInt(m) == 1 {
		m = nonZeroBetween(-9, 9)
	}
	delta := pickOne([]float64{0.05, 0.1, 0.25, 0.5})
	eps := round2(float64(absInt(m)) * delta)
	a, b := between(-5, 5), between(-9, 9)
	return question{
		prompt: fmt.Sprintf(
			"Pada bukti ε-δ untuk lim (x→%d) (%s) = %d, jika ε = %s maka nilai δ terbesar yang menjamin |f(x) - L| < ε adalah?",
			a, formatPoly(m, b), m*a+b, fmtNum(eps),
		),
		options: buildOptions(delta, []float64{eps, round2(eps * float64(absInt(m))), round2(delta * 2), round2(delta / 2)}, optionCount, false),
	}
}

func genKContinuityParam(optionCount int) question {
	a := nonZeroBetween(-5, 5)
	b, c := between(-9, 9), between(-4, 4)
	correct := float64(c*c + b - a*c)
	return question{
		prompt: fmt.Sprintf(
			"Fungsi f(x) = %s + k untuk x < %d dan f(x) = %s untuk x ≥ %d kontinu di x = %d. Nilai k = ?",
			formatPoly(a, 0), c, formatPoly(1, 0, b), c, c,
		),
		options: buildOptions(correct, []float64{float64(c*c + b), -correct, float64(c*c + b + a*c), correct + 1}, optionCount, true),
	}
}

func genKRecursiveLimit(optionCount int) question {
	if rand.IntN(2) == 0 {
		a := pickOne([]int{2, 6, 12, 20, 30, 42})
		l := (1 + int(math.Round(math.Sqrt(float64(1+4*a))))) / 2
		correct := float64(l)
		return question{
			prompt:  fmt.Sprintf("Barisan x₁ = √%d dan x_(n+1) = √(%d + x_n). Nilai lim (n→∞) x_n adalah?", a, a),
			options: buildOptions(correct, []float64{round2(math.Sqrt(float64(a))), float64(a), correct + 1, correct - 1}, optionCount, false),
		}
	}
	for {
		r, c := between(2, 5), between(1, 12)
		val := float64(c*r) / float64(r-1)
		if !isExact2(val) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Barisan x₁ = 1 dan x_(n+1) = x_n/%d + %d. Nilai lim (n→∞) x_n adalah?", r, c),
			options: buildOptions(round2(val), []float64{float64(c), float64(c * r), round2(float64(c) / float64(r-1)), round2(val + 1)}, optionCount, false),
		}
	}
}

// ---------- Analisis Kompleks ----------

func genKComplexModulus(optionCount int) question {
	t := pickOne(pyTriples)
	k := between(1, 3)
	a, b := t.a*k*randSign(), t.b*k*randSign()
	if rand.IntN(2) == 0 {
		a, b = b, a
	}
	correct := float64(t.c * k)
	return question{
		prompt:  fmt.Sprintf("Nilai |z| untuk z = %s adalah?", fmtComplex(a, b)),
		options: buildOptions(correct, []float64{float64(absInt(a) + absInt(b)), correct * correct, correct + 1, correct - 1}, optionCount, false),
	}
}

func genKComplexProduct(optionCount int) question {
	a, b, c, d := nonZeroBetween(-6, 6), nonZeroBetween(-6, 6), nonZeroBetween(-6, 6), nonZeroBetween(-6, 6)
	re, im := complexMul(a, b, c, d)
	part, correct, other := "real", float64(re), float64(im)
	if rand.IntN(2) == 0 {
		part, correct, other = "imajiner", float64(im), float64(re)
	}
	return question{
		prompt:  fmt.Sprintf("Bagian %s dari (%s)(%s) adalah?", part, fmtComplex(a, b), fmtComplex(c, d)),
		options: buildOptions(correct, []float64{other, float64(a*c + b*d), -correct, float64(a*d - b*c)}, optionCount, true),
	}
}

func genKComplexPower(optionCount int) question {
	a, b := nonZeroBetween(-3, 3), nonZeroBetween(-3, 3)
	n := between(2, 4)
	re, im := complexPow(a, b, n)
	part, correct, other := "real", float64(re), float64(im)
	if rand.IntN(2) == 0 {
		part, correct, other = "imajiner", float64(im), float64(re)
	}
	return question{
		prompt:  fmt.Sprintf("Bagian %s dari (%s)^%d adalah?", part, fmtComplex(a, b), n),
		options: buildOptions(correct, []float64{other, float64(intPow(a, n)), -correct, correct + 1}, optionCount, true),
	}
}

func genKResidueSimple(optionCount int) question {
	for {
		a, b := between(-5, 5), between(-5, 5)
		p, q := between(-6, 6), between(-9, 9)
		if a == b || (p == 0 && q == 0) {
			continue
		}
		val := float64(p*a+q) / float64(a-b)
		if !isExact2(val) {
			continue
		}
		other := float64(p*b+q) / float64(b-a)
		return question{
			prompt: fmt.Sprintf(
				"Residu f(z) = (%s) / ((%s)(%s)) di z = %d adalah?",
				formatPolyVar("z", p, q), formatPolyVar("z", 1, -a), formatPolyVar("z", 1, -b), a,
			),
			options: buildOptions(round2(val), []float64{round2(other), -round2(val), float64(p*a + q), round2(val + 1)}, optionCount, true),
		}
	}
}

func genKContourIntegral(optionCount int) question {
	for {
		r := between(2, 4)
		a := between(-(r - 1), r-1)
		b := pickOne([]int{r + between(1, 3), -(r + between(1, 3))})
		if rand.IntN(2) == 0 { // dua-duanya di dalam
			b = between(-(r - 1), r-1)
		}
		p, q := between(-6, 6), between(-9, 9)
		if a == b || (p == 0 && q == 0) {
			continue
		}
		resA := float64(p*a+q) / float64(a-b)
		resB := float64(p*b+q) / float64(b-a)
		if !isExact2(resA) || !isExact2(resB) {
			continue
		}
		inside := resA
		if absInt(b) < r {
			inside += resB
		}
		correct := round2(2 * inside)
		return question{
			prompt: fmt.Sprintf(
				"∮ (%s) / ((%s)(%s)) dz sepanjang lingkaran |z| = %d (berlawanan arah jarum jam) bernilai kπi. Nilai k = ?",
				formatPolyVar("z", p, q), formatPolyVar("z", 1, -a), formatPolyVar("z", 1, -b), r,
			),
			options: buildOptions(correct, []float64{round2(2 * (resA + resB)), round2(inside), -correct, 0}, optionCount, true),
		}
	}
}

func genKArgument(optionCount int) question {
	type dir struct{ x, y, angle int }
	d := pickOne([]dir{{1, 0, 0}, {1, 1, 45}, {0, 1, 90}, {-1, 1, 135}, {-1, 0, 180}, {-1, -1, -135}, {0, -1, -90}, {1, -1, -45}})
	k := between(1, 6)
	correct := float64(d.angle)
	wrap := correct
	if wrap < 0 {
		wrap += 360
	}
	return question{
		prompt:  fmt.Sprintf("Argumen utama (dalam derajat, -180° < θ ≤ 180°) dari z = %s adalah?", fmtComplex(d.x*k, d.y*k)),
		options: buildOptions(correct, []float64{wrap, -correct, correct + 90, correct - 90, correct + 180}, optionCount, true),
	}
}

func genKHarmonicParam(optionCount int) question {
	a := nonZeroBetween(-5, 5)
	switch rand.IntN(3) {
	case 0:
		correct := float64(-a)
		return question{
			prompt:  fmt.Sprintf("Nilai k agar u(x, y) = %s + ky² merupakan fungsi harmonik adalah?", formatTerms([]int{a}, []string{"x²"})),
			options: buildOptions(correct, []float64{float64(a), float64(2 * a), 0, correct + 1}, optionCount, true),
		}
	case 1:
		correct := float64(-3 * a)
		return question{
			prompt:  fmt.Sprintf("Nilai k agar u(x, y) = %s + kxy² merupakan fungsi harmonik adalah?", formatTerms([]int{a}, []string{"x³"})),
			options: buildOptions(correct, []float64{float64(3 * a), float64(-a), float64(-6 * a), correct + 1}, optionCount, true),
		}
	default:
		correct := float64(absInt(a))
		return question{
			prompt:  fmt.Sprintf("Nilai k > 0 agar u(x, y) = e^(%s) sin(ky) merupakan fungsi harmonik adalah?", formatPoly(a, 0)),
			options: buildOptions(correct, []float64{float64(a * a), correct + 1, 1, float64(2 * absInt(a))}, optionCount, false),
		}
	}
}

// ---------- Analisis Numerik ----------

func genKNewtonStep(optionCount int) question {
	for {
		x0 := between(1, 8)
		if rand.IntN(2) == 0 {
			a := between(2, 60)
			x1 := cleanFloat(float64(x0*x0+a) / float64(2*x0))
			if !isExact2(x1) || x0*x0 == a {
				continue
			}
			return question{
				prompt:  fmt.Sprintf("Dengan metode Newton-Raphson untuk f(x) = %s dan tebakan awal x₀ = %d, nilai x₁ adalah?", formatPoly(1, 0, -a), x0),
				options: buildOptions(round2(x1), []float64{round2(float64(x0) + float64(x0*x0-a)/float64(2*x0)), float64(x0*x0 - a), float64(x0), round2(x1 + 0.5)}, optionCount, true),
			}
		}
		a := between(2, 200)
		x1 := cleanFloat(float64(x0) - float64(x0*x0*x0-a)/float64(3*x0*x0))
		if !isExact2(x1) || x0*x0*x0 == a {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Dengan metode Newton-Raphson untuk f(x) = %s dan tebakan awal x₀ = %d, nilai x₁ adalah?", formatPoly(1, 0, 0, -a), x0),
			options: buildOptions(round2(x1), []float64{round2(float64(x0) + float64(x0*x0*x0-a)/float64(3*x0*x0)), float64(x0*x0*x0 - a), float64(x0), round2(x1 + 0.5)}, optionCount, true),
		}
	}
}

func genKBisection(optionCount int) question {
	for {
		a := between(-5, 2)
		w := between(1, 8)
		n := between(2, 5)
		val := float64(w) / float64(intPow(2, n))
		if !isExact2(val) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Metode bisection dimulai pada selang [%d, %d]. Lebar selang setelah %d iterasi adalah?", a, a+w, n),
			options: buildOptions(round2(val), []float64{round2(float64(w) / float64(n)), round2(2 * val), round2(val / 2), float64(w)}, optionCount, false),
		}
	}
}

func genKTrapezoid(optionCount int) question {
	a, b, c := between(-3, 3), between(-5, 5), between(-5, 10)
	for a == 0 {
		a = between(-3, 3)
	}
	n := between(2, 4)
	f := func(x int) int { return a*x*x + b*x + c }
	sum := 0.5 * float64(f(0)+f(n))
	all := 0
	for i := 0; i <= n; i++ {
		all += f(i)
		if i > 0 && i < n {
			sum += float64(f(i))
		}
	}
	exact := round2(float64(a*n*n*n)/3 + float64(b*n*n)/2 + float64(c*n))
	return question{
		prompt:  fmt.Sprintf("Aturan trapesium dengan h = 1 untuk ∫ dari 0 sampai %d (%s) dx menghasilkan?", n, formatPoly(a, b, c)),
		options: buildOptions(sum, []float64{exact, float64(all), sum + 1, 2 * sum}, optionCount, true),
	}
}

func genKSimpson(optionCount int) question {
	for {
		a, b, c, d := between(-2, 2), between(-3, 3), between(-5, 5), between(-5, 5)
		h := between(1, 3)
		f := func(x int) int { return a*x*x*x + b*x*x + c*x + d }
		s := cleanFloat(float64(h) / 3 * float64(f(0)+4*f(h)+f(2*h)))
		if !isExact2(s) || (a == 0 && b == 0) {
			continue
		}
		trap := float64(h) / 2 * float64(f(0)+2*f(h)+f(2*h))
		return question{
			prompt:  fmt.Sprintf("Aturan Simpson 1/3 dengan h = %d untuk ∫ dari 0 sampai %d (%s) dx menghasilkan?", h, 2*h, formatPoly(a, b, c, d)),
			options: buildOptions(round2(s), []float64{round2(trap), float64(h * (f(0) + 4*f(h) + f(2*h))), round2(s + 1), round2(s / 2)}, optionCount, true),
		}
	}
}

func genKLagrange(optionCount int) question {
	a, b, c := nonZeroBetween(-3, 3), between(-6, 6), between(-9, 9)
	p := func(x int) int { return a*x*x + b*x + c }
	t := pickOne([]int{3, 4, -1})
	correct := float64(p(t))
	linear := float64(p(2) + (t-2)*(p(2)-p(1)))
	return question{
		prompt:  fmt.Sprintf("Polinom interpolasi berderajat 2 melalui titik (0, %d), (1, %d), dan (2, %d). Nilai polinom tersebut di x = %d adalah?", p(0), p(1), p(2), t),
		options: buildOptions(correct, []float64{linear, float64(p(2)), -correct, correct + 1}, optionCount, true),
	}
}

func genKEulerStep(optionCount int) question {
	for {
		a, b := nonZeroBetween(-3, 3), between(-4, 4)
		y0 := between(1, 6)
		h := pickOne([]float64{0.1, 0.5})
		y1 := cleanFloat(float64(y0) + h*float64(a*y0))
		y2 := cleanFloat(y1 + h*(float64(a)*y1+float64(b)*h))
		if !isExact2(y2) {
			continue
		}
		return question{
			prompt: fmt.Sprintf(
				"Dengan metode Euler (h = %s) untuk y' = %s dan y(0) = %d, nilai hampiran y(%s) adalah?",
				fmtNum(h), formatTerms([]int{a, b}, []string{"y", "x"}), y0, fmtNum(2*h),
			),
			options: buildOptions(round2(y2), []float64{round2(y1), round2(y2 + h), float64(y0), round2(float64(y0) * (1 + float64(a)*2*h))}, optionCount, true),
		}
	}
}

func genKRelativeError(optionCount int) question {
	for {
		v := between(50, 500)
		w := v + nonZeroBetween(-20, 20)
		diff := absInt(v - w)
		val := float64(diff) / float64(v) * 100
		if !isExact2(cleanFloat(val)) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Nilai sejati %d dihampiri dengan %d. Galat relatif (dalam %%) adalah?", v, w),
			options: buildOptions(round2(val), []float64{float64(diff), round2(float64(diff) / float64(w) * 100), round2(val / 100), round2(val + 1)}, optionCount, false),
		}
	}
}

// ---------- PD Biasa ----------

func odeString(p, q int) string { return "y''" + linTerm(p, "y'") + linTerm(q, "y") }

func genKCharacteristicRoots(optionCount int) question {
	for {
		r1, r2 := between(-5, 5), between(-5, 5)
		if r1 == r2 {
			continue
		}
		hi, lo := max(r1, r2), min(r1, r2)
		ode := odeString(-(r1 + r2), r1*r2)
		if rand.IntN(2) == 0 {
			return question{
				prompt:  fmt.Sprintf("Solusi umum PD %s = 0 berbentuk y = C₁e^(r₁x) + C₂e^(r₂x). Nilai r terbesar adalah?", ode),
				options: buildOptions(float64(hi), []float64{float64(lo), float64(-hi), float64(-(r1 + r2)), float64(r1 * r2)}, optionCount, true),
			}
		}
		A, B := between(-5, 5), between(-9, 9)
		c1 := float64(B-lo*A) / float64(hi-lo)
		if !isExact2(c1) {
			continue
		}
		c2 := float64(A) - c1
		term := func(coef string, r int) string { // e^(0x) = 1 -> tulis konstantanya saja
			if r == 0 {
				return coef
			}
			return fmt.Sprintf("%se^(%s)", coef, formatPoly(r, 0))
		}
		return question{
			prompt: fmt.Sprintf(
				"PD %s = 0 dengan y(0) = %d dan y'(0) = %d memiliki solusi y = %s + %s. Nilai C₁ = ?",
				ode, A, B, term("C₁", hi), term("C₂", lo),
			),
			options: buildOptions(round2(c1), []float64{round2(c2), -round2(c1), float64(A), float64(B)}, optionCount, true),
		}
	}
}

func genKSeparable(optionCount int) question {
	t := pickOne(pyTriples)
	k := between(1, 3)
	if rand.IntN(2) == 0 {
		// dy/dx = x/y, y(0) = a -> y² = x² + a²
		return question{
			prompt:  fmt.Sprintf("Solusi PD dy/dx = x/y dengan y(0) = %d (y > 0). Nilai y(%d) adalah?", t.a*k, t.b*k),
			options: buildOptions(float64(t.c*k), []float64{float64((t.a + t.b) * k), float64(t.a * k), float64(t.b * k), float64(t.c*k + 1)}, optionCount, false),
		}
	}
	// dy/dx = -x/y, y(0) = c -> x² + y² = c²
	return question{
		prompt:  fmt.Sprintf("Solusi PD dy/dx = -x/y dengan y(0) = %d (y > 0). Nilai y(%d) adalah?", t.c*k, t.a*k),
		options: buildOptions(float64(t.b*k), []float64{float64((t.c - t.a) * k), float64(t.c * k), float64(t.a * k), float64(t.b*k + 1)}, optionCount, false),
	}
}

func genKLinearFirstOrder(optionCount int) question {
	for {
		a, m := between(1, 4), between(0, 3)
		val := 1 / float64(m+1+a)
		if !isExact2(val) {
			continue
		}
		rhs := map[int]string{0: "1", 1: "x"}[m]
		if m >= 2 {
			rhs = fmt.Sprintf("x^%d", m)
		}
		return question{
			prompt:  fmt.Sprintf("Solusi khusus PD y' + (%d/x)y = %s berbentuk y = Cx^%d. Nilai C = ?", a, rhs, m+1),
			options: buildOptions(round2(val), []float64{round2(1 / float64(m+1)), round2(1 / float64(a)), float64(m + 1 + a), 1}, optionCount, false),
		}
	}
}

func genKExactEquation(optionCount int) question {
	c := nonZeroBetween(-5, 5)
	p, q := between(-9, 9), nonZeroBetween(-9, 9)
	correct := float64(2 * c)
	return question{
		prompt:  fmt.Sprintf("Persamaan (kxy%s) dx + (%s) dy = 0 merupakan PD eksak. Nilai k = ?", linTerm(p, ""), formatTerms([]int{c, q}, []string{"x²", "y"})),
		options: buildOptions(correct, []float64{float64(c), -correct, round2(float64(c) / 2), correct + 1}, optionCount, true),
	}
}

func genKLaplaceEval(optionCount int) question {
	for {
		s := between(1, 4)
		if rand.IntN(2) == 0 {
			a, b, c := between(-5, 5), between(-5, 5), between(-5, 5)
			if a == 0 && b == 0 {
				continue
			}
			sf := float64(s)
			val := cleanFloat(2*float64(a)/(sf*sf*sf) + float64(b)/(sf*sf) + float64(c)/sf)
			if !isExact2(val) {
				continue
			}
			return question{
				prompt:  fmt.Sprintf("Transformasi Laplace F(s) dari f(t) = %s, dievaluasi pada s = %d, adalah?", formatPolyVar("t", a, b, c), s),
				options: buildOptions(round2(val), []float64{round2(float64(a)/(sf*sf*sf) + float64(b)/(sf*sf) + float64(c)/sf), float64(a*s*s + b*s + c), round2(val + 1), -round2(val)}, optionCount, true),
			}
		}
		a := nonZeroBetween(-5, 3)
		k := between(1, 9)
		if s <= a {
			continue
		}
		val := float64(k) / float64(s-a)
		if !isExact2(val) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Transformasi Laplace F(s) dari f(t) = %de^(%s), dievaluasi pada s = %d, adalah?", k, formatPolyVar("t", a, 0), s),
			options: buildOptions(round2(val), []float64{round2(float64(k) / float64(s-a+1)), float64(k), round2(float64(k) / float64(s)), round2(val + 1)}, optionCount, true),
		}
	}
}

func genKSystemEigen(optionCount int) question {
	l1 := between(-5, 6)
	l2 := between(-5, 6)
	for l2 == l1 {
		l2 = between(-5, 6)
	}
	p, pinv := unimodular2()
	a := matMulInt(matMulInt(p, diagInt(l1, l2)), pinv)
	hi := max(l1, l2)
	return question{
		prompt:  fmt.Sprintf("Sistem PD X' = AX dengan A = %s (baris dipisah titik koma). Nilai eigen terbesar dari A (laju mode dominan) adalah?", formatMatrix(a)),
		options: buildOptions(float64(hi), []float64{float64(min(l1, l2)), float64(l1 + l2), float64(l1 * l2), float64(hi + 1)}, optionCount, true),
	}
}

// ---------- PD Parsial ----------

func genKPDEClassify(optionCount int) question {
	if rand.IntN(2) == 0 {
		m, s, t := between(1, 3), between(1, 4), between(1, 4)
		A, C := m*s*s, m*t*t
		correct := float64(2 * m * s * t)
		return question{
			prompt:  fmt.Sprintf("Nilai k > 0 agar PD %s + k·u_xy + %s = 0 bertipe parabolik adalah?", formatTerms([]int{A}, []string{"u_xx"}), formatTerms([]int{C}, []string{"u_yy"})),
			options: buildOptions(correct, []float64{float64(4 * A * C), correct * correct, float64(m * s * t), correct + 1}, optionCount, false),
		}
	}
	A, B, C := nonZeroBetween(-4, 5), between(-6, 6), nonZeroBetween(-4, 5)
	correct := float64(B*B - 4*A*C)
	return question{
		prompt:  fmt.Sprintf("Nilai diskriminan B² - 4AC untuk PD %s = 0 adalah?", formatTerms([]int{A, B, C}, []string{"u_xx", "u_xy", "u_yy"})),
		options: buildOptions(correct, []float64{float64(B*B + 4*A*C), float64(4*A*C - B*B), float64(B - 4*A*C), correct + 1}, optionCount, true),
	}
}

func sinTerm(n int, v string) string {
	if n == 1 {
		return "sin(" + v + ")"
	}
	return fmt.Sprintf("sin(%d%s)", n, v)
}

func genKHeatDecay(optionCount int) question {
	alpha, n := between(1, 6), between(1, 5)
	correct := float64(alpha * n * n)
	return question{
		prompt:  fmt.Sprintf("u(x, t) = e^(-kt) %s adalah solusi persamaan panas u_t = %s. Nilai k = ?", sinTerm(n, "x"), formatTerms([]int{alpha}, []string{"u_xx"})),
		options: buildOptions(correct, []float64{float64(alpha * n), float64(alpha + n*n), float64(n * n), float64(alpha)}, optionCount, false),
	}
}

func genKWaveFreq(optionCount int) question {
	c, n := between(1, 6), between(1, 5)
	correct := float64(c * n)
	return question{
		prompt:  fmt.Sprintf("u(x, t) = %s cos(kt) memenuhi persamaan gelombang u_tt = %s. Nilai k > 0 adalah?", sinTerm(n, "x"), formatTerms([]int{c * c}, []string{"u_xx"})),
		options: buildOptions(correct, []float64{float64(c * c * n), float64(c * n * n), float64(c), float64(n)}, optionCount, false),
	}
}

func genKLaplaceHarmonicPDE(optionCount int) question {
	if rand.IntN(2) == 0 {
		a := nonZeroBetween(-5, 5)
		correct := float64(absInt(a))
		return question{
			prompt:  fmt.Sprintf("Nilai k > 0 agar u(x, y) = e^(%s) cos(ky) memenuhi persamaan Laplace u_xx + u_yy = 0 adalah?", formatPoly(a, 0)),
			options: buildOptions(correct, []float64{float64(a * a), correct + 1, float64(2 * absInt(a)), 1}, optionCount, false),
		}
	}
	// u = a(x⁴ + y⁴) + kx²y²  ->  Δu = (12a + 2k)(x² + y²) = 0  ->  k = -6a
	a := nonZeroBetween(-4, 4)
	correct := float64(-6 * a)
	return question{
		prompt:  fmt.Sprintf("Nilai k agar u(x, y) = %s + kx²y² memenuhi persamaan Laplace u_xx + u_yy = 0 adalah?", formatTerms([]int{a, a}, []string{"x⁴", "y⁴"})),
		options: buildOptions(correct, []float64{-correct, float64(-12 * a), float64(-3 * a), 0}, optionCount, true),
	}
}

func genKSeparationEigen(optionCount int) question {
	m, n := between(1, 4), between(1, 5)
	l := "π"
	if m != 1 {
		l = fmt.Sprintf("π/%d", m)
	}
	correct := float64(n * n * m * m)
	return question{
		prompt:  fmt.Sprintf("Masalah nilai batas X'' + λX = 0, X(0) = X(%s) = 0 memiliki nilai eigen λ_n. Nilai λ_%d = ?", l, n),
		options: buildOptions(correct, []float64{float64(n * m), float64(n * n), float64(m * m), correct + 1}, optionCount, false),
	}
}
