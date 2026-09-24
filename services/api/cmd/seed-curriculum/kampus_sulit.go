package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"sort"
	"strings"
)

// ---------- Kampus: Campuran Sulit (batch 5) ----------
//
// Versi lebih sulit (multi-langkah) dari 25 materi kampus, dicampur rata.

func kampusHardMix() func(int) question {
	return mixOf(
		mixOf(genKHLHopital2, genKHFenceOpt, genKHNormalLine),
		mixOf(genKHSeriesNR, genKHGammaIntegral, genKHPolarArea),
		mixOf(genKHDoubleTriangle, genKHLagrange, genKHPolarIntegral),
		mixOf(genKHGreenDisk, genKHDivergenceSphere, genKHStokesCircle),
		mixOf(genKHEigen3, genKHMatrixPower, genKHCayleyHamilton),
		mixOf(genKHOrderProduct, genKHHomCount, genKHElementsOfOrder, genKHPermCompose),
		mixOf(genKHBoolean4, genKHOverflow8, genKHHammingGray),
		mixOf(genKHLimitE, genKHSqrtSeq, genKHRatioLimit),
		mixOf(genKHDoublePole, genKHContour3, genKHDeMoivre),
		mixOf(genKHNewtonTwo, genKHSimpson4, genKHGaussSeidel, genKHSecant),
		mixOf(genKHConstantForcing, genKHRepeatedRoot, genKHWronskian, genKHUndetermined),
		mixOf(genKHDAlembert, genKHHeatMode, genKHVarCoeffDisc),
		mixOf(genKHDerangement, genKHStirling2, genKHRecurrence, genKHCatalan),
		mixOf(genKHShortestPath6, genKHMST, genKHWalks, genKHPerfectMatchings),
		mixOf(genKHCRT3, genKHLastTwoDigits, genKHLinearCongruenceCount),
		mixOf(genKHCovTable, genKHVarSum, genKHPdfMean),
		mixOf(genKHSampleSize, genKHPooledVariance, genKHChiSquare, genKHFStat),
		mixOf(genKHMultipleReg, genKHSSE, genKHSlopeGeneral),
		mixOf(genKHSkewLines, genKHTetraVolume, genKHParallelPlanes),
		mixOf(genKHReflectRotate, genKHAreaImageComposite),
		mixOf(genKHConnectedSum, genKHGenusTriangulation),
		mixOf(genKHFourierX2, genKHConvolution),
		mixOf(genKHTransportation2x3, genKHAssignment4, genKHCPM),
		mixOf(genKHIRR, genKHBondPrice, genKHPVAnnuity),
		mixOf(genKHPCA2, genKHGradientDescent, genKHNormalEquation),
	)
}

// ---------- Kalkulus I ----------

func genKHLHopital2(optionCount int) question {
	for {
		a, b := nonZeroBetween(-6, 6), pickOne([]int{1, 2, 4, 5})
		val := float64(a*a) / float64(2*b)
		if !isExact2(val) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("lim (x→0) (e^(%s) - 1%s) / (%s) = ?", formatPoly(a, 0), linTerm(-a, "x"), formatTerms([]int{b}, []string{"x²"})),
			options: buildOptions(round2(val), []float64{round2(float64(a*a) / float64(b)), round2(float64(a) / float64(2*b)), 0, round2(val + 1)}, optionCount, false),
		}
	}
}

func genKHFenceOpt(optionCount int) question {
	// pagar 3 sisi (sisi ke-4 sungai): 2x + y = P, luas xy maksimum di x = P/4, luas P²/8
	p := 8 * between(5, 30)
	if rand.IntN(2) == 0 {
		correct := float64(p * p / 8)
		return question{
			prompt:  fmt.Sprintf("Kawat sepanjang %d m dipakai memagari tanah persegi panjang di tepi sungai (sisi tepi sungai tidak dipagar). Luas maksimum tanah yang bisa dipagari adalah? (m²)", p),
			options: buildOptions(correct, []float64{float64(p * p / 16), float64(p * p / 4), float64(p * p / 18), correct + float64(p)}, optionCount, false),
		}
	}
	correct := float64(p / 2)
	return question{
		prompt:  fmt.Sprintf("Kawat sepanjang %d m dipakai memagari tanah persegi panjang di tepi sungai (sisi tepi sungai tidak dipagar). Agar luasnya maksimum, panjang sisi yang sejajar sungai adalah? (m)", p),
		options: buildOptions(correct, []float64{float64(p / 4), float64(p / 3), float64(p), correct + 1}, optionCount, false),
	}
}

func genKHNormalLine(optionCount int) question {
	for {
		a, b, c := nonZeroBetween(-3, 3), between(-6, 6), between(-9, 9)
		k := between(-3, 3)
		m := 2*a*k + b
		if m == 0 {
			continue
		}
		fk := a*k*k + b*k + c
		intercept := float64(fk) + float64(k)/float64(m) // garis normal: y = fk - (x - k)/m
		if !isExact2(intercept) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Garis normal kurva y = %s di titik berabsis x = %d memotong sumbu y di titik (0, n). Nilai n = ?", formatPoly(a, b, c), k),
			options: buildOptions(round2(intercept), []float64{float64(fk - m*k), float64(fk), round2(float64(fk) - float64(k)/float64(m)), round2(intercept + 1)}, optionCount, true),
		}
	}
}

// ---------- Kalkulus II ----------

func genKHSeriesNR(optionCount int) question {
	for {
		k := between(2, 6)
		c := between(1, 5)
		val := float64(c*k) / float64((k-1)*(k-1)) // Σ c·n/kⁿ = c·k/(k-1)²
		if !isExact2(val) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Σ (n=1 sampai ∞) %dn / %d^n = ?", c, k),
			options: buildOptions(round2(val), []float64{round2(float64(c*k) / float64(k-1)), round2(float64(c) / float64(k-1)), round2(val + 1), round2(val * 2)}, optionCount, false),
		}
	}
}

func genKHGammaIntegral(optionCount int) question {
	for {
		n, a := between(1, 4), between(1, 5)
		val := float64(factorial(n)) / float64(intPow(a, n+1))
		if !isExact2(val) {
			continue
		}
		xn := "x"
		if n > 1 {
			xn = fmt.Sprintf("x^%d", n)
		}
		return question{
			prompt:  fmt.Sprintf("∫ dari 0 sampai ∞ %s·e^(%s) dx = ?", xn, formatPoly(-a, 0)),
			options: buildOptions(round2(val), []float64{float64(factorial(n)), round2(1 / float64(intPow(a, n+1))), round2(float64(factorial(n-1)) / float64(intPow(a, n))), round2(val + 1)}, optionCount, false),
		}
	}
}

func genKHPolarArea(optionCount int) question {
	a := between(1, 8)
	if rand.IntN(2) == 0 {
		// kardioid r = a(1 + cos θ): luas 3πa²/2
		correct := 1.5 * float64(a*a)
		return question{
			prompt:  fmt.Sprintf("Luas daerah di dalam kardioid r = %d(1 + cos θ) adalah kπ. Nilai k = ?", a),
			options: buildOptions(correct, []float64{float64(a * a), float64(2 * a * a), 0.5 * float64(a*a), correct + 1}, optionCount, false),
		}
	}
	// r = a cos θ: lingkaran diameter a, luas πa²/4
	correct := float64(a*a) / 4
	return question{
		prompt:  fmt.Sprintf("Luas daerah di dalam kurva polar r = %d cos θ adalah kπ. Nilai k = ?", a),
		options: buildOptions(correct, []float64{float64(a * a), float64(a*a) / 2, float64(a) / 2, correct + 1}, optionCount, false),
	}
}

// ---------- Kalkulus III ----------

func genKHDoubleTriangle(optionCount int) question {
	for {
		p, q, a := nonZeroBetween(-4, 5), nonZeroBetween(-4, 5), between(1, 4)
		a3 := float64(a * a * a)
		val := a3 * (float64(p)/3 + float64(q)/6) // ∫₀^a ∫₀^x (px + qy) dy dx
		if !isExact2(cleanFloat(val)) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("∫ (x: 0→%d) ∫ (y: 0→x) (%s) dy dx = ?", a, formatTerms([]int{p, q}, []string{"x", "y"})),
			options: buildOptions(round2(val), []float64{round2(a3 * (float64(p)/2 + float64(q)/2)), round2(a3 * (float64(p)/6 + float64(q)/3)), round2(val + 1), -round2(val)}, optionCount, true),
		}
	}
}

func genKHLagrange(optionCount int) question {
	if rand.IntN(2) == 0 {
		t := pickOne(pyTriples)
		a, b := t.a*randSign(), t.b*randSign()
		r := between(1, 5)
		correct := float64(t.c * r)
		return question{
			prompt:  fmt.Sprintf("Nilai maksimum f(x, y) = %s dengan kendala x² + y² = %d adalah?", formatTerms([]int{a, b}, []string{"x", "y"}), r*r),
			options: buildOptions(correct, []float64{float64((absInt(a) + absInt(b)) * r), float64(t.c * r * r), float64(r), correct + 1}, optionCount, false),
		}
	}
	s := 2 * between(2, 20)
	correct := float64(s * s / 4)
	return question{
		prompt:  fmt.Sprintf("Nilai maksimum f(x, y) = xy dengan kendala x + y = %d adalah?", s),
		options: buildOptions(correct, []float64{float64(s * s / 2), float64(s), float64(s*s/4 - 1), float64(s / 2)}, optionCount, false),
	}
}

func genKHPolarIntegral(optionCount int) question {
	r := between(1, 5)
	r4 := float64(intPow(r, 4))
	correct := r4 / 2 // ∬ (x² + y²) dA pada cakram jari-jari R = πR⁴/2
	return question{
		prompt:  fmt.Sprintf("∬ (x² + y²) dA pada cakram x² + y² ≤ %d bernilai kπ. Nilai k = ?", r*r),
		options: buildOptions(correct, []float64{r4, r4 / 4, float64(r * r), correct + 1}, optionCount, false),
	}
}

// ---------- Kalkulus Vektor ----------

func genKHGreenDisk(optionCount int) question {
	r, c := between(1, 4), between(1, 3)
	r4 := float64(intPow(r, 4))
	correct := 1.5 * float64(c) * r4 // ∮ -cy³ dx + cx³ dy = ∬ 3c(x² + y²) dA = 3cπR⁴/2
	return question{
		prompt: fmt.Sprintf(
			"Dengan teorema Green, ∮ (%s) dx + (%s) dy sepanjang lingkaran x² + y² = %d (berlawanan arah jarum jam) bernilai kπ. Nilai k = ?",
			formatTerms([]int{-c}, []string{"y³"}), formatTerms([]int{c}, []string{"x³"}), r*r,
		),
		options: buildOptions(correct, []float64{float64(3*c) * r4, float64(c) * r4 / 2, float64(3 * c * r * r), correct + 1}, optionCount, false),
	}
}

func genKHDivergenceSphere(optionCount int) question {
	for {
		a, b, c := between(1, 4), between(1, 4), between(1, 4)
		r := between(1, 3)
		val := 4 * float64((a+b+c)*r*r*r) / 3 // (a+b+c)·(4/3)πR³
		if !isExact2(val) {
			continue
		}
		return question{
			prompt: fmt.Sprintf(
				"Fluks F = (%s, %s, %s) keluar dari bola x² + y² + z² = %d bernilai kπ. Nilai k = ?",
				formatTerms([]int{a}, []string{"x"}), formatTerms([]int{b}, []string{"y"}), formatTerms([]int{c}, []string{"z"}), r*r,
			),
			options: buildOptions(round2(val), []float64{float64(4 * (a + b + c) * r * r), float64((a + b + c) * r * r * r), round2(4 * float64(r*r*r) / 3), round2(val + 1)}, optionCount, false),
		}
	}
}

func genKHStokesCircle(optionCount int) question {
	k, r := between(1, 5), between(1, 5)
	correct := float64(2 * k * r * r) // ∮ -ky dx + kx dy = 2kπR²
	return question{
		prompt: fmt.Sprintf(
			"Sirkulasi medan F = (%s, %s, 0) sepanjang lingkaran x² + y² = %d pada bidang z = 0 (berlawanan arah jarum jam) bernilai kπ. Nilai k = ?",
			formatTerms([]int{-k}, []string{"y"}), formatTerms([]int{k}, []string{"x"}), r*r,
		),
		options: buildOptions(correct, []float64{float64(k * r * r), float64(2 * k * r), 0, correct + 1}, optionCount, false),
	}
}

// ---------- Aljabar Linear ----------

// unimodular3: P = L·U (segitiga satuan) dengan det 1, plus inversnya (adjoin).
func unimodular3() (p, pinv [][]int) {
	l := [][]int{{1, 0, 0}, {between(-1, 1), 1, 0}, {between(-1, 1), between(-1, 1), 1}}
	u := [][]int{{1, between(-1, 1), between(-1, 1)}, {0, 1, between(-1, 1)}, {0, 0, 1}}
	p = matMulInt(l, u)
	pinv = make([][]int, 3)
	for i := range pinv {
		pinv[i] = make([]int, 3)
	}
	for i := range 3 {
		for j := range 3 {
			var minor [][]int
			for r := range 3 {
				if r == j {
					continue
				}
				var row []int
				for c := range 3 {
					if c != i {
						row = append(row, p[r][c])
					}
				}
				minor = append(minor, row)
			}
			cof := minor[0][0]*minor[1][1] - minor[0][1]*minor[1][0]
			if (i+j)%2 == 1 {
				cof = -cof
			}
			pinv[i][j] = cof // det(P) = 1 -> P⁻¹ = adj(P)
		}
	}
	return p, pinv
}

func genKHEigen3(optionCount int) question {
	eig := rand.Perm(9)[:3]
	for i := range eig {
		eig[i] -= 3 // -3..5, beda semua
	}
	p, pinv := unimodular3()
	a := matMulInt(matMulInt(p, diagInt(eig...)), pinv)
	sorted := append([]int{}, eig...)
	sort.Ints(sorted)
	prod := eig[0] * eig[1] * eig[2]
	type variant struct {
		label   string
		correct int
	}
	v := pickOne([]variant{{"Nilai eigen terbesar", sorted[2]}, {"Nilai eigen terkecil", sorted[0]}, {"Hasil kali semua nilai eigen", prod}})
	return question{
		prompt:  fmt.Sprintf("Diketahui A = %s (baris dipisah titik koma). %s dari A adalah?", formatMatrix(a), v.label),
		options: buildOptions(float64(v.correct), []float64{float64(sorted[0]), float64(sorted[1]), float64(sorted[2]), float64(prod), float64(eig[0] + eig[1] + eig[2])}, optionCount, true),
	}
}

func genKHMatrixPower(optionCount int) question {
	l1 := pickOne([]int{-2, -1, 1, 2, 3})
	l2 := pickOne([]int{-2, -1, 1, 2, 3})
	for l2 == l1 {
		l2 = pickOne([]int{-2, -1, 1, 2, 3})
	}
	n := between(3, 6)
	p, pinv := unimodular2()
	a := matMulInt(matMulInt(p, diagInt(l1, l2)), pinv)
	an := matMulInt(matMulInt(p, diagInt(intPow(l1, n), intPow(l2, n))), pinv)
	an1 := matMulInt(matMulInt(p, diagInt(intPow(l1, n-1), intPow(l2, n-1))), pinv)
	i, j := rand.IntN(2), rand.IntN(2)
	correct := float64(an[i][j])
	return question{
		prompt:  fmt.Sprintf("Diketahui A = %s (baris dipisah titik koma). Elemen baris ke-%d kolom ke-%d dari A^%d adalah?", formatMatrix(a), i+1, j+1, n),
		options: buildOptions(correct, []float64{float64(intPow(a[i][j], n)), float64(n * a[i][j]), float64(an1[i][j]), -correct, correct + 1}, optionCount, true),
	}
}

func genKHCayleyHamilton(optionCount int) question {
	a := [][]int{{between(-5, 6), between(-5, 6)}, {between(-5, 6), between(-5, 6)}}
	tr, det := a[0][0]+a[1][1], a[0][0]*a[1][1]-a[0][1]*a[1][0]
	if rand.IntN(2) == 0 {
		return question{
			prompt:  fmt.Sprintf("Menurut teorema Cayley-Hamilton, A² = αA + βI untuk A = %s (baris dipisah titik koma). Nilai α = ?", formatMatrix(a)),
			options: buildOptions(float64(tr), []float64{float64(-tr), float64(det), float64(-det), float64(tr + 1)}, optionCount, true),
		}
	}
	return question{
		prompt:  fmt.Sprintf("Menurut teorema Cayley-Hamilton, A² = αA + βI untuk A = %s (baris dipisah titik koma). Nilai β = ?", formatMatrix(a)),
		options: buildOptions(float64(-det), []float64{float64(det), float64(tr), float64(-tr), float64(-det + 1)}, optionCount, true),
	}
}

// ---------- Aljabar Abstrak ----------

func genKHOrderProduct(optionCount int) question {
	m, n := between(2, 12), between(2, 12)
	a, b := between(0, m-1), between(0, n-1)
	for a == 0 && b == 0 {
		b = between(0, n-1)
	}
	oa, ob := m/gcd(m, a), n/gcd(n, b)
	correct := float64(lcm(oa, ob))
	return question{
		prompt:  fmt.Sprintf("Orde unsur (%d, %d) di grup Z_%d × Z_%d adalah?", a, b, m, n),
		options: buildOptions(correct, []float64{float64(oa * ob), float64(max(oa, ob)), float64(m * n), float64(lcm(m, n))}, optionCount, false),
	}
}

func genKHHomCount(optionCount int) question {
	m, n := between(2, 30), between(2, 30)
	return question{
		prompt:  fmt.Sprintf("Banyak homomorfisma grup dari Z_%d ke Z_%d adalah?", m, n),
		options: buildOptions(float64(gcd(m, n)), []float64{float64(lcm(m, n)), float64(m), float64(n), 1}, optionCount, false),
	}
}

func genKHElementsOfOrder(optionCount int) question {
	n := between(6, 60)
	var divs []int
	for d := 2; d <= n; d++ {
		if n%d == 0 {
			divs = append(divs, d)
		}
	}
	k := pickOne(divs)
	return question{
		prompt:  fmt.Sprintf("Banyak unsur berorde %d di grup siklik Z_%d adalah?", k, n),
		options: buildOptions(float64(eulerPhi(k)), []float64{float64(k), float64(n / k), float64(eulerPhi(n)), float64(k - 1)}, optionCount, false),
	}
}

func cycleNotation(p []int) string {
	seen := make([]bool, len(p))
	var parts []string
	for i := range p {
		if seen[i] || p[i] == i {
			seen[i] = true
			continue
		}
		var cyc []string
		for j := i; !seen[j]; j = p[j] {
			seen[j] = true
			cyc = append(cyc, fmt.Sprintf("%d", j+1))
		}
		parts = append(parts, "("+strings.Join(cyc, " ")+")")
	}
	if len(parts) == 0 {
		return "e"
	}
	return strings.Join(parts, "")
}

func permOrder(p []int) int {
	seen := make([]bool, len(p))
	order := 1
	for i := range p {
		if seen[i] {
			continue
		}
		l := 0
		for j := i; !seen[j]; j = p[j] {
			seen[j] = true
			l++
		}
		order = lcm(order, l)
	}
	return order
}

func genKHPermCompose(optionCount int) question {
	for {
		n := 6
		s, t := rand.Perm(n), rand.Perm(n)
		comp := make([]int, n) // σ ∘ τ: τ dulu, lalu σ
		for i := range comp {
			comp[i] = s[t[i]]
		}
		os, ot, oc := permOrder(s), permOrder(t), permOrder(comp)
		if os == 1 || ot == 1 {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Di S_6, σ = %s dan τ = %s. Orde dari σ ∘ τ (τ dikerjakan lebih dulu) adalah?", cycleNotation(s), cycleNotation(t)),
			options: buildOptions(float64(oc), []float64{float64(os), float64(ot), float64(lcm(os, ot)), float64(os * ot), float64(oc + 1)}, optionCount, false),
		}
	}
}

// ---------- Aljabar Boolean ----------

func negateBool(e logicExpr) logicExpr {
	return logicExpr{"(" + e.text + ")'", func(v []bool) bool { return !e.eval(v) }}
}

func genKHBoolean4(optionCount int) question {
	left := combineBool(boolLiteral(0), boolLiteral(1))
	right := combineBool(boolLiteral(2), boolLiteral(3))
	var l, r logicExpr
	if rand.IntN(2) == 0 {
		l = negateBool(left)
	} else {
		l = parenthesize(left)
	}
	if rand.IntN(2) == 0 {
		r = negateBool(right)
	} else {
		r = parenthesize(right)
	}
	expr := combineBool(l, r)
	correct := countTrueRows(expr, 4)
	return question{
		prompt:  fmt.Sprintf("Banyak minterm dari fungsi Boolean F(A, B, C, D) = %s adalah?", expr.text),
		options: buildOptions(float64(correct), []float64{float64(16 - correct), float64(correct + 1), float64(correct - 1), float64(correct + 2), 16}, optionCount, false),
	}
}

func genKHOverflow8(optionCount int) question {
	a, b := between(-128, 127), between(-128, 127)
	if rand.IntN(2) == 0 { // paksa overflow
		a, b = between(64, 127), between(64, 127)
		if rand.IntN(2) == 0 {
			a, b = -a-1, -b
		}
	}
	stored := int(int8(a + b))
	bText := fmt.Sprintf("%d", b)
	if b < 0 {
		bText = "(" + bText + ")"
	}
	return question{
		prompt:  fmt.Sprintf("Pada penjumlahan 8-bit komplemen dua, %d + %s menghasilkan nilai (desimal) yang tersimpan di register adalah?", a, bText),
		options: buildOptions(float64(stored), []float64{float64(a + b), float64(uint8(int8(a + b))), float64(-stored), float64(stored + 1)}, optionCount, true),
	}
}

func genKHHammingGray(optionCount int) question {
	switch rand.IntN(3) {
	case 0:
		x, y := between(0, 255), between(0, 255)
		dist := 0
		for d := x ^ y; d > 0; d >>= 1 {
			dist += d & 1
		}
		return question{prompt: fmt.Sprintf("Jarak Hamming antara %08b dan %08b adalah?", x, y), options: buildOptions(float64(dist), []float64{float64(8 - dist), float64(dist + 1), float64(dist - 1), float64(absInt(x-y) % 9)}, optionCount, false)}
	case 1:
		n := between(1, 255)
		g := n ^ (n >> 1)
		return question{prompt: fmt.Sprintf("Kode Gray dari bilangan %d, jika dibaca sebagai bilangan biner biasa, bernilai (desimal)?", n), options: buildOptions(float64(g), []float64{float64(n), float64(n >> 1), float64(n ^ (n<<1)&255), float64(g + 1)}, optionCount, false)}
	default:
		g := between(1, 255)
		n := 0
		for x := g; x > 0; x >>= 1 {
			n ^= x
		}
		return question{prompt: fmt.Sprintf("Kode Gray %08b jika dikonversi ke bilangan biner biasa bernilai (desimal)?", g), options: buildOptions(float64(n), []float64{float64(g), float64(g ^ (g >> 1)), float64(n + 1), float64(g >> 1)}, optionCount, false)}
	}
}

// ---------- Analisis Real ----------

func genKHLimitE(optionCount int) question {
	a, b := nonZeroBetween(-5, 5), nonZeroBetween(-4, 5)
	correct := float64(a * b)
	sign := "+"
	if a < 0 {
		sign = "-"
	}
	return question{
		prompt:  fmt.Sprintf("lim (n→∞) (1 %s %d/n)^(%s) = e^k. Nilai k = ?", sign, absInt(a), formatPolyVar("n", b, 0)),
		options: buildOptions(correct, []float64{float64(a + b), float64(a), float64(b), -correct}, optionCount, true),
	}
}

func genKHSqrtSeq(optionCount int) question {
	a := between(-10, 10)
	b := between(-10, 10)
	for b == a {
		b = between(-10, 10)
	}
	c, d := between(-9, 9), between(-9, 9)
	correct := float64(a-b) / 2
	return question{
		prompt:  fmt.Sprintf("lim (n→∞) (√(%s) - √(%s)) = ?", formatPolyVar("n", 1, a, c), formatPolyVar("n", 1, b, d)),
		options: buildOptions(correct, []float64{float64(a - b), float64(a+b) / 2, 0, correct + 1}, optionCount, true),
	}
}

func genKHRatioLimit(optionCount int) question {
	switch rand.IntN(3) {
	case 0:
		c, p := pickOne([]int{2, 4, 5, 10}), between(1, 5)
		correct := 1 / float64(c)
		return question{prompt: fmt.Sprintf("Untuk a_n = n^%d / %d^n, nilai lim (n→∞) a_(n+1)/a_n adalah?", p, c), options: buildOptions(correct, []float64{float64(c), 1, 0, float64(p) / float64(c)}, optionCount, false)}
	case 1:
		c := between(2, 9)
		return question{prompt: fmt.Sprintf("Untuk a_n = %d^n / n!, nilai lim (n→∞) a_(n+1)/a_n adalah?", c), options: buildOptions(0, []float64{float64(c), 1, round2(1 / float64(c)), 0.5}, optionCount, false)}
	default:
		return question{prompt: "Untuk a_n = (n!)² / (2n)!, nilai lim (n→∞) a_(n+1)/a_n adalah?", options: buildOptions(0.25, []float64{0.5, 1, 0, 4, 2}, optionCount, false)}
	}
}

// ---------- Analisis Kompleks ----------

func genKHDoublePole(optionCount int) question {
	a, p, q := nonZeroBetween(-5, 5), between(-6, 6), between(-9, 9)
	correct := float64(2*a + p) // Res [(z² + pz + q)/(z - a)²] = d/dz(z² + pz + q) di z = a
	return question{
		prompt:  fmt.Sprintf("Residu f(z) = (%s) / (%s)² di z = %d adalah?", formatPolyVar("z", 1, p, q), formatPolyVar("z", 1, -a), a),
		options: buildOptions(correct, []float64{float64(a*a + p*a + q), float64(p), float64(2 * a), correct + 1}, optionCount, true),
	}
}

func genKHContour3(optionCount int) question {
	for {
		poles := rand.Perm(9)[:3]
		for i := range poles {
			poles[i] -= 4 // -4..4
		}
		r := between(1, 4)
		k := between(1, 12)
		conflict, anyInside := false, false
		inside, all := 0.0, 0.0
		for i, pole := range poles {
			if absInt(pole) == r {
				conflict = true
			}
			den := 1
			for j, other := range poles {
				if j != i {
					den *= pole - other
				}
			}
			res := float64(k) / float64(den)
			if !isExact2(res) {
				conflict = true
			}
			all += res
			if absInt(pole) < r {
				inside += res
				anyInside = true
			}
		}
		if conflict || !anyInside {
			continue
		}
		correct := round2(2 * inside)
		factors := make([]string, 3)
		for i, pole := range poles {
			factors[i] = "(" + formatPolyVar("z", 1, -pole) + ")"
		}
		return question{
			prompt:  fmt.Sprintf("∮ %d / (%s) dz sepanjang lingkaran |z| = %d (berlawanan arah jarum jam) bernilai kπi. Nilai k = ?", k, strings.Join(factors, ""), r),
			options: buildOptions(correct, []float64{round2(inside), round2(2 * all), -correct, round2(correct + 1)}, optionCount, true),
		}
	}
}

func genKHDeMoivre(optionCount int) question {
	a := between(1, 2) * randSign()
	b := between(1, 2) * randSign()
	if rand.IntN(2) == 0 {
		b = a * randSign() // bentuk a ± ai
	}
	n := between(4, 8)
	re, im := complexPow(a, b, n)
	part, correct, other := "real", float64(re), float64(im)
	if rand.IntN(2) == 0 {
		part, correct, other = "imajiner", float64(im), float64(re)
	}
	return question{
		prompt:  fmt.Sprintf("Bagian %s dari (%s)^%d adalah?", part, fmtComplex(a, b), n),
		options: buildOptions(correct, []float64{other, -correct, float64(intPow(a, n) + intPow(b, n)), correct + 1}, optionCount, true),
	}
}

// ---------- Analisis Numerik ----------

func genKHNewtonTwo(optionCount int) question {
	for {
		a, x0 := between(2, 60), between(1, 8)
		x1 := cleanFloat(float64(x0*x0+a) / float64(2*x0))
		x2 := cleanFloat((x1*x1 + float64(a)) / (2 * x1))
		if !isExact2(x1) || !isExact2(x2) || x1 == x2 {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Metode Newton-Raphson untuk f(x) = %s dengan x₀ = %d. Nilai x₂ adalah?", formatPoly(1, 0, -a), x0),
			options: buildOptions(round2(x2), []float64{round2(x1), round2(math.Sqrt(float64(a))), float64(x0), round2(x2 + 0.5)}, optionCount, false),
		}
	}
}

func genKHSimpson4(optionCount int) question {
	for {
		a, b, c := between(-3, 3), between(-4, 4), between(-6, 6)
		if a == 0 && b == 0 {
			continue
		}
		f := func(x int) int { return a*x*x + b*x + c }
		s := cleanFloat((float64(f(0)+4*f(1)+2*f(2)+4*f(3)+f(4)) / 3))
		if !isExact2(s) {
			continue
		}
		trap := 0.5*float64(f(0)+f(4)) + float64(f(1)+f(2)+f(3))
		return question{
			prompt:  fmt.Sprintf("Aturan Simpson 1/3 komposit dengan n = 4 (h = 1) untuk ∫ dari 0 sampai 4 (%s) dx menghasilkan?", formatPoly(a, b, c)),
			options: buildOptions(round2(s), []float64{trap, float64(f(0) + 4*f(1) + 2*f(2) + 4*f(3) + f(4)), round2(s + 1), -round2(s)}, optionCount, true),
		}
	}
}

func genKHGaussSeidel(optionCount int) question {
	for {
		a11, a22 := between(4, 10), between(4, 10)
		a12, a21 := between(-3, 3), between(-3, 3)
		b1, b2 := between(-20, 30), between(-20, 30)
		x1 := float64(b1) / float64(a11)
		y1 := cleanFloat((float64(b2) - float64(a21)*x1) / float64(a22))
		if !isExact2(x1) || !isExact2(y1) {
			continue
		}
		jacobiY := float64(b2) / float64(a22)
		return question{
			prompt: fmt.Sprintf(
				"Sistem %s = %d dan %s = %d diselesaikan dengan metode Gauss-Seidel dari tebakan awal (0, 0). Nilai y setelah iterasi pertama adalah?",
				formatTerms([]int{a11, a12}, []string{"x", "y"}), b1, formatTerms([]int{a21, a22}, []string{"x", "y"}), b2,
			),
			options: buildOptions(round2(y1), []float64{round2(jacobiY), round2(x1), -round2(y1), round2(y1 + 1)}, optionCount, true),
		}
	}
}

func genKHSecant(optionCount int) question {
	for {
		a := between(2, 60)
		x0, x1 := between(1, 6), between(1, 8)
		if x0 == x1 || x0+x1 == 0 {
			continue
		}
		x2 := cleanFloat(float64(x1) - float64(x1*x1-a)/float64(x1+x0)) // f(x) = x² - a
		if !isExact2(x2) {
			continue
		}
		newton := float64(x1*x1+a) / float64(2*x1)
		return question{
			prompt:  fmt.Sprintf("Metode secant untuk f(x) = %s dengan x₀ = %d dan x₁ = %d. Nilai x₂ adalah?", formatPoly(1, 0, -a), x0, x1),
			options: buildOptions(round2(x2), []float64{round2(newton), float64(x1), round2(float64(x0+x1) / 2), round2(x2 + 0.5)}, optionCount, true),
		}
	}
}

// ---------- PD Biasa ----------

func genKHConstantForcing(optionCount int) question {
	for {
		p, q, c := between(-5, 5), nonZeroBetween(-6, 6), nonZeroBetween(-30, 30)
		val := float64(c) / float64(q)
		if !isExact2(val) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Solusi khusus konstan dari PD %s = %d adalah y_p = ?", odeString(p, q), c),
			options: buildOptions(round2(val), []float64{float64(c), -round2(val), round2(float64(c) / float64(absInt(q)+1)), round2(val + 1)}, optionCount, true),
		}
	}
}

func genKHRepeatedRoot(optionCount int) question {
	r := nonZeroBetween(-4, 4)
	a, b := between(-5, 5), between(-9, 9)
	correct := float64(b - r*a) // y = (C₁ + C₂x)e^{rx}: C₁ = y(0), C₂ = y'(0) - r·y(0)
	return question{
		prompt: fmt.Sprintf(
			"PD %s = 0 dengan y(0) = %d dan y'(0) = %d memiliki solusi y = (C₁ + C₂x)e^(%s). Nilai C₂ = ?",
			odeString(-2*r, r*r), a, b, formatPoly(r, 0),
		),
		options: buildOptions(correct, []float64{float64(b), float64(b + r*a), float64(a), -correct}, optionCount, true),
	}
}

func genKHWronskian(optionCount int) question {
	if rand.IntN(2) == 0 {
		r1, r2 := between(-5, 5), between(-5, 5)
		for r2 == r1 {
			r2 = between(-5, 5)
		}
		correct := float64(r2 - r1)
		return question{
			prompt:  fmt.Sprintf("Wronskian W(e^(%s), e^(%s)) di x = 0 adalah?", formatPoly(r1, 0), formatPoly(r2, 0)),
			options: buildOptions(correct, []float64{-correct, float64(r1 * r2), float64(r1 + r2), 1}, optionCount, true),
		}
	}
	a, b := between(1, 5), between(1, 5)
	for b == a {
		b = between(1, 5)
	}
	correct := float64(b - a) // W(x^a, x^b)(1) = b - a
	return question{
		prompt:  fmt.Sprintf("Wronskian W(x^%d, x^%d) di x = 1 adalah?", a, b),
		options: buildOptions(correct, []float64{-correct, float64(a * b), float64(a + b), 1}, optionCount, true),
	}
}

func genKHUndetermined(optionCount int) question {
	for {
		q, k, m := nonZeroBetween(-6, 6), nonZeroBetween(-12, 12), between(-12, 12)
		A, B := float64(k)/float64(q), float64(m)/float64(q)
		if !isExact2(A) || !isExact2(B) {
			continue
		}
		target, correct, other := "A", round2(A), round2(B)
		if rand.IntN(2) == 0 {
			target, correct, other = "B", round2(B), round2(A)
		}
		return question{
			prompt:  fmt.Sprintf("Solusi khusus PD y''%s = %s berbentuk y_p = Ax + B. Nilai %s = ?", linTerm(q, "y"), formatPoly(k, m), target),
			options: buildOptions(correct, []float64{other, -correct, float64(k), float64(m)}, optionCount, true),
		}
	}
}

// ---------- PD Parsial ----------

func genKHDAlembert(optionCount int) question {
	c := between(1, 3)
	a, b := nonZeroBetween(-3, 3), between(-5, 5)
	f := func(x int) int { return a*x*x + b*x }
	x0, t0 := between(-3, 3), between(1, 3)
	correct := float64(f(x0-c*t0)+f(x0+c*t0)) / 2
	return question{
		prompt: fmt.Sprintf(
			"Persamaan gelombang u_tt = %s dengan u(x, 0) = %s dan u_t(x, 0) = 0. Nilai u(%d, %d) adalah?",
			formatTerms([]int{c * c}, []string{"u_xx"}), formatPoly(a, b, 0), x0, t0,
		),
		options: buildOptions(correct, []float64{float64(f(x0)), float64(f(x0 + c*t0)), float64(f(x0 - c*t0)), correct + 1}, optionCount, true),
	}
}

func genKHHeatMode(optionCount int) question {
	alpha, n := between(1, 5), between(2, 6)
	a, b := between(1, 5), between(1, 5)
	correct := float64(alpha * n * n)
	return question{
		prompt: fmt.Sprintf(
			"Persamaan panas u_t = %s pada 0 < x < π dengan u(0, t) = u(π, t) = 0 dan u(x, 0) = %d sin(x) + %d %s. Suku %s pada solusinya meluruh dengan faktor e^(-kt). Nilai k = ?",
			formatTerms([]int{alpha}, []string{"u_xx"}), a, b, sinTerm(n, "x"), sinTerm(n, "x"),
		),
		options: buildOptions(correct, []float64{float64(alpha), float64(alpha * n), float64(n * n), float64(b * n * n)}, optionCount, false),
	}
}

func genKHVarCoeffDisc(optionCount int) question {
	a, b := nonZeroBetween(-3, 3), nonZeroBetween(-3, 3)
	x0, y0 := between(-3, 3), between(-3, 3)
	B, C := a*x0, b*y0
	correct := float64(B*B - 4*C) // A = 1
	return question{
		prompt: fmt.Sprintf(
			"Nilai diskriminan B² - 4AC dari PD u_xx + (%s)u_xy + (%s)u_yy = 0 di titik (%d, %d) adalah?",
			formatPoly(a, 0), formatPolyVar("y", b, 0), x0, y0,
		),
		options: buildOptions(correct, []float64{float64(B*B + 4*C), float64(B - 4*C), float64(4*C - B*B), correct + 1}, optionCount, true),
	}
}

// ---------- Matematika Diskrit ----------

func genKHDerangement(optionCount int) question {
	n := between(3, 8)
	d := []int{1, 0}
	for i := 2; i <= n; i++ {
		d = append(d, (i-1)*(d[i-1]+d[i-2]))
	}
	correct := float64(d[n])
	return question{
		prompt:  fmt.Sprintf("Banyak cara memasukkan %d surat ke %d amplop beralamat berbeda sehingga tidak ada satu pun surat yang masuk ke amplop yang benar adalah?", n, n),
		options: buildOptions(correct, []float64{float64(factorial(n)), float64(factorial(n) - 1), float64(d[n-1]), float64(factorial(n - 1))}, optionCount, false),
	}
}

func stirling2(n, k int) int {
	s := make([][]int, n+1)
	for i := range s {
		s[i] = make([]int, k+1)
	}
	s[0][0] = 1
	for i := 1; i <= n; i++ {
		for j := 1; j <= k; j++ {
			s[i][j] = j*s[i-1][j] + s[i-1][j-1]
		}
	}
	return s[n][k]
}

func genKHStirling2(optionCount int) question {
	n := between(4, 7)
	k := between(2, n-1)
	correct := float64(stirling2(n, k))
	return question{
		prompt:  fmt.Sprintf("Banyak cara mempartisi himpunan %d unsur menjadi %d himpunan bagian tak kosong (bilangan Stirling jenis kedua) adalah?", n, k),
		options: buildOptions(correct, []float64{float64(combination(n, k)), correct * float64(factorial(k)), float64(stirling2(n, k-1)), correct + 1}, optionCount, false),
	}
}

func genKHRecurrence(optionCount int) question {
	c1, c2 := between(1, 3), between(-2, 3)
	a0, a1 := between(0, 5), between(0, 5)
	n := between(5, 8)
	seq := []int{a0, a1}
	for i := 2; i <= n+1; i++ {
		seq = append(seq, c1*seq[i-1]+c2*seq[i-2])
	}
	correct := float64(seq[n])
	return question{
		prompt: fmt.Sprintf(
			"Barisan didefinisikan a_n = %s dengan a₀ = %d dan a₁ = %d. Nilai a_%d adalah?",
			formatTerms([]int{c1, c2}, []string{"a_(n-1)", "a_(n-2)"}), a0, a1, n,
		),
		options: buildOptions(correct, []float64{float64(seq[n-1]), float64(seq[n+1]), correct + 1, -correct}, optionCount, true),
	}
}

func genKHCatalan(optionCount int) question {
	n := between(3, 9)
	catalan := func(k int) int { return combination(2*k, k) / (k + 1) }
	correct := float64(catalan(n))
	cands := []float64{float64(combination(2*n, n)), float64(catalan(n - 1)), float64(catalan(n + 1)), float64(intPow(2, n))}
	switch rand.IntN(3) {
	case 0:
		return question{prompt: fmt.Sprintf("Banyak cara menyusun %d pasang tanda kurung yang seimbang adalah?", n), options: buildOptions(correct, cands, optionCount, false)}
	case 1:
		return question{prompt: fmt.Sprintf("Banyak pohon biner (terurut) berbeda dengan %d simpul adalah?", n), options: buildOptions(correct, cands, optionCount, false)}
	default:
		return question{prompt: fmt.Sprintf("Banyak cara triangulasi poligon konveks bersisi %d adalah?", n+2), options: buildOptions(correct, cands, optionCount, false)}
	}
}

// ---------- Teori Graf ----------

func genKHShortestPath6(optionCount int) question {
	n := 6
	edges := randomWeightedGraph(n, between(3, 6))
	d := shortestDistances(n, edges)
	correct := float64(d[0][n-1])
	return question{
		prompt:  fmt.Sprintf("Graf berbobot tak berarah dengan sisi (bobot): %s. Panjang lintasan terpendek dari A ke F adalah?", formatEdges(edges)),
		options: buildOptions(correct, []float64{correct + 1, correct + 2, correct - 1, float64(mstWeight(n, edges)), float64(d[0][n-2])}, optionCount, false),
	}
}

func genKHMST(optionCount int) question {
	n := between(5, 6)
	edges := randomWeightedGraph(n, between(3, 6))
	correct := float64(mstWeight(n, edges))
	total := 0
	for _, e := range edges {
		total += e.w
	}
	return question{
		prompt:  fmt.Sprintf("Graf berbobot tak berarah dengan sisi (bobot): %s. Bobot total pohon merentang minimum (MST) adalah?", formatEdges(edges)),
		options: buildOptions(correct, []float64{float64(total), correct + 1, correct + 2, correct - 1}, optionCount, false),
	}
}

func genKHWalks(optionCount int) question {
	for {
		n := 4
		adj := make([][]int, n)
		for i := range adj {
			adj[i] = make([]int, n)
		}
		var edgeText []string
		for i := range n {
			for j := i + 1; j < n; j++ {
				if rand.IntN(2) == 0 {
					adj[i][j], adj[j][i] = 1, 1
					edgeText = append(edgeText, fmt.Sprintf("{%d,%d}", i+1, j+1))
				}
			}
		}
		if len(edgeText) < 3 {
			continue
		}
		k := between(2, 4)
		power := adj
		for range k - 1 {
			power = matMulInt(power, adj)
		}
		i, j := rand.IntN(n), rand.IntN(n)
		correct := float64(power[i][j])
		if correct == 0 {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Graf dengan simpul {1, 2, 3, 4} dan sisi %s. Banyak jalan (walk) dengan panjang %d dari simpul %d ke simpul %d adalah?", strings.Join(edgeText, ", "), k, i+1, j+1),
			options: buildOptions(correct, []float64{correct + 1, correct - 1, correct + 2, float64(adj[i][j]), float64(len(edgeText))}, optionCount, false),
		}
	}
}

func genKHPerfectMatchings(optionCount int) question {
	n := between(2, 5)
	if rand.IntN(2) == 0 {
		correct := 1
		for k := 2*n - 1; k > 1; k -= 2 {
			correct *= k
		}
		return question{prompt: fmt.Sprintf("Banyak perfect matching pada graf lengkap K_%d adalah?", 2*n), options: buildOptions(float64(correct), []float64{float64(factorial(n)), float64(factorial(2 * n)), float64(combination(2*n, 2)), float64(correct * 2)}, optionCount, false)}
	}
	return question{prompt: fmt.Sprintf("Banyak perfect matching pada graf bipartit lengkap K_%d,%d adalah?", n, n), options: buildOptions(float64(factorial(n)), []float64{float64(n * n), float64(intPow(2, n)), float64(factorial(n - 1)), float64(factorial(n) + 1)}, optionCount, false)}
}

// ---------- Teori Bilangan ----------

func genKHCRT3(optionCount int) question {
	moduli := []int{3, 4, 5, 7, 11}
	rand.Shuffle(len(moduli), func(i, j int) { moduli[i], moduli[j] = moduli[j], moduli[i] })
	m := moduli[:3]
	r := []int{between(0, m[0]-1), between(0, m[1]-1), between(0, m[2]-1)}
	prod := m[0] * m[1] * m[2]
	x := 1
	for ; x <= prod; x++ {
		if x%m[0] == r[0] && x%m[1] == r[1] && x%m[2] == r[2] {
			break
		}
	}
	return question{
		prompt:  fmt.Sprintf("Bilangan bulat positif terkecil x yang memenuhi x ≡ %d (mod %d), x ≡ %d (mod %d), dan x ≡ %d (mod %d) adalah?", r[0], m[0], r[1], m[1], r[2], m[2]),
		options: buildOptions(float64(x), []float64{float64(x + prod), float64(prod), float64(r[0] + r[1] + r[2]), float64(x + 1)}, optionCount, false),
	}
}

func genKHLastTwoDigits(optionCount int) question {
	a, b := between(2, 99), between(100, 5000)
	correct := modPow(a, b, 100)
	return question{
		prompt:  fmt.Sprintf("Dua angka terakhir dari %d^%d, dibaca sebagai bilangan (00-99), adalah?", a, b),
		options: buildOptions(float64(correct), []float64{float64(modPow(a, b, 10)), float64(a * b % 100), float64((correct + 50) % 100), float64(a % 100)}, optionCount, false),
	}
}

func genKHLinearCongruenceCount(optionCount int) question {
	m := between(6, 40)
	a := between(2, m-1)
	b := between(0, m-1)
	g := gcd(a, m)
	correct := 0
	if b%g == 0 {
		correct = g
	}
	return question{
		prompt:  fmt.Sprintf("Banyak solusi x di {0, 1, ..., %d} dari kongruensi %dx ≡ %d (mod %d) adalah?", m-1, a, b, m),
		options: buildOptions(float64(correct), []float64{0, 1, float64(g), float64(m / g), float64(m)}, optionCount, false),
	}
}

// ---------- Probabilitas ----------

func genKHCovTable(optionCount int) question {
	parts := []int{1, 1, 1, 1}
	for range 6 {
		parts[rand.IntN(4)]++
	}
	p := make([]float64, 4) // P(0,0), P(0,1), P(1,0), P(1,1)
	for i, x := range parts {
		p[i] = float64(x) / 10
	}
	ex, ey := p[2]+p[3], p[1]+p[3]
	cov := cleanFloat(p[3] - ex*ey)
	return question{
		prompt:  fmt.Sprintf("Distribusi peluang bersama X dan Y (bernilai 0 atau 1): P(0,0) = %s, P(0,1) = %s, P(1,0) = %s, P(1,1) = %s. Nilai Cov(X, Y) adalah?", fmtNum(p[0]), fmtNum(p[1]), fmtNum(p[2]), fmtNum(p[3])),
		options: buildOptions(round2(cov), []float64{p[3], round2(ex * ey), -round2(cov), round2(cov + 0.1)}, optionCount, true),
	}
}

func genKHVarSum(optionCount int) question {
	vx, vy := between(1, 16), between(1, 16)
	a, b := nonZeroBetween(-3, 3), nonZeroBetween(-3, 3)
	y := formatTerms([]int{a, b}, []string{"X", "Y"})
	if rand.IntN(2) == 0 {
		correct := float64(a*a*vx + b*b*vy)
		return question{
			prompt:  fmt.Sprintf("X dan Y saling bebas dengan Var(X) = %d dan Var(Y) = %d. Nilai Var(%s) adalah?", vx, vy, y),
			options: buildOptions(correct, []float64{float64(a*vx + b*vy), float64(vx + vy), float64(a*a*vx - b*b*vy), correct + 1}, optionCount, true),
		}
	}
	cov := between(-3, 3)
	correct := float64(a*a*vx + b*b*vy + 2*a*b*cov)
	return question{
		prompt:  fmt.Sprintf("Var(X) = %d, Var(Y) = %d, dan Cov(X, Y) = %d. Nilai Var(%s) adalah?", vx, vy, cov, y),
		options: buildOptions(correct, []float64{float64(a*a*vx + b*b*vy), float64(a*a*vx + b*b*vy + a*b*cov), float64(a*vx + b*vy + 2*cov), correct + 1}, optionCount, true),
	}
}

func genKHPdfMean(optionCount int) question {
	for {
		m, a := between(0, 3), between(1, 10)
		mean := float64((m+1)*a) / float64(m+2) // f(x) = (m+1)x^m / a^(m+1) di [0, a]
		if !isExact2(mean) {
			continue
		}
		f := "konstan"
		switch m {
		case 1:
			f = "sebanding dengan x"
		case 2, 3:
			f = fmt.Sprintf("sebanding dengan x^%d", m)
		}
		return question{
			prompt:  fmt.Sprintf("Fungsi kepadatan peluang X %s pada selang [0, %d] (dan 0 di luar). Nilai E(X) adalah?", f, a),
			options: buildOptions(round2(mean), []float64{float64(a) / 2, float64(a), round2(float64(m*a) / float64(m+1)), round2(mean + 1)}, optionCount, false),
		}
	}
}

// ---------- Statistika Matematika ----------

func genKHSampleSize(optionCount int) question {
	sigma, e := between(5, 40), between(1, 5)
	v := math.Pow(1.96*float64(sigma)/float64(e), 2)
	n := int(math.Ceil(v - 1e-9))
	return question{
		prompt:  fmt.Sprintf("Ukuran sampel minimum agar galat maksimum estimasi μ tidak lebih dari %d pada tingkat kepercayaan 95%% (z = 1.96) dengan σ = %d adalah?", e, sigma),
		options: buildOptions(float64(n), []float64{float64(int(math.Floor(v))), float64(4 * sigma * sigma / (e * e)), round2(1.96 * float64(sigma) / float64(e)), float64(n + 1)}, optionCount, false),
	}
}

func genKHPooledVariance(optionCount int) question {
	for {
		n1, n2 := between(5, 30), between(5, 30)
		s1, s2 := between(4, 64), between(4, 64)
		val := float64((n1-1)*s1+(n2-1)*s2) / float64(n1+n2-2)
		if !isExact2(val) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Dua sampel: n₁ = %d dengan s₁² = %d dan n₂ = %d dengan s₂² = %d. Variansi gabungan (pooled variance) s_p² adalah?", n1, s1, n2, s2),
			options: buildOptions(round2(val), []float64{float64(s1+s2) / 2, round2(float64(n1*s1+n2*s2) / float64(n1+n2)), round2(math.Sqrt(val)), round2(val + 1)}, optionCount, false),
		}
	}
}

func genKHChiSquare(optionCount int) question {
	for {
		k := between(3, 4)
		e := pickOne([]int{10, 20, 25, 40, 50})
		obs := make([]int, k)
		exp := make([]int, k)
		sum := 0
		chi := 0.0
		for i := range k - 1 {
			d := between(-8, 8)
			obs[i] = e + d
			sum += d
		}
		obs[k-1] = e - sum
		for i := range k {
			exp[i] = e
			d := obs[i] - e
			chi += float64(d*d) / float64(e)
		}
		if chi == 0 || !isExact2(cleanFloat(chi)) {
			continue
		}
		sq := 0
		for i := range k {
			sq += (obs[i] - e) * (obs[i] - e)
		}
		return question{
			prompt:  fmt.Sprintf("Uji kecocokan: frekuensi observasi %s dengan frekuensi harapan %s. Nilai statistik chi-kuadrat adalah?", joinInts(obs), joinInts(exp)),
			options: buildOptions(round2(chi), []float64{float64(sq), round2(chi / float64(k)), round2(math.Sqrt(chi)), round2(chi + 1)}, optionCount, false),
		}
	}
}

func genKHFStat(optionCount int) question {
	for {
		s1, s2 := between(2, 60), between(2, 60)
		if s1 == s2 {
			continue
		}
		hi, lo := max(s1, s2), min(s1, s2)
		f := float64(hi) / float64(lo)
		if !isExact2(f) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Uji kesamaan dua variansi dengan s₁² = %d dan s₂² = %d. Nilai statistik F (variansi besar dibagi variansi kecil) adalah?", s1, s2),
			options: buildOptions(round2(f), []float64{round2(float64(lo) / float64(hi)), float64(hi - lo), round2(math.Sqrt(f)), round2(f + 1)}, optionCount, false),
		}
	}
}

// ---------- Statistika Terapan & Regresi ----------

func genKHMultipleReg(optionCount int) question {
	b0 := between(-10, 20)
	b1, b2 := float64(nonZeroBetween(-8, 8))/2, float64(nonZeroBetween(-8, 8))/2
	x1, x2 := between(1, 10), between(1, 10)
	correct := float64(b0) + b1*float64(x1) + b2*float64(x2)
	return question{
		prompt:  fmt.Sprintf("Model regresi berganda ŷ = %d + (%s)x₁ + (%s)x₂. Prediksi ŷ untuk x₁ = %d dan x₂ = %d adalah?", b0, fmtNum(b1), fmtNum(b2), x1, x2),
		options: buildOptions(correct, []float64{float64(b0) + b1*float64(x2) + b2*float64(x1), b1*float64(x1) + b2*float64(x2), correct + float64(b0), correct + 1}, optionCount, true),
	}
}

func genKHSSE(optionCount int) question {
	a, b := between(-5, 10), nonZeroBetween(-3, 4)
	n := between(4, 5)
	pairs := make([]string, n)
	sse, sae := 0, 0
	for i := range n {
		x := i + 1
		e := between(-3, 3)
		y := a + b*x + e
		pairs[i] = fmt.Sprintf("(%d, %d)", x, y)
		sse += e * e
		sae += absInt(e)
	}
	return question{
		prompt:  fmt.Sprintf("Model ŷ = %s dan data (x, y): %s. Jumlah kuadrat galat (SSE) adalah?", formatTerms([]int{a, b}, []string{"", "x"}), strings.Join(pairs, ", ")),
		options: buildOptions(float64(sse), []float64{float64(sae), float64(sse) / float64(n), float64(sse + 1), float64(sse * 2)}, optionCount, false),
	}
}

func genKHSlopeGeneral(optionCount int) question {
	for {
		n := between(4, 5)
		xs := rand.Perm(12)[:n]
		sx, sy, sxx, sxy := 0, 0, 0, 0
		pairs := make([]string, n)
		for i := range n {
			y := between(0, 30)
			pairs[i] = fmt.Sprintf("(%d, %d)", xs[i], y)
			sx += xs[i]
			sy += y
			sxx += xs[i] * xs[i]
			sxy += xs[i] * y
		}
		den := n*sxx - sx*sx
		b := float64(n*sxy-sx*sy) / float64(den)
		if den == 0 || !isExact2(b) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Data (x, y): %s. Kemiringan b garis regresi kuadrat terkecil y = a + bx adalah?", strings.Join(pairs, ", ")),
			options: buildOptions(round2(b), []float64{round2(float64(sxy) / float64(sxx)), -round2(b), round2(float64(sy) / float64(sx)), round2(b + 1)}, optionCount, true),
		}
	}
}

// ---------- Geometri Analitik ----------

func cross(u, v []int) []int {
	return []int{u[1]*v[2] - u[2]*v[1], u[2]*v[0] - u[0]*v[2], u[0]*v[1] - u[1]*v[0]}
}

func genKHSkewLines(optionCount int) question {
	for {
		p := []int{between(-3, 3), between(-3, 3), between(-3, 3)}
		q := []int{between(-3, 3), between(-3, 3), between(-3, 3)}
		u := []int{between(-2, 2), between(-2, 2), between(-2, 2)}
		v := []int{between(-2, 2), between(-2, 2), between(-2, 2)}
		n := cross(u, v)
		n2 := n[0]*n[0] + n[1]*n[1] + n[2]*n[2]
		if n2 == 0 {
			continue
		}
		norm := int(math.Round(math.Sqrt(float64(n2))))
		if norm*norm != n2 {
			continue
		}
		w := []int{q[0] - p[0], q[1] - p[1], q[2] - p[2]}
		dot := absInt(w[0]*n[0] + w[1]*n[1] + w[2]*n[2])
		dist := float64(dot) / float64(norm)
		if dot == 0 || !isExact2(dist) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Jarak antara garis (x, y, z) = (%s) + t(%s) dan garis (x, y, z) = (%s) + s(%s) adalah?", joinInts(p), joinInts(u), joinInts(q), joinInts(v)),
			options: buildOptions(round2(dist), []float64{float64(dot), round2(math.Sqrt(float64(w[0]*w[0] + w[1]*w[1] + w[2]*w[2]))), round2(dist + 1), float64(norm)}, optionCount, false),
		}
	}
}

func genKHTetraVolume(optionCount int) question {
	for {
		pts := make([][]int, 4)
		for i := range pts {
			pts[i] = []int{between(-3, 3), between(-3, 3), between(-3, 3)}
		}
		m := make([][]int, 3)
		for i := range m {
			m[i] = []int{pts[i+1][0] - pts[0][0], pts[i+1][1] - pts[0][1], pts[i+1][2] - pts[0][2]}
		}
		det := absInt(det3(m))
		vol := float64(det) / 6
		if det == 0 || !isExact2(vol) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Volume bidang empat dengan titik sudut A(%s), B(%s), C(%s), dan D(%s) adalah?", joinInts(pts[0]), joinInts(pts[1]), joinInts(pts[2]), joinInts(pts[3])),
			options: buildOptions(round2(vol), []float64{float64(det), round2(float64(det) / 3), round2(float64(det) / 2), round2(vol + 1)}, optionCount, false),
		}
	}
}

func genKHParallelPlanes(optionCount int) question {
	for {
		n, length := randomQuadrupleNormal()
		d1, d2 := between(-30, 30), between(-30, 30)
		k := between(2, 3)
		if d1 == d2 {
			continue
		}
		dist := float64(absInt(d1-d2)) / float64(length)
		if !isExact2(dist) {
			continue
		}
		scaled := []int{k * n[0], k * n[1], k * n[2]}
		names := []string{"x", "y", "z"}
		return question{
			prompt:  fmt.Sprintf("Jarak antara bidang sejajar %s = %d dan %s = %d adalah?", formatTerms(n, names), d1, formatTerms(scaled, names), k*d2),
			options: buildOptions(round2(dist), []float64{round2(float64(absInt(d1-k*d2)) / float64(length)), float64(absInt(d1 - d2)), round2(dist * float64(k)), round2(dist + 1)}, optionCount, false),
		}
	}
}

// ---------- Geometri Transformasi ----------

func genKHReflectRotate(optionCount int) question {
	l := pickOne(reflectionLines)
	x, y := nonZeroBetween(-9, 9), nonZeroBetween(-9, 9)
	rx, ry := reflectPoint(l, x, y)
	fx, fy := cleanFloat(-ry), cleanFloat(rx) // lalu rotasi 90° berlawanan arah jarum jam
	// urutan kebalik: rotasi dulu baru cermin
	bx, by := reflectPoint(l, -y, x)
	axis, correct, other, wrong := "x", fx, fy, bx
	if rand.IntN(2) == 0 {
		axis, correct, other, wrong = "y", fy, fx, by
	}
	return question{
		prompt:  fmt.Sprintf("Titik P(%d, %d) dicerminkan terhadap garis %s, lalu dirotasi 90° berlawanan arah jarum jam terhadap titik asal. Koordinat %s bayangan akhirnya adalah?", x, y, l.text, axis),
		options: buildOptions(round2(correct), []float64{round2(other), round2(wrong), -round2(correct), round2(correct + 1)}, optionCount, true),
	}
}

func genKHAreaImageComposite(optionCount int) question {
	for {
		m1 := [][]int{{between(-3, 3), between(-3, 3)}, {between(-3, 3), between(-3, 3)}}
		m2 := [][]int{{between(-3, 3), between(-3, 3)}, {between(-3, 3), between(-3, 3)}}
		d1 := m1[0][0]*m1[1][1] - m1[0][1]*m1[1][0]
		d2 := m2[0][0]*m2[1][1] - m2[0][1]*m2[1][0]
		if d1 == 0 || d2 == 0 {
			continue
		}
		area := between(2, 20)
		correct := float64(absInt(d1*d2) * area)
		return question{
			prompt: fmt.Sprintf(
				"Bangun datar seluas %d satuan luas ditransformasi oleh matriks %s, lalu oleh matriks %s (baris dipisah titik koma). Luas bayangan akhirnya adalah?",
				area, formatMatrix(m1), formatMatrix(m2),
			),
			options: buildOptions(correct, []float64{float64(absInt(d1+d2) * area), float64(absInt(d1) * area), float64(area), correct + float64(area)}, optionCount, false),
		}
	}
}

// ---------- Topologi ----------

func genKHConnectedSum(optionCount int) question {
	type surface struct {
		name string
		chi  int
	}
	g := between(2, 4)
	surfaces := []surface{
		{"bola (S²)", 2}, {"torus (T²)", 0}, {"bidang proyektif (RP²)", 1}, {"botol Klein", 0},
		{fmt.Sprintf("permukaan terorientasi genus %d", g), 2 - 2*g},
	}
	m, n := pickOne(surfaces), pickOne(surfaces)
	correct := float64(m.chi + n.chi - 2)
	return question{
		prompt:  fmt.Sprintf("Karakteristik Euler dari jumlahan terhubung (connected sum) %s # %s adalah?", m.name, n.name),
		options: buildOptions(correct, []float64{float64(m.chi + n.chi), float64(m.chi * n.chi), float64(m.chi + n.chi - 1), correct - 2}, optionCount, true),
	}
}

func genKHGenusTriangulation(optionCount int) question {
	// triangulasi: 3F = 2E, V - E + F = 2 - 2g  ->  F = 2(V - 2 + 2g), E = 3(V - 2 + 2g)
	g, v := between(0, 4), between(7, 20)
	f := 2 * (v - 2 + 2*g)
	e := 3 * (v - 2 + 2*g)
	chi := v - e + f
	return question{
		prompt:  fmt.Sprintf("Suatu triangulasi permukaan tertutup terorientasi memiliki %d titik, %d sisi, dan %d segitiga. Genus permukaan tersebut adalah?", v, e, f),
		options: buildOptions(float64(g), []float64{float64(chi), float64(2 * g), float64(g + 1), float64(absInt(chi))}, optionCount, true),
	}
}

// ---------- Metode Matematika ----------

func genKHFourierX2(optionCount int) question {
	for {
		c, n := nonZeroBetween(-3, 3), between(1, 10)
		sign := 1
		if n%2 == 1 {
			sign = -1
		}
		val := float64(4*c*sign) / float64(n*n) // f(x) = cx² -> a_n = 4c(-1)^n/n²
		if !isExact2(val) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Deret Fourier f(x) = %s pada (-π, π) memiliki koefisien kosinus a_n. Nilai a_%d = ?", formatTerms([]int{c}, []string{"x²"}), n),
			options: buildOptions(round2(val), []float64{-round2(val), round2(float64(4*c) / float64(n)), round2(float64(2*c*sign) / float64(n*n)), 0}, optionCount, true),
		}
	}
}

func genKHConvolution(optionCount int) question {
	for {
		m, n := between(0, 2), between(0, 2)
		t := between(1, 3)
		coef := float64(factorial(m)*factorial(n)) / float64(factorial(m+n+1))
		val := coef * float64(intPow(t, m+n+1)) // (t^m * t^n)(t) = m!n!/(m+n+1)! · t^(m+n+1)
		if !isExact2(cleanFloat(val)) {
			continue
		}
		term := func(k int) string { return []string{"1", "t", "t²"}[k] }
		return question{
			prompt:  fmt.Sprintf("Konvolusi (f * g)(t) = ∫ dari 0 sampai t f(τ)g(t - τ) dτ dengan f(t) = %s dan g(t) = %s. Nilai (f * g)(%d) adalah?", term(m), term(n), t),
			options: buildOptions(round2(val), []float64{float64(intPow(t, m+n)), float64(intPow(t, m+n+1)), round2(val * 2), round2(val + 1)}, optionCount, false),
		}
	}
}

// ---------- Riset Operasi ----------

func genKHTransportation2x3(optionCount int) question {
	for {
		s1, s2 := between(10, 50), between(10, 50)
		d1 := between(5, s1+s2-10)
		d2 := between(3, s1+s2-d1-2)
		d3 := s1 + s2 - d1 - d2
		c := [][]int{{between(1, 15), between(1, 15), between(1, 15)}, {between(1, 15), between(1, 15), between(1, 15)}}
		best, worst := 1<<30, -1
		for x11 := 0; x11 <= min(s1, d1); x11++ {
			for x12 := 0; x12 <= min(s1-x11, d2); x12++ {
				x13 := s1 - x11 - x12
				if x13 > d3 {
					continue
				}
				cost := c[0][0]*x11 + c[0][1]*x12 + c[0][2]*x13 + c[1][0]*(d1-x11) + c[1][1]*(d2-x12) + c[1][2]*(d3-x13)
				best, worst = min(best, cost), max(worst, cost)
			}
		}
		if worst < 0 || best == worst {
			continue
		}
		return question{
			prompt: fmt.Sprintf(
				"Masalah transportasi: pasokan P1 = %d, P2 = %d; permintaan G1 = %d, G2 = %d, G3 = %d. Biaya per unit P1→(G1, G2, G3) = (%s) dan P2→(G1, G2, G3) = (%s). Biaya total minimum adalah?",
				s1, s2, d1, d2, d3, joinInts(c[0]), joinInts(c[1]),
			),
			options: buildOptions(float64(best), []float64{float64(worst), float64(best + 1), float64(best + c[0][0]), float64((best + worst) / 2)}, optionCount, false),
		}
	}
}

func genKHAssignment4(optionCount int) question {
	c := make([][]int, 4)
	rows := make([]string, 4)
	for i := range c {
		c[i] = []int{between(1, 20), between(1, 20), between(1, 20), between(1, 20)}
		rows[i] = fmt.Sprintf("%c = [%s]", 'A'+i, joinInts(c[i]))
	}
	best, worst, diag := assignmentCosts(c)
	return question{
		prompt:  fmt.Sprintf("Empat pekerja ditugaskan ke empat pekerjaan (1-4), satu pekerja satu pekerjaan, dengan biaya: %s. Biaya penugasan total minimum adalah?", strings.Join(rows, ", ")),
		options: buildOptions(float64(best), []float64{float64(worst), float64(diag), float64(best + 1), float64(best + 2)}, optionCount, false),
	}
}

func genKHCPM(optionCount int) question {
	n := between(5, 7)
	dur := make([]int, n)
	preds := make([][]int, n)
	ef := make([]int, n)
	parts := make([]string, n)
	total, longest := 0, 0
	for i := range n {
		dur[i] = between(1, 9)
		total += dur[i]
		if i > 0 {
			for _, p := range rand.Perm(i)[:between(1, min(2, i))] {
				preds[i] = append(preds[i], p)
			}
			sort.Ints(preds[i])
		}
		start := 0
		names := make([]string, len(preds[i]))
		for k, p := range preds[i] {
			start = max(start, ef[p])
			names[k] = string(rune('A' + p))
		}
		ef[i] = start + dur[i]
		longest = max(longest, ef[i])
		pre := "–"
		if len(names) > 0 {
			pre = strings.Join(names, ",")
		}
		parts[i] = fmt.Sprintf("%c(%d, %s)", 'A'+i, dur[i], pre)
	}
	maxDur := 0
	for _, d := range dur {
		maxDur = max(maxDur, d)
	}
	return question{
		prompt:  fmt.Sprintf("Proyek dengan aktivitas (durasi hari, pendahulu): %s. Durasi minimum penyelesaian proyek (lintasan kritis) adalah? (hari)", strings.Join(parts, "; ")),
		options: buildOptions(float64(longest), []float64{float64(total), float64(longest - 1), float64(longest + 1), float64(maxDur)}, optionCount, false),
	}
}

// ---------- Matematika Keuangan ----------

func genKHIRR(optionCount int) question {
	r := pickOne([]int{10, 20, 25})
	m := between(1, 20)
	var i0, cf1, cf2 int
	switch r { // arus kas sama dua tahun: I = A(1/(1+r) + 1/(1+r)²)
	case 10:
		i0, cf1 = 210*m, 121*m
	case 20:
		i0, cf1 = 220*m, 144*m
	default:
		i0, cf1 = 36*m, 25*m
	}
	cf2 = cf1
	if rand.IntN(2) == 0 { // pola obligasi: bunga saja lalu pokok + bunga
		i0 = 100 * m
		cf1 = r * m
		cf2 = (100 + r) * m
	}
	return question{
		prompt:  fmt.Sprintf("Investasi Rp%d ribu menghasilkan Rp%d ribu di akhir tahun 1 dan Rp%d ribu di akhir tahun 2. IRR investasi tersebut adalah? (%%)", i0, cf1, cf2),
		options: buildOptions(float64(r), []float64{5, 10, 15, 20, 25, 30}, optionCount, false),
	}
}

func genKHBondPrice(optionCount int) question {
	for {
		face := between(1, 20) * 1000
		c := pickOne([]int{5, 8, 10, 12, 15})
		y := pickOne([]int{10, 20})
		coupon := float64(face*c) / 100
		yf := 1 + float64(y)/100
		price := cleanFloat(coupon/yf + (coupon+float64(face))/(yf*yf))
		if !isExact2(price) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Obligasi bernilai nominal Rp%d ribu, kupon %d%% per tahun, jatuh tempo 2 tahun. Dengan yield %d%% per tahun, harga wajar obligasi adalah? (ribu rupiah)", face, c, y),
			options: buildOptions(round2(price), []float64{float64(face), round2(2*coupon + float64(face)), round2(float64(face) / (yf * yf)), round2(price + 100)}, optionCount, false),
		}
	}
}

func genKHPVAnnuity(optionCount int) question {
	i := pickOne([]int{10, 20})
	n := between(2, 3)
	m := between(1, 10)
	// A kelipatan (1+i)^n supaya PV bulat: PV = A·((1+i)^n - 1)/(i·(1+i)^n)
	base := intPow(100+i, n) / intPow(10, n) // 121/1331 atau 144/1728
	a := m * base
	pv := a * (intPow(100+i, n) - intPow(100, n)) / (i * intPow(100+i, n) / 100)
	return question{
		prompt:  fmt.Sprintf("Nilai sekarang anuitas Rp%d ribu yang dibayar setiap akhir tahun selama %d tahun dengan bunga %d%% per tahun adalah? (ribu rupiah)", a, n, i),
		options: buildOptions(float64(pv), []float64{float64(a * n), float64(a * n * 100 / (100 + i)), float64(pv + a), float64(pv - 10)}, optionCount, false),
	}
}

// ---------- Aljabar Linear Terapan ----------

func genKHPCA2(optionCount int) question {
	for {
		t := pickOne(pyTriples)
		k := between(1, 3)
		c := between(1, 20)
		a := c + t.a*k*randSign()
		if t.b*k%2 != 0 {
			continue
		}
		b := t.b * k / 2 * randSign()
		if (a+c+t.c*k)%2 != 0 {
			continue
		}
		l1, l2 := (a+c+t.c*k)/2, (a+c-t.c*k)/2
		if l2 <= 0 || a <= 0 {
			continue
		}
		pct := float64(l1) / float64(l1+l2) * 100
		if !isExact2(cleanFloat(pct)) {
			continue
		}
		cov := formatMatrix([][]int{{a, b}, {b, c}})
		if rand.IntN(2) == 0 {
			return question{prompt: fmt.Sprintf("Matriks kovariansi Σ = %s (baris dipisah titik koma). Variansi pada komponen utama pertama (nilai eigen terbesar) adalah?", cov), options: buildOptions(float64(l1), []float64{float64(l2), float64(max(a, c)), float64(a + c), float64(l1 + 1)}, optionCount, false)}
		}
		return question{prompt: fmt.Sprintf("Matriks kovariansi Σ = %s (baris dipisah titik koma). Persentase variansi yang dijelaskan komponen utama pertama adalah? (%%)", cov), options: buildOptions(round2(pct), []float64{round2(float64(max(a, c)) / float64(a+c) * 100), round2(100 - pct), round2(pct + 5), float64(l1)}, optionCount, false)}
	}
}

func genKHGradientDescent(optionCount int) question {
	for {
		a := between(1, 3)
		c := between(-5, 5)
		w0 := between(-5, 8)
		eta := pickOne([]float64{0.05, 0.1, 0.2, 0.25})
		w1 := cleanFloat(float64(w0) - eta*2*float64(a)*(float64(w0)-float64(c)))
		w2 := cleanFloat(w1 - eta*2*float64(a)*(w1-float64(c)))
		if !isExact2(w2) || w0 == c {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Gradient descent pada f(w) = %d(%s)² dengan w₀ = %d dan learning rate η = %s. Nilai w₂ adalah?", a, formatPolyVar("w", 1, -c), w0, fmtNum(eta)),
			options: buildOptions(round2(w2), []float64{round2(w1), float64(c), round2(float64(w0) + eta*2*float64(a)*(float64(w0)-float64(c))), round2(w2 + 1)}, optionCount, true),
		}
	}
}

func genKHNormalEquation(optionCount int) question {
	// x = 0..3: Sxx = 5, x̄ = 1,5
	ys := make([]int, 4)
	sy, sxy := 0, 0
	pairs := make([]string, 4)
	for i := range ys {
		ys[i] = between(-5, 20)
		sy += ys[i]
		sxy += i * ys[i]
		pairs[i] = fmt.Sprintf("(%d, %d)", i, ys[i])
	}
	ybar := float64(sy) / 4
	b1 := cleanFloat((float64(sxy) - 1.5*float64(sy)) / 5)
	b0 := cleanFloat(ybar - 1.5*b1)
	if rand.IntN(2) == 0 {
		return question{prompt: fmt.Sprintf("Dengan persamaan normal (XᵀX)β = Xᵀy untuk model y = β₀ + β₁x dan data %s, nilai β₁ adalah?", strings.Join(pairs, ", ")), options: buildOptions(round2(b1), []float64{round2(b0), round2(float64(sxy) / 14), round2(ybar), round2(b1 + 1)}, optionCount, true)}
	}
	return question{prompt: fmt.Sprintf("Dengan persamaan normal (XᵀX)β = Xᵀy untuk model y = β₀ + β₁x dan data %s, nilai β₀ adalah?", strings.Join(pairs, ", ")), options: buildOptions(round2(b0), []float64{round2(b1), round2(ybar), round2(ybar + 1.5*b1), round2(b0 + 1)}, optionCount, true)}
}
