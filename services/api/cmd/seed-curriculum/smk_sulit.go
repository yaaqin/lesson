package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strings"
)

// ---------- SMK/SMA - Campuran Sulit (batch 5) ----------
//
// Versi lebih sulit dari 20 materi batch 2-3 (soal multi-langkah: SPLTV, teorema
// sisa 2 pembagi, rumus jumlah sudut, median data berkelompok, volume benda
// putar, irisan kerucut bentuk umum, dst), dicampur jadi 10 level + 1 ujian.
// Level: bank 40, wajib 25, 90 detik/soal. Ujian: bank 60, wajib 40, 60 menit.

const (
	smkHardLevelRequired    = 25
	smkHardLevelBank        = 40
	smkHardLevelTimeSeconds = 90 // per soal
	smkHardExamRequired     = 40
	smkHardExamBank         = 60
	smkHardExamTimeSeconds  = 3600 // total sesi ujian
)

func smkCampuranSulitChallenges() []challengeSpec {
	all := mixOf(
		mixOf(genExpEquationDiffBase, genLogChain, genLogEquation),
		mixOf(genArithFromTwoTerms, genGeoFromTwoTerms, genArithInsertion),
		genSPLTV,
		mixOf(genPolyRemainderTwoFactors, genInverseComposition, genFactorTheorem),
		mixOf(genTrigSumAngle, genHeronArea, genTrigEquationSum),
		mixOf(genVectorProjection, genVectorPerpendicular, genVectorAngle),
		mixOf(genSelectComposition, genArrangementAdjacent, genCircularArrangement, genProbabilityTwoDraws),
		genLinearProgramMin,
		mixOf(genSingularMatrix, genDetProperties, genMatrixEquation),
		mixOf(genCircleTangentToLine, genMatrixTransformPoint, genRotateAroundPoint),
		mixOf(genLimitSqrt, genLimitSqrtInfinity, genLimitCosine, genLimitCubicFactor),
		mixOf(genChainRule, genProductRule, genBoxOptimization, genTangentLineIntercept),
		mixOf(genIntegralPower, genAreaParabolaLine, genVolumeRotation),
		genGroupedData,
		mixOf(genPyramid, genCuboidPointToMidpoint, genAngleLinePlane),
		mixOf(genParabolaGeneral, genEllipseGeneral, genHyperbolaGeneral),
		mixOf(genAnnuityPrincipalN, genPresentValue, genDepreciationStraight, genDecliningBalance),
		mixOf(genSPLTVWord, genLinearProgramMinWord, genMatrixThreeProducts),
		mixOf(genTwoElevations, genBearingDistance, genCosineNavigation),
		genTruthTableHard,
	)

	challenges := make([]challengeSpec, 0, 11)
	for level := 1; level <= 10; level++ {
		name := fmt.Sprintf("Campuran Sulit Level %d", level)
		challenges = append(challenges, buildChallenge(name, false, smkHardLevelRequired, smkHardLevelBank, smkPassThresholdPcnt, smkOptionCount, smkHardLevelTimeSeconds, all))
	}
	challenges = append(challenges, buildChallenge("Ujian Campuran Sulit SMK/SMA", true, smkHardExamRequired, smkHardExamBank, smkPassThresholdPcnt, smkOptionCount, smkHardExamTimeSeconds, all))
	return challenges
}

// ---------- helper ----------

// formatTerms: "2x - y + 3z" dari koefisien & nama variabel ("" = konstanta).
func formatTerms(coeffs []int, names []string) string {
	var b strings.Builder
	for i, c := range coeffs {
		if c == 0 {
			continue
		}
		switch {
		case b.Len() == 0 && c < 0:
			b.WriteString("-")
		case b.Len() > 0 && c < 0:
			b.WriteString(" - ")
		case b.Len() > 0:
			b.WriteString(" + ")
		}
		if absInt(c) != 1 || names[i] == "" {
			fmt.Fprintf(&b, "%d", absInt(c))
		}
		b.WriteString(names[i])
	}
	if b.Len() == 0 {
		return "0"
	}
	return b.String()
}

func det3(m [][]int) int {
	return m[0][0]*(m[1][1]*m[2][2]-m[1][2]*m[2][1]) -
		m[0][1]*(m[1][0]*m[2][2]-m[1][2]*m[2][0]) +
		m[0][2]*(m[1][0]*m[2][1]-m[1][1]*m[2][0])
}

// isExact2: nilai desimal yang pas dalam 2 angka di belakang koma.
func isExact2(v float64) bool { return math.Abs(v*100-math.Round(v*100)) < 1e-6 }

// ---------- Eksponen & Logaritma ----------

func genExpEquationDiffBase(optionCount int) question {
	b := pickOne([]int{2, 3})
	for {
		m, n := between(1, 3), between(1, 3)
		if m == n {
			continue
		}
		x, p := between(-4, 6), between(-3, 4)
		num := m*(x+p) - n*x // n(x + q) = m(x + p)
		if num%n != 0 {
			continue
		}
		q := num / n
		if q == p || absInt(q) > 9 {
			continue
		}
		correct := float64(x)
		return question{
			prompt:  fmt.Sprintf("Jika %d^(%s) = %d^(%s), nilai x = ?", intPow(b, m), formatPoly(1, p), intPow(b, n), formatPoly(1, q)),
			options: buildOptions(correct, []float64{-correct, correct + 1, correct - 1, float64(q - p), float64(p - q)}, optionCount, true),
		}
	}
}

func genLogChain(optionCount int) question {
	base := pickOne([]int{2, 3})
	mid := pickOne([]int{5, 7, 11})
	k := between(2, map[int]int{2: 7, 3: 5}[base])
	correct := float64(k)
	return question{
		prompt:  fmt.Sprintf("Log basis %d dari %d × Log basis %d dari %d = ?", base, mid, mid, intPow(base, k)),
		options: buildOptions(correct, []float64{correct + 1, correct - 1, 2 * correct, float64(mid), float64(k * mid)}, optionCount, false),
	}
}

func genLogEquation(optionCount int) question {
	// (x + p)(x + q) = b^n dengan x + p = b^i, x + q = b^(n-i); akar lain bikin argumen negatif
	base := pickOne([]int{2, 3})
	n := between(3, map[int]int{2: 6, 3: 4}[base])
	i := between(0, (n-1)/2)
	u, v := intPow(base, i), intPow(base, n-i)
	x := between(1, 10)
	p, q := u-x, v-x
	other := -v - p
	correct := float64(x)
	return question{
		prompt: fmt.Sprintf(
			"Nilai x yang memenuhi Log basis %d dari (%s) + Log basis %d dari (%s) = %d adalah?",
			base, formatPoly(1, p), base, formatPoly(1, q), n,
		),
		options: buildOptions(correct, []float64{float64(other), float64(u), float64(v), correct + 1}, optionCount, true),
	}
}

// ---------- Barisan & Deret ----------

func genArithFromTwoTerms(optionCount int) question {
	a, d := between(-10, 15), nonZeroBetween(-5, 8)
	i := between(2, 5)
	j := i + between(2, 6)
	n := between(10, 25)
	un := a + (n-1)*d
	correct := float64(n * (a + un) / 2)
	return question{
		prompt: fmt.Sprintf(
			"Pada barisan aritmetika diketahui U%d = %d dan U%d = %d. Jumlah %d suku pertama barisan tersebut adalah?",
			i, a+(i-1)*d, j, a+(j-1)*d, n,
		),
		options: buildOptions(correct, []float64{float64(un), 2 * correct, correct + float64(n*d), correct - float64(a), correct + float64(d)}, optionCount, true),
	}
}

func genGeoFromTwoTerms(optionCount int) question {
	// selisih indeks ganjil (3) biar rasio tunggal (gak ambigu ±r)
	r := pickOne([]int{2, 3, -2})
	a := between(1, 4)
	i := between(1, 2)
	j := i + 3
	ui, uj := a*intPow(r, i-1), a*intPow(r, j-1)
	maxN := 9
	if r == 3 {
		maxN = 7
	}
	if rand.IntN(2) == 0 {
		n := between(j+1, maxN)
		correct := float64(a * intPow(r, n-1))
		return question{
			prompt:  fmt.Sprintf("Pada barisan geometri diketahui U%d = %d dan U%d = %d. Nilai U%d = ?", i, ui, j, uj, n),
			options: buildOptions(correct, []float64{float64(a * intPow(r, n)), float64(a * intPow(r, n-2)), -correct, correct + float64(uj)}, optionCount, true),
		}
	}
	n := between(5, 8)
	correct := float64(a * (intPow(r, n) - 1) / (r - 1))
	return question{
		prompt:  fmt.Sprintf("Pada deret geometri diketahui U%d = %d dan U%d = %d. Jumlah %d suku pertama deret tersebut adalah?", i, ui, j, uj, n),
		options: buildOptions(correct, []float64{float64(a * intPow(r, n-1)), float64(a * (intPow(r, n+1) - 1) / (r - 1)), -correct, correct + float64(a)}, optionCount, true),
	}
}

func genArithInsertion(optionCount int) question {
	a := between(1, 20)
	k := between(3, 9)
	d := between(2, 7)
	last := a + (k+1)*d
	correct := float64((k + 2) * (a + last) / 2)
	return question{
		prompt: fmt.Sprintf(
			"Di antara bilangan %d dan %d disisipkan %d bilangan sehingga semuanya membentuk barisan aritmetika. Jumlah semua suku barisan yang terbentuk adalah?",
			a, last, k,
		),
		options: buildOptions(correct, []float64{round2(float64(k*(a+last)) / 2), round2(float64((k+1)*(a+last)) / 2), correct + float64(d), correct - float64(a)}, optionCount, false),
	}
}

// ---------- Sistem Persamaan Linear Tiga Variabel ----------

func randomInvertible3() [][]int {
	for {
		m := make([][]int, 3)
		for i := range m {
			m[i] = []int{between(-3, 4), between(-3, 4), between(-3, 4)}
		}
		if det3(m) != 0 {
			return m
		}
	}
}

func genSPLTV(optionCount int) question {
	sol := []int{between(-5, 8), between(-5, 8), between(-5, 8)}
	m := randomInvertible3()
	names := []string{"x", "y", "z"}
	eqs := make([]string, 3)
	for i := range m {
		rhs := m[i][0]*sol[0] + m[i][1]*sol[1] + m[i][2]*sol[2]
		eqs[i] = fmt.Sprintf("%s = %d", formatTerms(m[i], names), rhs)
	}
	sum := sol[0] + sol[1] + sol[2]
	target, correct := "x + y + z", float64(sum)
	if idx := rand.IntN(4); idx < 3 {
		target, correct = names[idx], float64(sol[idx])
	}
	return question{
		prompt:  fmt.Sprintf("Diketahui sistem persamaan %s. Nilai %s = ?", strings.Join(eqs, "; "), target),
		options: buildOptions(correct, []float64{float64(sol[0]), float64(sol[1]), float64(sol[2]), float64(sum), correct + 1, correct - 1}, optionCount, true),
	}
}

// ---------- Fungsi & Polinomial ----------

func genPolyRemainderTwoFactors(optionCount int) question {
	p, q := nonZeroBetween(-6, 6), between(-9, 9)
	a := nonZeroBetween(-4, 4)
	b := nonZeroBetween(-4, 4)
	for b == a {
		b = nonZeroBetween(-4, 4)
	}
	r1, r2 := p*a+q, p*b+q
	target, correct, other := "p", float64(p), float64(q)
	if rand.IntN(2) == 0 {
		target, correct, other = "q", float64(q), float64(p)
	}
	return question{
		prompt: fmt.Sprintf(
			"Suku banyak P(x) jika dibagi (%s) bersisa %d, dan jika dibagi (%s) bersisa %d. Jika sisa pembagian P(x) oleh (%s)(%s) adalah px + q, nilai %s = ?",
			formatPoly(1, -a), r1, formatPoly(1, -b), r2, formatPoly(1, -a), formatPoly(1, -b), target,
		),
		options: buildOptions(correct, []float64{other, float64(r1), float64(r2), correct + 1, -correct}, optionCount, true),
	}
}

func genInverseComposition(optionCount int) question {
	a, b := between(2, 5), between(-9, 9)
	c, d := between(2, 5), between(-9, 9)
	t := between(-6, 8)
	v := a*(c*t+d) + b
	correct := float64(t)
	return question{
		prompt:  fmt.Sprintf("Diketahui f(x) = %s dan g(x) = %s. Nilai (f ∘ g)⁻¹(%d) = ?", formatPoly(a, b), formatPoly(c, d), v),
		options: buildOptions(correct, []float64{float64(a*(c*v+d) + b), round2(float64(v-b) / float64(a)), correct + 1, correct - 1, -correct}, optionCount, true),
	}
}

func genFactorTheorem(optionCount int) question {
	r := nonZeroBetween(-3, 3)
	k := between(-6, 6)
	b := between(-9, 9)
	c := -(r*r*r + k*r*r + b*r) // P(r) = 0
	correct := float64(k)
	return question{
		prompt: fmt.Sprintf(
			"Jika (%s) adalah faktor dari P(x) = x³ + kx²%s%s, nilai k = ?",
			formatPoly(1, -r), linTerm(b, "x"), linTerm(c, ""),
		),
		options: buildOptions(correct, []float64{-correct, correct + 1, correct - 1, float64(c), float64(r)}, optionCount, true),
	}
}

// ---------- Trigonometri ----------

func genTrigSumAngle(optionCount int) question {
	// A, B lancip dari segitiga 3-4-5 -> semua hasil = n/25 (desimal pas)
	sA := pickOne([]int{3, 4})
	sB := pickOne([]int{3, 4})
	cA, cB := 7-sA, 7-sB
	vals := map[string]float64{
		"sin(A + B)": float64(sA*cB+cA*sB) / 25,
		"cos(A + B)": float64(cA*cB-sA*sB) / 25,
		"sin(A - B)": float64(sA*cB-cA*sB) / 25,
		"cos(A - B)": float64(cA*cB+sA*sB) / 25,
	}
	all := []float64{}
	for _, v := range vals {
		all = append(all, round2(v))
	}
	all = append(all, round2(float64(sA+sB)/5), round2(float64(sA*sB)/25))
	if rand.IntN(3) == 0 {
		double := map[string]float64{"sin 2A": float64(2*sA*cA) / 25, "cos 2A": float64(cA*cA-sA*sA) / 25}
		key := pickOne([]string{"sin 2A", "cos 2A"})
		return question{
			prompt:  fmt.Sprintf("Diketahui sin A = %d/5 dengan A sudut lancip. Nilai %s = ?", sA, key),
			options: buildOptions(round2(double[key]), append(all, round2(float64(2*sA)/5)), optionCount, true),
		}
	}
	key := pickOne([]string{"sin(A + B)", "cos(A + B)", "sin(A - B)", "cos(A - B)"})
	return question{
		prompt:  fmt.Sprintf("Diketahui sin A = %d/5 dan sin B = %d/5 dengan A dan B sudut lancip. Nilai %s = ?", sA, sB, key),
		options: buildOptions(round2(vals[key]), all, optionCount, true),
	}
}

func genHeronArea(optionCount int) question {
	type heron struct{ a, b, c, area int }
	h := pickOne([]heron{{13, 14, 15, 84}, {5, 5, 6, 12}, {5, 5, 8, 12}, {10, 13, 13, 60}, {9, 10, 17, 36}, {7, 15, 20, 42}, {6, 25, 29, 60}, {8, 29, 35, 84}})
	k := between(1, 3)
	sides := []int{h.a * k, h.b * k, h.c * k}
	rand.Shuffle(3, func(i, j int) { sides[i], sides[j] = sides[j], sides[i] })
	correct := float64(h.area * k * k)
	s := (sides[0] + sides[1] + sides[2]) / 2
	return question{
		prompt:  fmt.Sprintf("Luas segitiga dengan panjang sisi %d cm, %d cm, dan %d cm adalah? (cm²)", sides[0], sides[1], sides[2]),
		options: buildOptions(correct, []float64{float64(s), float64(sides[0] * sides[1] / 2), 2 * correct, correct + float64(k*k)}, optionCount, false),
	}
}

func genTrigEquationSum(optionCount int) question {
	type trigEq struct {
		eq  string
		sol []int
	}
	e := pickOne([]trigEq{
		{"sin x = 1/2", []int{30, 150}}, {"sin x = -1/2", []int{210, 330}},
		{"cos x = 1/2", []int{60, 300}}, {"cos x = -1/2", []int{120, 240}},
		{"tan x = 1", []int{45, 225}}, {"tan x = -1", []int{135, 315}},
		{"2 sin x = √3", []int{60, 120}}, {"2 cos x = √3", []int{30, 330}},
		{"sin 2x = 1", []int{45, 225}}, {"cos 2x = 0", []int{45, 135, 225, 315}},
		{"2 sin² x = 1", []int{45, 135, 225, 315}}, {"tan² x = 3", []int{60, 120, 240, 300}},
	})
	sum := 0
	for _, s := range e.sol {
		sum += s
	}
	correct := float64(sum)
	return question{
		prompt:  fmt.Sprintf("Jumlah semua nilai x yang memenuhi %s untuk 0° ≤ x ≤ 360° adalah? (derajat)", e.eq),
		options: buildOptions(correct, []float64{float64(e.sol[0]), float64(e.sol[len(e.sol)-1]), 360, 180, correct + 180, correct - 90}, optionCount, false),
	}
}

// ---------- Vektor ----------

func genVectorProjection(optionCount int) question {
	for {
		var v []int
		var length int
		if rand.IntN(2) == 0 {
			t := pickOne(pyTriples)
			v, length = []int{t.a, t.b}, t.c
		} else {
			q := pickOne(pyQuadruples)
			v, length = []int{q.a, q.b, q.c}, q.d
		}
		for i := range v {
			v[i] *= randSign()
		}
		rand.Shuffle(len(v), func(i, j int) { v[i], v[j] = v[j], v[i] })
		u := make([]int, len(v))
		dot := 0
		for i := range u {
			u[i] = between(-6, 9)
			dot += u[i] * v[i]
		}
		correct := float64(dot) / float64(length)
		if dot == 0 || !isExact2(correct) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Panjang proyeksi skalar vektor u = (%s) pada vektor v = (%s) adalah?", joinInts(u), joinInts(v)),
			options: buildOptions(round2(correct), []float64{float64(dot), round2(float64(dot) / float64(length*length)), -round2(correct), round2(correct + 1)}, optionCount, true),
		}
	}
}

func genVectorPerpendicular(optionCount int) question {
	for {
		c := nonZeroBetween(-4, 4)
		a, b, d, e := between(-5, 6), between(-5, 6), between(-5, 6), between(-5, 6)
		s := a*d + b*e
		if s%c != 0 {
			continue
		}
		k := -s / c
		if absInt(k) > 12 {
			continue
		}
		correct := float64(k)
		return question{
			prompt:  fmt.Sprintf("Vektor u = (k, %d, %d) tegak lurus dengan vektor v = (%d, %d, %d). Nilai k = ?", a, b, c, d, e),
			options: buildOptions(correct, []float64{-correct, float64(s), correct + 1, correct - 1}, optionCount, true),
		}
	}
}

func genVectorAngle(optionCount int) question {
	base := pickOne([][2][3]int{
		{{1, 0, 0}, {1, 1, 0}}, {{1, 1, 0}, {0, 1, 1}}, {{1, 1, 0}, {1, -1, 0}}, {{1, 0, 0}, {-1, 1, 0}},
		{{1, 1, 0}, {-1, 0, -1}}, {{1, 2, 2}, {2, 1, -2}}, {{0, 1, 1}, {1, 0, 1}},
	})
	// skala positif + permutasi & pencerminan koordinat yang sama -> sudut tetap
	perm := []int{0, 1, 2}
	rand.Shuffle(3, func(i, j int) { perm[i], perm[j] = perm[j], perm[i] })
	signs := []int{randSign(), randSign(), randSign()}
	k1, k2 := between(1, 3), between(1, 3)
	u, v := make([]int, 3), make([]int, 3)
	for i := range 3 {
		u[i] = base[0][perm[i]] * signs[i] * k1
		v[i] = base[1][perm[i]] * signs[i] * k2
	}
	dot, nu, nv := 0.0, 0.0, 0.0
	for i := range 3 {
		dot += float64(u[i] * v[i])
		nu += float64(u[i] * u[i])
		nv += float64(v[i] * v[i])
	}
	angle := math.Round(math.Acos(dot/math.Sqrt(nu*nv)) * 180 / math.Pi)
	return question{
		prompt:  fmt.Sprintf("Besar sudut antara vektor u = (%s) dan v = (%s) adalah? (derajat)", joinInts(u), joinInts(v)),
		options: buildOptions(angle, []float64{30, 45, 60, 90, 120, 135, 150}, optionCount, false),
	}
}

// ---------- Kombinatorik & Peluang ----------

func genSelectComposition(optionCount int) question {
	r, b := between(4, 8), between(3, 7)
	i, j := between(1, 3), between(1, 3)
	correct := float64(combination(r, i) * combination(b, j))
	return question{
		prompt: fmt.Sprintf(
			"Sebuah kotak berisi %d bola merah dan %d bola biru. Diambil %d bola sekaligus. Banyak cara terambil %d bola merah dan %d bola biru adalah?",
			r, b, i+j, i, j,
		),
		options: buildOptions(correct, []float64{float64(combination(r+b, i+j)), float64(combination(r, i) + combination(b, j)), 2 * correct, float64(combination(r, j) * combination(b, i))}, optionCount, false),
	}
}

func genArrangementAdjacent(optionCount int) question {
	n := between(4, 7)
	adjacent := 2 * factorial(n-1)
	notAdjacent := factorial(n) - adjacent
	cands := []float64{float64(factorial(n)), float64(factorial(n - 1)), float64(2 * factorial(n-2))}
	if rand.IntN(2) == 0 {
		return question{
			prompt:  fmt.Sprintf("Sebanyak %d orang, termasuk Andi dan Budi, duduk berjajar dalam satu baris. Banyak susunan duduk jika Andi dan Budi selalu berdampingan adalah?", n),
			options: buildOptions(float64(adjacent), append(cands, float64(notAdjacent)), optionCount, false),
		}
	}
	return question{
		prompt:  fmt.Sprintf("Sebanyak %d orang, termasuk Andi dan Budi, duduk berjajar dalam satu baris. Banyak susunan duduk jika Andi dan Budi tidak boleh berdampingan adalah?", n),
		options: buildOptions(float64(notAdjacent), append(cands, float64(adjacent)), optionCount, false),
	}
}

func genCircularArrangement(optionCount int) question {
	n := between(4, 8)
	if rand.IntN(2) == 0 {
		correct := float64(factorial(n - 1))
		return question{
			prompt:  fmt.Sprintf("Banyak cara %d orang duduk mengelilingi sebuah meja bundar adalah?", n),
			options: buildOptions(correct, []float64{float64(factorial(n)), float64(factorial(n - 2)), float64(factorial(n-1) / 2), correct + 1}, optionCount, false),
		}
	}
	correct := float64(2 * factorial(n-2))
	return question{
		prompt:  fmt.Sprintf("Sebanyak %d orang duduk mengelilingi meja bundar. Banyak cara duduk jika 2 orang tertentu harus selalu berdampingan adalah?", n),
		options: buildOptions(correct, []float64{float64(factorial(n - 1)), float64(factorial(n - 2)), float64(2 * factorial(n-1)), correct + 2}, optionCount, false),
	}
}

func genProbabilityTwoDraws(optionCount int) question {
	for {
		r, w := between(2, 8), between(2, 8)
		n := r + w
		total := float64(combination(n, 2))
		bothRed := float64(combination(r, 2)) / total
		oneEach := float64(r*w) / total
		label, correct, other := "keduanya merah", bothRed, oneEach
		if rand.IntN(2) == 0 {
			label, correct, other = "1 merah dan 1 putih", oneEach, bothRed
		}
		if !isExact2(correct) {
			continue
		}
		withReplacement := float64(r*r) / float64(n*n)
		return question{
			prompt: fmt.Sprintf(
				"Sebuah kantong berisi %d kelereng merah dan %d kelereng putih. Diambil 2 kelereng sekaligus secara acak. Peluang terambil %s adalah?",
				r, w, label,
			),
			options: buildOptions(round2(correct), []float64{round2(other), round2(withReplacement), round2(float64(r) / float64(n)), round2(1 - correct)}, optionCount, false),
		}
	}
}

// ---------- Program Linear (minimum) ----------

func lpMinValues(cons []lpConstraint, p, q int) (best float64, others []float64) {
	values := lpObjectiveValues(cons, true, p, q)
	return values[0], values[1:]
}

func genLinearProgramMin(optionCount int) question {
	cons := randomLPConstraints()
	p, q := between(2, 9), between(2, 9)
	best, others := lpMinValues(cons, p, q)
	return question{
		prompt: fmt.Sprintf(
			"Nilai minimum f(x, y) = %dx + %dy yang memenuhi %s ≥ %d, %s ≥ %d, x ≥ 0, y ≥ 0 adalah?",
			p, q, formatLinearXY(cons[0].c, cons[0].d), cons[0].e, formatLinearXY(cons[1].c, cons[1].d), cons[1].e,
		),
		options: buildOptions(best, append(others, best+1, best-1, 0), optionCount, false),
	}
}

// ---------- Matriks ----------

func genSingularMatrix(optionCount int) question {
	// det = (x + p)(x + s) - q·r = (x - r1)(x - r2)
	r1 := between(-5, 5)
	r2 := between(-5, 5)
	for r2 == r1 {
		r2 = between(-5, 5)
	}
	p := between(-4, 4)
	s := -(r1 + r2) - p
	prod := p*s - r1*r2 // = q·r
	q, r := 0, nonZeroBetween(-5, 5)
	if prod != 0 {
		var divs []int
		for d := 1; d <= absInt(prod); d++ {
			if prod%d == 0 && d <= 12 && absInt(prod/d) <= 12 {
				divs = append(divs, d)
			}
		}
		if len(divs) == 0 {
			divs = []int{1}
		}
		q = pickOne(divs) * randSign()
		r = prod / q
	}
	cell := func(c int) string { return "x" + linTerm(c, "") }
	matrix := fmt.Sprintf("[%s, %d; %d, %s]", cell(p), q, r, cell(s))
	sum, product := float64(r1+r2), float64(r1*r2)
	if rand.IntN(2) == 0 {
		return question{
			prompt:  fmt.Sprintf("Matriks A = %s (baris dipisah titik koma) merupakan matriks singular. Jumlah semua nilai x yang memenuhi adalah?", matrix),
			options: buildOptions(sum, []float64{-sum, product, sum + 1, sum - 1}, optionCount, true),
		}
	}
	return question{
		prompt:  fmt.Sprintf("Matriks A = %s (baris dipisah titik koma) merupakan matriks singular. Hasil kali semua nilai x yang memenuhi adalah?", matrix),
		options: buildOptions(product, []float64{-product, sum, product + 1, product - 1}, optionCount, true),
	}
}

func genDetProperties(optionCount int) question {
	dA, dB := nonZeroBetween(-5, 5), nonZeroBetween(-4, 4)
	k := between(2, 3)
	type variant struct {
		expr    string
		correct int
		naive   int
	}
	v := pickOne([]variant{
		{fmt.Sprintf("%dA", k), k * k * dA, k * dA},
		{"AB", dA * dB, dA + dB},
		{"A²B", dA * dA * dB, 2 * dA * dB},
		{fmt.Sprintf("%dAB", k), k * k * dA * dB, k * dA * dB},
		{"(AB)ᵀ", dA * dB, -dA * dB},
	})
	correct := float64(v.correct)
	return question{
		prompt:  fmt.Sprintf("A dan B matriks berordo 2×2 dengan det(A) = %d dan det(B) = %d. Nilai det(%s) = ?", dA, dB, v.expr),
		options: buildOptions(correct, []float64{float64(v.naive), -correct, float64(dA * dB), correct + 1}, optionCount, true),
	}
}

func genMatrixEquation(optionCount int) question {
	m, n := nonZeroBetween(-3, 3), nonZeroBetween(-3, 3)
	a := [][]int{{1 + m*n, m}, {n, 1}}
	inv := [][]int{{1, -m}, {-n, 1 + m*n}}
	b := [][]int{{between(-5, 6), between(-5, 6)}, {between(-5, 6), between(-5, 6)}}
	mul := func(x, y [][]int) [][]int {
		return [][]int{
			{x[0][0]*y[0][0] + x[0][1]*y[1][0], x[0][0]*y[0][1] + x[0][1]*y[1][1]},
			{x[1][0]*y[0][0] + x[1][1]*y[1][0], x[1][0]*y[0][1] + x[1][1]*y[1][1]},
		}
	}
	left, right := mul(inv, b), mul(b, inv) // AX = B -> X = A⁻¹B ; XA = B -> X = BA⁻¹
	i, j := rand.IntN(2), rand.IntN(2)
	eq, x, wrong := "AX = B", left, right
	if rand.IntN(2) == 0 {
		eq, x, wrong = "XA = B", right, left
	}
	correct := float64(x[i][j])
	return question{
		prompt: fmt.Sprintf(
			"Diketahui A = %s dan B = %s (baris dipisah titik koma). Jika %s, elemen baris ke-%d kolom ke-%d matriks X adalah?",
			formatMatrix(a), formatMatrix(b), eq, i+1, j+1,
		),
		options: buildOptions(correct, []float64{float64(wrong[i][j]), float64(mul(a, b)[i][j]), -correct, correct + 1}, optionCount, true),
	}
}

// ---------- Lingkaran & Transformasi ----------

func genCircleTangentToLine(optionCount int) question {
	t := pickOne(pyTriples[:3])
	p, q := t.a*randSign(), t.b*randSign()
	if rand.IntN(2) == 0 {
		p, q = q, p
	}
	a, b := between(-5, 5), between(-5, 5)
	r := between(1, 6)
	c := randSign()*r*t.c - (p*a + q*b) // |pa + qb + c| / √(p²+q²) = r
	correct := float64(r)
	return question{
		prompt: fmt.Sprintf(
			"Jari-jari lingkaran yang berpusat di titik (%d, %d) dan menyinggung garis %s = 0 adalah?",
			a, b, formatTerms([]int{p, q, c}, []string{"x", "y", ""}),
		),
		options: buildOptions(correct, []float64{float64(r * t.c), float64(r * r), correct + 1, float64(absInt(c))}, optionCount, false),
	}
}

func genMatrixTransformPoint(optionCount int) question {
	m := [][]int{{between(-3, 3), between(-3, 3)}, {between(-3, 3), between(-3, 3)}}
	x, y := between(-6, 6), between(-6, 6)
	e, f := nonZeroBetween(-5, 5), nonZeroBetween(-5, 5)
	x1, y1 := m[0][0]*x+m[0][1]*y+e, m[1][0]*x+m[1][1]*y+f
	// salah urutan: translasi dulu baru matriks
	x2, y2 := m[0][0]*(x+e)+m[0][1]*(y+f), m[1][0]*(x+e)+m[1][1]*(y+f)
	axis, correct, other, wrongOrder, noTrans := "x", x1, y1, x2, x1-e
	if rand.IntN(2) == 0 {
		axis, correct, other, wrongOrder, noTrans = "y", y1, x1, y2, y1-f
	}
	return question{
		prompt: fmt.Sprintf(
			"Titik P(%d, %d) ditransformasi oleh matriks %s (baris dipisah titik koma), lalu ditranslasi oleh T(%d, %d). Koordinat %s bayangan titik P adalah?",
			x, y, formatMatrix(m), e, f, axis,
		),
		options: buildOptions(float64(correct), []float64{float64(other), float64(wrongOrder), float64(noTrans), float64(correct + 1)}, optionCount, true),
	}
}

func genRotateAroundPoint(optionCount int) question {
	x, y := between(-6, 8), between(-6, 8)
	a, b := nonZeroBetween(-4, 4), nonZeroBetween(-4, 4)
	var desc string
	var rx, ry, ox, oy int // hasil benar & hasil kalau pusatnya dianggap O
	if rand.IntN(2) == 0 {
		desc = fmt.Sprintf("dirotasi 90° berlawanan arah jarum jam dengan pusat (%d, %d)", a, b)
		rx, ry = a-(y-b), b+(x-a)
		ox, oy = -y, x
	} else {
		desc = fmt.Sprintf("dicerminkan terhadap garis x = %d, lalu dicerminkan terhadap garis y = %d", a, b)
		rx, ry = 2*a-x, 2*b-y
		ox, oy = -x, -y
	}
	axis, correct, other, originGuess := "x", rx, ry, ox
	if rand.IntN(2) == 0 {
		axis, correct, other, originGuess = "y", ry, rx, oy
	}
	return question{
		prompt:  fmt.Sprintf("Titik P(%d, %d) %s. Koordinat %s bayangan titik P adalah?", x, y, desc, axis),
		options: buildOptions(float64(correct), []float64{float64(other), float64(originGuess), float64(-correct), float64(correct + 1)}, optionCount, true),
	}
}

// ---------- Limit ----------

func genLimitSqrt(optionCount int) question {
	// (√(x + p) - q) / (x - a), q² = a + p -> 1/(2q)
	q := pickOne([]int{1, 2, 5})
	a := nonZeroBetween(-5, 9)
	p := q*q - a
	correct := round2(1 / float64(2*q))
	return question{
		prompt:  fmt.Sprintf("lim (x→%d) (√(%s) - %d) / (%s) = ?", a, formatPoly(1, p), q, formatPoly(1, -a)),
		options: buildOptions(correct, []float64{float64(2 * q), round2(1 / float64(q)), 0, float64(q), 1}, optionCount, false),
	}
}

func genLimitSqrtInfinity(optionCount int) question {
	s := pickOne([]int{1, 2, 5})
	lead := s * s
	b := between(-10, 10)
	d := between(-10, 10)
	for d == b {
		d = between(-10, 10)
	}
	c1, c2 := between(-9, 9), between(-9, 9)
	correct := round2(float64(b-d) / float64(2*s))
	return question{
		prompt:  fmt.Sprintf("lim (x→∞) (√(%s) - √(%s)) = ?", formatPoly(lead, b, c1), formatPoly(lead, d, c2)),
		options: buildOptions(correct, []float64{float64(b - d), round2(float64(b-d) / float64(s)), 0, round2(float64(b+d) / float64(2*s))}, optionCount, true),
	}
}

func genLimitCosine(optionCount int) question {
	for {
		a := between(2, 6)
		b := pickOne([]int{1, 2, 4, 5})
		val := float64(a*a) / float64(2*b)
		if !isExact2(val) {
			continue
		}
		denom := "x²"
		if b != 1 {
			denom = fmt.Sprintf("%dx²", b)
		}
		return question{
			prompt:  fmt.Sprintf("lim (x→0) (1 - cos(%dx)) / (%s) = ?", a, denom),
			options: buildOptions(val, []float64{round2(float64(a*a) / float64(b)), round2(float64(a) / float64(2*b)), 0, round2(float64(a*a) / float64(4*b))}, optionCount, false),
		}
	}
}

func genLimitCubicFactor(optionCount int) question {
	// (x³ - a³) / (x² - a²) -> 3a² / 2a = 1,5a
	a := nonZeroBetween(-4, 5)
	correct := 1.5 * float64(a)
	return question{
		prompt:  fmt.Sprintf("lim (x→%d) (%s) / (%s) = ?", a, formatPoly(1, 0, 0, -a*a*a), formatPoly(1, 0, -a*a)),
		options: buildOptions(correct, []float64{float64(3 * a * a), float64(3 * a), float64(a), 0, -correct}, optionCount, true),
	}
}

// ---------- Turunan ----------

func genChainRule(optionCount int) question {
	a, b := nonZeroBetween(-3, 4), between(-5, 5)
	n := between(2, 4)
	k := between(-2, 3)
	inner := a*k + b
	correct := float64(n * a * intPow(inner, n-1))
	return question{
		prompt:  fmt.Sprintf("Jika f(x) = (%s)^%d, nilai f'(%d) = ?", formatPoly(a, b), n, k),
		options: buildOptions(correct, []float64{float64(n * intPow(inner, n-1)), float64(intPow(inner, n)), float64(a * intPow(inner, n-1)), correct + 1, -correct}, optionCount, true),
	}
}

func genProductRule(optionCount int) question {
	a, b := nonZeroBetween(-3, 4), between(-5, 5)
	c, d := nonZeroBetween(-3, 3), between(-6, 6)
	k := between(-2, 3)
	correct := float64(a*(c*k*k+d) + (a*k+b)*(2*c*k))
	return question{
		prompt:  fmt.Sprintf("Jika f(x) = (%s)(%s), nilai f'(%d) = ?", formatPoly(a, b), formatPoly(c, 0, d), k),
		options: buildOptions(correct, []float64{float64(a * 2 * c * k), float64((a*k + b) * (c*k*k + d)), correct + 1, correct - 1, -correct}, optionCount, true),
	}
}

func genBoxOptimization(optionCount int) question {
	// karton sisi 6t, potong x di sudut: V = x(6t - 2x)², maksimum di x = t, V = 16t³
	t := between(1, 5)
	side := 6 * t
	vol := func(x int) float64 { return float64(x * (side - 2*x) * (side - 2*x)) }
	if rand.IntN(2) == 0 {
		return question{
			prompt:  fmt.Sprintf("Selembar karton persegi bersisi %d cm dipotong persegi kecil yang sama besar di keempat sudutnya, lalu dilipat menjadi kotak tanpa tutup. Agar volume kotak maksimum, panjang sisi persegi yang dipotong adalah? (cm)", side),
			options: buildOptions(float64(t), []float64{float64(2 * t), float64(3 * t), round2(float64(t) / 2), float64(t + 1)}, optionCount, false),
		}
	}
	return question{
		prompt:  fmt.Sprintf("Selembar karton persegi bersisi %d cm dipotong persegi kecil yang sama besar di keempat sudutnya, lalu dilipat menjadi kotak tanpa tutup. Volume maksimum kotak tersebut adalah? (cm³)", side),
		options: buildOptions(vol(t), []float64{vol(2 * t), float64(side * side * t), float64(side * side * side / 27), vol(t) + float64(t*t)}, optionCount, false),
	}
}

func genTangentLineIntercept(optionCount int) question {
	a, b, c := nonZeroBetween(-3, 4), between(-8, 8), between(-9, 9)
	k := nonZeroBetween(-3, 4)
	fk := a*k*k + b*k + c
	m := 2*a*k + b
	correct := float64(fk - m*k)
	return question{
		prompt:  fmt.Sprintf("Garis singgung kurva y = %s di titik berabsis x = %d memotong sumbu y di titik (0, n). Nilai n = ?", formatPoly(a, b, c), k),
		options: buildOptions(correct, []float64{float64(fk), float64(m), float64(c), float64(fk + m*k), correct + 1}, optionCount, true),
	}
}

// ---------- Integral ----------

func genIntegralPower(optionCount int) question {
	n := between(3, 4)
	c := between(-3, 3)
	a := between(-2, 1)
	b := a + between(1, 3)
	F := func(x int) int { return intPow(x+c, n) }
	correct := float64(F(b) - F(a))
	return question{
		prompt:  fmt.Sprintf("∫ dari %d sampai %d %d(%s)^%d dx = ?", a, b, n, formatPoly(1, c), n-1),
		options: buildOptions(correct, []float64{float64(F(b)), float64(intPow(b+c, n-1) - intPow(a+c, n-1)), float64(n) * correct, correct + 1, -correct}, optionCount, true),
	}
}

func genAreaParabolaLine(optionCount int) question {
	// y = x² dan y = (r1 + r2)x - r1·r2 berpotongan di r1, r2 -> luas (r2 - r1)³/6
	d := pickOne([]int{3, 6, 9})
	r1 := between(-4, 3)
	r2 := r1 + d
	d3 := float64(d * d * d)
	correct := round2(d3 / 6)
	return question{
		prompt:  fmt.Sprintf("Luas daerah yang dibatasi kurva y = x² dan garis y = %s adalah? (satuan luas)", formatPoly(r1+r2, -r1*r2)),
		options: buildOptions(correct, []float64{round2(d3 / 2), round2(d3 / 3), round2(float64(d*d) / 2), correct + 1}, optionCount, false),
	}
}

func genVolumeRotation(optionCount int) question {
	suffix := "diputar 360° mengelilingi sumbu x. Volume benda putar yang terjadi adalah kπ satuan volume. Nilai k = ?"
	switch rand.IntN(3) {
	case 0:
		for {
			m, b := between(1, 3), between(1, 6)
			num := m * m * b * b * b
			if num%3 != 0 {
				continue
			}
			correct := float64(num / 3)
			curve := "x"
			if m != 1 {
				curve = fmt.Sprintf("%dx", m)
			}
			return question{
				prompt:  fmt.Sprintf("Daerah yang dibatasi garis y = %s, sumbu x, dan garis x = %d %s", curve, b, suffix),
				options: buildOptions(correct, []float64{round2(float64(m*b*b) / 2), float64(num), round2(float64(m*m*b*b) / 2), correct + 1}, optionCount, false),
			}
		}
	case 1:
		b := between(2, 9)
		correct := round2(float64(b*b) / 2)
		return question{
			prompt:  fmt.Sprintf("Daerah yang dibatasi kurva y = √x, sumbu x, dan garis x = %d %s", b, suffix),
			options: buildOptions(correct, []float64{float64(b * b), round2(2 * math.Pow(float64(b), 1.5) / 3), float64(b), correct + 1}, optionCount, false),
		}
	default:
		b := between(1, 3)
		b5 := float64(intPow(b, 5))
		correct := round2(b5 / 5)
		return question{
			prompt:  fmt.Sprintf("Daerah yang dibatasi kurva y = x², sumbu x, dan garis x = %d %s", b, suffix),
			options: buildOptions(correct, []float64{round2(float64(intPow(b, 3)) / 3), b5, round2(b5 / 3), correct + 1}, optionCount, false),
		}
	}
}

// ---------- Statistika data berkelompok ----------

type groupedData struct {
	start, width int
	freqs        []int
	text         string
}

func randomGroupedData() groupedData {
	width := pickOne([]int{5, 10})
	start := between(3, 6)*10 + 1
	freqs := make([]int, 5)
	parts := make([]string, 5)
	for i := range freqs {
		freqs[i] = between(2, 12)
		lo := start + i*width
		parts[i] = fmt.Sprintf("%d–%d: %d", lo, lo+width-1, freqs[i])
	}
	return groupedData{start, width, freqs, strings.Join(parts, ", ")}
}

func (g groupedData) lowerBoundary(i int) float64 {
	return float64(g.start+i*g.width) - 0.5
}

// quantile: L + ((pos - F) / f)·p, pos = n·frac (median frac = 1/2, Q1 = 1/4, Q3 = 3/4).
func (g groupedData) quantile(frac float64) (value float64, class int) {
	n := 0
	for _, f := range g.freqs {
		n += f
	}
	pos := float64(n) * frac
	cum := 0
	for i, f := range g.freqs {
		if float64(cum+f) >= pos {
			return g.lowerBoundary(i) + (pos-float64(cum))/float64(f)*float64(g.width), i
		}
		cum += f
	}
	return 0, -1
}

func genGroupedData(optionCount int) question {
	for {
		g := randomGroupedData()
		var label string
		var value float64
		var class int
		switch rand.IntN(4) {
		case 0:
			label = "Median"
			value, class = g.quantile(0.5)
		case 1:
			label = "Kuartil bawah (Q1)"
			value, class = g.quantile(0.25)
		case 2:
			label = "Kuartil atas (Q3)"
			value, class = g.quantile(0.75)
		default:
			// modus: kelas frekuensi tertinggi harus tunggal
			label = "Modus"
			class = 0
			for i, f := range g.freqs {
				if f > g.freqs[class] {
					class = i
				}
			}
			unique := true
			for i, f := range g.freqs {
				if i != class && f == g.freqs[class] {
					unique = false
				}
			}
			if !unique {
				continue
			}
			prev, next := 0, 0
			if class > 0 {
				prev = g.freqs[class-1]
			}
			if class < len(g.freqs)-1 {
				next = g.freqs[class+1]
			}
			d1, d2 := g.freqs[class]-prev, g.freqs[class]-next
			value = g.lowerBoundary(class) + float64(d1)/float64(d1+d2)*float64(g.width)
		}
		if class < 0 || !isExact2(value) {
			continue
		}
		correct := round2(value)
		lower := g.lowerBoundary(class)
		mid := lower + float64(g.width)/2
		return question{
			prompt:  fmt.Sprintf("Perhatikan data berkelompok berikut (interval: frekuensi): %s. %s data tersebut adalah?", g.text, label),
			options: buildOptions(correct, []float64{lower, mid, round2(correct + float64(g.width)/2), round2(correct - 1), round2(lower + float64(g.width))}, optionCount, false),
		}
	}
}

// ---------- Dimensi Tiga ----------

func genPyramid(optionCount int) question {
	// limas persegi T.ABCD sisi alas 2m, tinggi t, rusuk tegak e: e² = t² + 2m²
	for {
		m, t := between(1, 8), between(1, 20)
		e2 := t*t + 2*m*m
		e := int(math.Round(math.Sqrt(float64(e2))))
		if e*e != e2 {
			continue
		}
		s := 2 * m
		slant := round2(math.Sqrt(float64(e*e - m*m)))
		if rand.IntN(2) == 0 && (s*s*t)%3 == 0 {
			correct := float64(s * s * t / 3)
			return question{
				prompt:  fmt.Sprintf("Limas beraturan T.ABCD memiliki alas persegi dengan panjang sisi %d cm dan panjang rusuk tegak %d cm. Volume limas tersebut adalah? (cm³)", s, e),
				options: buildOptions(correct, []float64{float64(s * s * t), float64(s * s * e / 3), round2(float64(s*s) * slant / 3), correct + float64(s)}, optionCount, false),
			}
		}
		correct := float64(t)
		return question{
			prompt:  fmt.Sprintf("Limas beraturan T.ABCD memiliki alas persegi dengan panjang sisi %d cm dan panjang rusuk tegak %d cm. Tinggi limas tersebut adalah? (cm)", s, e),
			options: buildOptions(correct, []float64{slant, float64(e), float64(m), correct + 1}, optionCount, false),
		}
	}
}

func genCuboidPointToMidpoint(optionCount int) question {
	// A(0,0,0), titik tengah GH = (p/2, l, t) -> jarak² = (p/2)² + l² + t²
	q := pickOne(pyQuadruples)
	dims := []int{q.a, q.b, q.c}
	rand.Shuffle(3, func(i, j int) { dims[i], dims[j] = dims[j], dims[i] })
	p, l, t := 2*dims[0], dims[1], dims[2]
	correct := float64(q.d)
	return question{
		prompt:  fmt.Sprintf("Balok ABCD.EFGH dengan AB = %d cm, BC = %d cm, dan AE = %d cm. Jarak titik A ke titik tengah rusuk GH adalah? (cm)", p, l, t),
		options: buildOptions(correct, []float64{round2(math.Sqrt(float64(p*p + l*l + t*t))), round2(math.Sqrt(float64(l*l + t*t))), float64(p/2 + l + t), correct + 1}, optionCount, false),
	}
}

func genAngleLinePlane(optionCount int) question {
	type pair struct {
		desc  string
		angle int
	}
	p := pickOne([]pair{
		{"garis BG dan bidang ABCD", 45}, {"garis AH dan bidang ABCD", 45}, {"garis AE dan bidang ABCD", 90},
		{"garis AC dan bidang EFGH", 0}, {"garis BE dan bidang ABCD", 45}, {"garis EG dan bidang BDHF", 90},
		{"garis AF dan bidang BCGF", 45}, {"bidang ABCD dan bidang ADHE", 90}, {"bidang ABCD dan bidang ABGH", 45},
		{"bidang ACGE dan bidang BDHF", 90},
	})
	return question{
		prompt:  fmt.Sprintf("Pada kubus ABCD.EFGH dengan rusuk %d cm, besar sudut antara %s adalah? (derajat)", between(4, 12), p.desc),
		options: buildOptions(float64(p.angle), []float64{0, 30, 45, 60, 90}, optionCount, false),
	}
}

// ---------- Irisan Kerucut (bentuk umum) ----------

func genParabolaGeneral(optionCount int) question {
	// (y - k)² = 4p(x - h)  ->  y² - 2ky - 4px + (k² + 4ph) = 0
	p := nonZeroBetween(-4, 4)
	h, k := between(-5, 5), between(-5, 5)
	eq := "y²" + linTerm(-2*k, "y") + linTerm(-4*p, "x") + linTerm(k*k+4*p*h, "") + " = 0"
	cands := []float64{float64(h), float64(p), float64(h - p), float64(h + p), float64(k), float64(-h), float64(4 * p)}
	switch rand.IntN(3) {
	case 0:
		return question{prompt: fmt.Sprintf("Koordinat x titik fokus parabola %s adalah?", eq), options: buildOptions(float64(h+p), cands, optionCount, true)}
	case 1:
		return question{prompt: fmt.Sprintf("Persamaan garis direktris parabola %s adalah x = ?", eq), options: buildOptions(float64(h-p), cands, optionCount, true)}
	default:
		return question{prompt: fmt.Sprintf("Koordinat x titik puncak parabola %s adalah?", eq), options: buildOptions(float64(h), cands, optionCount, true)}
	}
}

func genEllipseGeneral(optionCount int) question {
	// b²(x - h)² + a²(y - k)² = a²b², a > b, c² = a² - b²
	t := pickOne([]pyTriple{{3, 4, 5}, {5, 12, 13}})
	a := t.c
	b, c := t.a, t.b
	if rand.IntN(2) == 0 {
		b, c = c, b
	}
	h, k := between(-4, 4), between(-4, 4)
	a2, b2 := a*a, b*b
	eq := fmt.Sprintf("%dx² + %dy²%s%s%s = 0", b2, a2, linTerm(-2*b2*h, "x"), linTerm(-2*a2*k, "y"), linTerm(b2*h*h+a2*k*k-a2*b2, ""))
	cands := []float64{float64(h), float64(h + a), float64(h + b), float64(h + c), float64(2 * a), float64(2 * c), float64(c)}
	switch rand.IntN(3) {
	case 0:
		return question{prompt: fmt.Sprintf("Koordinat x titik fokus sebelah kanan dari elips %s adalah?", eq), options: buildOptions(float64(h+c), cands, optionCount, true)}
	case 1:
		return question{prompt: fmt.Sprintf("Koordinat x titik pusat elips %s adalah?", eq), options: buildOptions(float64(h), cands, optionCount, true)}
	default:
		return question{prompt: fmt.Sprintf("Panjang sumbu mayor elips %s adalah?", eq), options: buildOptions(float64(2*a), cands, optionCount, true)}
	}
}

func genHyperbolaGeneral(optionCount int) question {
	// b²(x - h)² - a²(y - k)² = a²b², c² = a² + b²
	t := pickOne([]pyTriple{{3, 4, 5}, {5, 12, 13}})
	a, b, c := t.a, t.b, t.c
	if rand.IntN(2) == 0 {
		a, b = b, a
	}
	h, k := between(-4, 4), between(-4, 4)
	a2, b2 := a*a, b*b
	eq := fmt.Sprintf("%dx² - %dy²%s%s%s = 0", b2, a2, linTerm(-2*b2*h, "x"), linTerm(2*a2*k, "y"), linTerm(b2*h*h-a2*k*k-a2*b2, ""))
	cands := []float64{float64(h), float64(h + a), float64(h + b), float64(h + c), float64(c), float64(h - c)}
	if rand.IntN(2) == 0 {
		return question{prompt: fmt.Sprintf("Koordinat x titik fokus sebelah kanan dari hiperbola %s adalah?", eq), options: buildOptions(float64(h+c), cands, optionCount, true)}
	}
	return question{prompt: fmt.Sprintf("Koordinat x titik puncak sebelah kanan dari hiperbola %s adalah?", eq), options: buildOptions(float64(h+a), cands, optionCount, true)}
}

// ---------- Matematika Keuangan ----------

func genAnnuityPrincipalN(optionCount int) question {
	for {
		i := pickOne([]int{5, 10, 20})
		steps := between(1, 2)
		a1 := between(1, 50) * 200 // ribu rupiah
		num := a1 * intPow(100+i, steps)
		den := intPow(100, steps)
		if num%den != 0 {
			continue
		}
		correct := float64(num / den)
		return question{
			prompt: fmt.Sprintf(
				"Sebuah pinjaman dilunasi dengan sistem anuitas. Angsuran pokok pertama Rp%d ribu dan suku bunga %d%% per bulan. Besar angsuran pokok ke-%d adalah? (ribu rupiah)",
				a1, i, steps+1,
			),
			options: buildOptions(correct, []float64{float64(a1 + a1*i*steps/100), float64(a1), round2(correct * float64(100+i) / 100), correct + 100}, optionCount, false),
		}
	}
}

func genPresentValue(optionCount int) question {
	r := pickOne([]int{10, 20})
	n := pickOne([]int{2, 3})
	m := between(1, 20) * 1000
	future := m * intPow(100+r, n) / intPow(100, n)
	correct := float64(m)
	return question{
		prompt: fmt.Sprintf(
			"Agar setelah %d tahun tabungan menjadi Rp%d ribu dengan bunga majemuk %d%% per tahun, modal awal yang harus disimpan adalah? (ribu rupiah)",
			n, future, r,
		),
		options: buildOptions(correct, []float64{float64(future - future*r*n/100), round2(float64(future) / (1 + float64(r*n)/100)), float64(future), correct + 100}, optionCount, false),
	}
}

func genDepreciationStraight(optionCount int) question {
	n := between(4, 10)
	dep := between(5, 50) * 100
	salvage := between(1, 20) * 500
	price := salvage + n*dep
	k := between(1, n-1)
	correct := float64(price - k*dep)
	return question{
		prompt: fmt.Sprintf(
			"Sebuah mesin dibeli seharga Rp%d ribu dengan umur ekonomis %d tahun dan nilai sisa Rp%d ribu. Dengan metode garis lurus, nilai buku mesin setelah %d tahun adalah? (ribu rupiah)",
			price, n, salvage, k,
		),
		options: buildOptions(correct, []float64{float64(dep), float64(price - (k+1)*dep), float64(salvage + k*dep), round2(float64(price) - float64(k*price)/float64(n))}, optionCount, false),
	}
}

func genDecliningBalance(optionCount int) question {
	r := pickOne([]int{10, 20})
	n := pickOne([]int{2, 3})
	price := between(1, 50) * 1000
	correct := float64(price * intPow(100-r, n) / intPow(100, n))
	return question{
		prompt: fmt.Sprintf(
			"Sebuah kendaraan dibeli seharga Rp%d ribu dan disusutkan dengan metode saldo menurun dengan tarif %d%% per tahun. Nilai buku kendaraan setelah %d tahun adalah? (ribu rupiah)",
			price, r, n,
		),
		options: buildOptions(correct, []float64{float64(price - price*r*n/100), float64(price * intPow(100-r, n-1) / intPow(100, n-1)), correct - 100, correct + 100}, optionCount, false),
	}
}

// ---------- Matriks & Program Linear Terapan ----------

func genSPLTVWord(optionCount int) question {
	prices := []int{between(2, 15), between(2, 15), between(2, 15)} // ribu rupiah
	items := []string{"buku", "pensil", "penghapus"}
	var m [][]int
	for {
		m = make([][]int, 3)
		for i := range m {
			m[i] = []int{between(1, 4), between(1, 4), between(1, 4)}
		}
		if det3(m) != 0 {
			break
		}
	}
	buyers := []string{"Andi", "Budi", "Citra"}
	parts := make([]string, 3)
	for i := range m {
		total := m[i][0]*prices[0] + m[i][1]*prices[1] + m[i][2]*prices[2]
		parts[i] = fmt.Sprintf("%s membeli %d buku, %d pensil, dan %d penghapus seharga Rp%d ribu", buyers[i], m[i][0], m[i][1], m[i][2], total)
	}
	idx := rand.IntN(3)
	correct := float64(prices[idx])
	return question{
		prompt:  fmt.Sprintf("%s. Harga 1 %s adalah? (ribu rupiah)", strings.Join(parts, ". "), items[idx]),
		options: buildOptions(correct, []float64{float64(prices[(idx+1)%3]), float64(prices[(idx+2)%3]), float64(prices[0] + prices[1] + prices[2]), correct + 1, correct - 1}, optionCount, false),
	}
}

func genLinearProgramMinWord(optionCount int) question {
	cons := randomLPConstraints()
	p, q := between(3, 12), between(3, 12)
	best, others := lpMinValues(cons, p, q)
	return question{
		prompt: fmt.Sprintf(
			"Seorang peternak mencampur pakan A dan B. Setiap kg pakan A mengandung %d unit protein dan %d unit karbohidrat, setiap kg pakan B mengandung %d unit protein dan %d unit karbohidrat. Ternak membutuhkan minimal %d unit protein dan %d unit karbohidrat. Harga pakan A Rp%d ribu/kg dan pakan B Rp%d ribu/kg. Biaya minimum adalah? (ribu rupiah)",
			cons[0].c, cons[1].c, cons[0].d, cons[1].d, cons[0].e, cons[1].e, p, q,
		),
		options: buildOptions(best, append(others, best+float64(p), best-1), optionCount, false),
	}
}

func genMatrixThreeProducts(optionCount int) question {
	qty := [][]int{{between(5, 30), between(5, 30), between(5, 30)}, {between(5, 30), between(5, 30), between(5, 30)}}
	price := []int{between(5, 30), between(5, 30), between(5, 30)}
	revenue := func(s int) int { return qty[s][0]*price[0] + qty[s][1]*price[1] + qty[s][2]*price[2] }
	rA, rB := revenue(0), revenue(1)
	correct := float64(absInt(rA - rB))
	return question{
		prompt: fmt.Sprintf(
			"Toko A menjual %d kg gula, %d kg beras, dan %d kg tepung. Toko B menjual %d kg gula, %d kg beras, dan %d kg tepung. Harga per kg gula Rp%d ribu, beras Rp%d ribu, dan tepung Rp%d ribu. Selisih pendapatan kedua toko adalah? (ribu rupiah)",
			qty[0][0], qty[0][1], qty[0][2], qty[1][0], qty[1][1], qty[1][2], price[0], price[1], price[2],
		),
		options: buildOptions(correct, []float64{float64(rA), float64(rB), float64(rA + rB), correct + float64(price[0])}, optionCount, false),
	}
}

// ---------- Trigonometri Terapan ----------

func genTwoElevations(optionCount int) question {
	// tan α = a/b (< 1) dari jauh, lalu maju w m -> elevasi 45°: h = a·w/(b - a)
	t := pickOne(pyTriples)
	mult := between(1, 6)
	w := (t.b - t.a) * mult
	h := t.a * mult
	correct := float64(h)
	return question{
		prompt: fmt.Sprintf(
			"Dari titik P, puncak sebuah menara terlihat dengan sudut elevasi α (tan α = %d/%d). Setelah berjalan %d m mendekati menara, puncaknya terlihat dengan sudut elevasi 45°. Jika tinggi pengamat diabaikan, tinggi menara adalah? (m)",
			t.a, t.b, w,
		),
		options: buildOptions(correct, []float64{float64(w), float64(t.b * mult), round2(float64(w*t.a) / float64(t.b)), correct + float64(w)}, optionCount, false),
	}
}

func genBearingDistance(optionCount int) question {
	t := pickOne(pyTriples)
	k := between(1, 3)
	west := between(1, 10)
	east := west + t.a*k
	north := t.b * k
	correct := float64(t.c * k)
	return question{
		prompt:  fmt.Sprintf("Sebuah kapal berlayar %d km ke timur, lalu %d km ke utara, kemudian %d km ke barat. Jarak kapal dari titik awal sekarang adalah? (km)", east, north, west),
		options: buildOptions(correct, []float64{float64(east + north + west), round2(math.Sqrt(float64(east*east + north*north))), correct + float64(west), correct + 1}, optionCount, false),
	}
}

func genCosineNavigation(optionCount int) question {
	// sudut dalam di titik belok = 180° - perubahan arah
	t := pickOne(angleTriangles)
	k := between(1, 4)
	a, b := t.a*k, t.b*k
	start := pickOne([]int{0, 30, 45, 60})
	turn := 180 - t.angle
	second := (start + turn) % 360
	correct := float64(t.c * k)
	return question{
		prompt: fmt.Sprintf(
			"Sebuah kapal berlayar dari pelabuhan sejauh %d km dengan arah %03d°, lalu berbelok dan berlayar sejauh %d km dengan arah %03d°. Jarak kapal dari pelabuhan sekarang adalah? (km)",
			a, start, b, second,
		),
		options: buildOptions(correct, []float64{float64(a + b), float64(absInt(a - b)), round2(math.Sqrt(float64(a*a + b*b))), correct + float64(k)}, optionCount, false),
	}
}

// ---------- Logika Matematika ----------

func negateLogic(e logicExpr) logicExpr {
	return logicExpr{"~(" + e.text + ")", func(v []bool) bool { return !e.eval(v) }}
}

func genTruthTableHard(optionCount int) question {
	vars := between(3, 4)
	var expr logicExpr
	if vars == 4 {
		left := parenthesize(combineLogic(randomLiteral(0), randomLiteral(1)))
		right := parenthesize(combineLogic(randomLiteral(2), randomLiteral(3)))
		if rand.IntN(2) == 0 {
			left = negateLogic(combineLogic(randomLiteral(0), randomLiteral(1)))
		}
		expr = combineLogic(left, right)
	} else {
		inner := negateLogic(combineLogic(randomLiteral(0), randomLiteral(1)))
		if rand.IntN(2) == 0 {
			expr = combineLogic(inner, randomLiteral(2))
		} else {
			expr = combineLogic(parenthesize(combineLogic(randomLiteral(2), randomLiteral(0))), inner)
		}
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
		options: buildOptions(float64(correct), []float64{float64(rows - correct), float64(correct + 1), float64(correct - 1), float64(rows), float64(correct + 2)}, optionCount, false),
	}
}
