package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"sort"
	"strings"
)

// ---------- Kampus: materi batch 2 ----------

func kampusMateriBatch2() []materiSpec {
	return []materiSpec{
		{"Matematika Diskrit", mixOf(genKInclusionExclusion, genKRelationsCount, genKFunctionsCount, genKStarsBars, genKPigeonhole, genKSetsUnion3)},
		{"Teori Graf", mixOf(genKHandshake, genKGraphEdges, genKTreeEuler, genKChromatic, genKCayley, genKShortestPath5)},
		{"Teori Bilangan", mixOf(genKGcdLcm, genKModPow, genKModInverseGeneral, genKCRT2, genKDivisorFunctions, genKLastDigit)},
		{"Probabilitas", mixOf(genKBinomialMeanVar, genKDiscreteRV, genKLinearTransformRV, genKPdfConstant, genKUniformContinuous, genKBayes, genKPoissonGeometric)},
		{"Statistika Matematika", mixOf(genKSampleVariance, genKStandardError, genKZStatistic, genKConfidenceInterval, genKMLE, genKTStatistic)},
		{"Statistika Terapan & Regresi", mixOf(genKRegressionSlope, genKResidual, genKRSquared, genKCovariance, genKCorrelation)},
		{"Geometri Analitik", mixOf(genKPointPlaneDistance, genKSphere, genKParallelepiped, genKParallelogramArea, genKAngleBetweenPlanes, genKLinePlaneIntersection)},
		{"Geometri Transformasi", mixOf(genKRotationComposition, genKReflectLineRational, genKAreaUnderMatrix, genKDilationCenter)},
		{"Topologi Dasar", mixOf(genKEulerCharacteristic, genKIntervalSet, genKFiniteTopology, genKBettiGenus)},
		{"Metode Matematika Teknik & Fisika", mixOf(genKFourierCoeff, genKGammaBeta, genKDiracDelta, genKJacobian, genKTaylorCoeff, genKLaplaceInverse)},
		{"Riset Operasi", mixOf(genKLPBox, genKTransportation2x2, genKAssignment3, genKQueueMM1, genKEOQ)},
		{"Matematika Keuangan Lanjut", mixOf(genKNPV, genKFVAnnuity, genKPerpetuity, genKEffectiveRate, genKPayback)},
		{"Aljabar Linear Terapan (Data Science)", mixOf(genKCosineSimilarity, genKVectorNorms, genKMatrixChain, genKPCAVariance, genKLeastSquaresOrigin, genKMSE, genKFrobenius)},
	}
}

// ---------- helper graf ----------

type wEdge struct{ u, v, w int }

const nodeNames = "ABCDEFGH"

// randomWeightedGraph: graf terhubung n simpul (pohon acak + sisi tambahan), bobot 1-9.
func randomWeightedGraph(n, extra int) []wEdge {
	perm := rand.Perm(n)
	used := map[[2]int]bool{}
	var edges []wEdge
	add := func(u, v int) {
		if u > v {
			u, v = v, u
		}
		if u == v || used[[2]int{u, v}] {
			return
		}
		used[[2]int{u, v}] = true
		edges = append(edges, wEdge{u, v, between(1, 9)})
	}
	for i := 1; i < n; i++ {
		add(perm[i], perm[rand.IntN(i)])
	}
	for range extra {
		add(rand.IntN(n), rand.IntN(n))
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].u != edges[j].u {
			return edges[i].u < edges[j].u
		}
		return edges[i].v < edges[j].v
	})
	return edges
}

func formatEdges(edges []wEdge) string {
	parts := make([]string, len(edges))
	for i, e := range edges {
		parts[i] = fmt.Sprintf("%c–%c: %d", nodeNames[e.u], nodeNames[e.v], e.w)
	}
	return strings.Join(parts, ", ")
}

func shortestDistances(n int, edges []wEdge) [][]int {
	const inf = 1 << 30
	d := make([][]int, n)
	for i := range d {
		d[i] = make([]int, n)
		for j := range d[i] {
			if i != j {
				d[i][j] = inf
			}
		}
	}
	for _, e := range edges {
		d[e.u][e.v] = min(d[e.u][e.v], e.w)
		d[e.v][e.u] = min(d[e.v][e.u], e.w)
	}
	for k := range n {
		for i := range n {
			for j := range n {
				if d[i][k]+d[k][j] < d[i][j] {
					d[i][j] = d[i][k] + d[k][j]
				}
			}
		}
	}
	return d
}

func mstWeight(n int, edges []wEdge) int {
	sorted := append([]wEdge{}, edges...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].w < sorted[j].w })
	parent := make([]int, n)
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(x int) int {
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}
	total := 0
	for _, e := range sorted {
		if a, b := find(e.u), find(e.v); a != b {
			parent[a] = b
			total += e.w
		}
	}
	return total
}

// ---------- Matematika Diskrit ----------

func genKInclusionExclusion(optionCount int) question {
	n := between(50, 500)
	a := between(2, 9)
	b := between(2, 9)
	for b == a {
		b = between(2, 9)
	}
	either := n/a + n/b - n/lcm(a, b)
	if rand.IntN(2) == 0 {
		return question{
			prompt:  fmt.Sprintf("Banyak bilangan bulat dari 1 sampai %d yang habis dibagi %d atau %d adalah?", n, a, b),
			options: buildOptions(float64(either), []float64{float64(n/a + n/b), float64(n - either), float64(n / lcm(a, b)), float64(either + 1)}, optionCount, false),
		}
	}
	return question{
		prompt:  fmt.Sprintf("Banyak bilangan bulat dari 1 sampai %d yang tidak habis dibagi %d maupun %d adalah?", n, a, b),
		options: buildOptions(float64(n-either), []float64{float64(either), float64(n - n/a - n/b), float64(n - n/lcm(a, b)), float64(n - either + 1)}, optionCount, false),
	}
}

func genKRelationsCount(optionCount int) question {
	n := between(2, 4)
	all, refl, sym := intPow(2, n*n), intPow(2, n*n-n), intPow(2, n*(n+1)/2)
	cands := []float64{float64(all), float64(refl), float64(sym), float64(intPow(2, n)), float64(n * n)}
	switch rand.IntN(3) {
	case 0:
		return question{prompt: fmt.Sprintf("Banyak relasi pada himpunan dengan %d anggota adalah?", n), options: buildOptions(float64(all), cands, optionCount, false)}
	case 1:
		return question{prompt: fmt.Sprintf("Banyak relasi refleksif pada himpunan dengan %d anggota adalah?", n), options: buildOptions(float64(refl), cands, optionCount, false)}
	default:
		return question{prompt: fmt.Sprintf("Banyak relasi simetris pada himpunan dengan %d anggota adalah?", n), options: buildOptions(float64(sym), cands, optionCount, false)}
	}
}

func genKFunctionsCount(optionCount int) question {
	m, n := between(2, 5), between(2, 6)
	switch rand.IntN(3) {
	case 0:
		correct := float64(intPow(n, m))
		return question{
			prompt:  fmt.Sprintf("Banyak fungsi dari himpunan A (%d anggota) ke himpunan B (%d anggota) adalah?", m, n),
			options: buildOptions(correct, []float64{float64(intPow(m, n)), float64(m * n), float64(combination(n+m-1, m)), correct + 1}, optionCount, false),
		}
	case 1:
		if n < m {
			m, n = n, m
		}
		correct := float64(factorial(n) / factorial(n-m))
		return question{
			prompt:  fmt.Sprintf("Banyak fungsi satu-satu (injektif) dari himpunan A (%d anggota) ke himpunan B (%d anggota) adalah?", m, n),
			options: buildOptions(correct, []float64{float64(intPow(n, m)), float64(combination(n, m)), float64(m * n), correct + 1}, optionCount, false),
		}
	default:
		m = between(3, 7)
		correct := float64(intPow(2, m) - 2)
		return question{
			prompt:  fmt.Sprintf("Banyak fungsi pada (surjektif) dari himpunan dengan %d anggota ke himpunan dengan 2 anggota adalah?", m),
			options: buildOptions(correct, []float64{float64(intPow(2, m)), float64(intPow(m, 2)), float64(2 * m), correct + 2}, optionCount, false),
		}
	}
}

func genKStarsBars(optionCount int) question {
	k, n := between(3, 5), between(5, 15)
	vars := make([]string, k)
	for i := range vars {
		vars[i] = fmt.Sprintf("x%d", i+1)
	}
	eq := strings.Join(vars, " + ")
	nonneg, pos := combination(n+k-1, k-1), combination(n-1, k-1)
	cands := []float64{float64(nonneg), float64(pos), float64(combination(n+k, k)), float64(combination(n, k))}
	if rand.IntN(2) == 0 {
		return question{prompt: fmt.Sprintf("Banyak solusi bilangan bulat tak negatif dari %s = %d adalah?", eq, n), options: buildOptions(float64(nonneg), cands, optionCount, false)}
	}
	return question{prompt: fmt.Sprintf("Banyak solusi bilangan bulat positif dari %s = %d adalah?", eq, n), options: buildOptions(float64(pos), cands, optionCount, false)}
}

func genKPigeonhole(optionCount int) question {
	c, k := between(3, 8), between(2, 6)
	correct := float64(c*(k-1) + 1)
	return question{
		prompt:  fmt.Sprintf("Sebuah kotak berisi banyak kaus kaki dengan %d warna. Minimal berapa kaus kaki harus diambil agar pasti ada %d kaus kaki berwarna sama?", c, k),
		options: buildOptions(correct, []float64{float64(c * k), float64(c * (k - 1)), float64(c + k), float64(c*k + 1)}, optionCount, false),
	}
}

func genKSetsUnion3(optionCount int) question {
	// region: hanya A, hanya B, hanya C, A∩B saja, A∩C saja, B∩C saja, A∩B∩C
	r := make([]int, 7)
	for i := range r {
		r[i] = between(0, 12)
	}
	r[6] = between(1, 8)
	A, B, C := r[0]+r[3]+r[4]+r[6], r[1]+r[3]+r[5]+r[6], r[2]+r[4]+r[5]+r[6]
	AB, AC, BC := r[3]+r[6], r[4]+r[6], r[5]+r[6]
	union := 0
	for _, v := range r {
		union += v
	}
	return question{
		prompt: fmt.Sprintf(
			"|A| = %d, |B| = %d, |C| = %d, |A ∩ B| = %d, |A ∩ C| = %d, |B ∩ C| = %d, dan |A ∩ B ∩ C| = %d. Nilai |A ∪ B ∪ C| adalah?",
			A, B, C, AB, AC, BC, r[6],
		),
		options: buildOptions(float64(union), []float64{float64(A + B + C), float64(union - r[6]), float64(union + r[6]), float64(A + B + C - AB - AC - BC)}, optionCount, false),
	}
}

// ---------- Teori Graf ----------

func genKHandshake(optionCount int) question {
	n := between(5, 8)
	edges := randomWeightedGraph(n, between(1, 6))
	deg := make([]int, n)
	for _, e := range edges {
		deg[e.u]++
		deg[e.v]++
	}
	sort.Sort(sort.Reverse(sort.IntSlice(deg)))
	sum := 2 * len(edges)
	return question{
		prompt:  fmt.Sprintf("Sebuah graf sederhana memiliki barisan derajat %s. Banyak sisi graf tersebut adalah?", joinInts(deg)),
		options: buildOptions(float64(len(edges)), []float64{float64(sum), float64(len(edges) + 1), float64(n), float64(len(edges) - 1)}, optionCount, false),
	}
}

func genKGraphEdges(optionCount int) question {
	switch rand.IntN(4) {
	case 0:
		n := between(4, 12)
		correct := float64(n * (n - 1) / 2)
		return question{prompt: fmt.Sprintf("Banyak sisi graf lengkap K_%d adalah?", n), options: buildOptions(correct, []float64{float64(n * (n - 1)), float64(n * n), float64(n), correct + 1}, optionCount, false)}
	case 1:
		m, n := between(2, 8), between(2, 8)
		correct := float64(m * n)
		return question{prompt: fmt.Sprintf("Banyak sisi graf bipartit lengkap K_%d,%d adalah?", m, n), options: buildOptions(correct, []float64{float64(m + n), float64((m + n) * (m + n - 1) / 2), correct + 1, correct - 1}, optionCount, false)}
	case 2:
		n := between(4, 12)
		correct := float64(2 * n)
		return question{prompt: fmt.Sprintf("Banyak sisi graf roda W_%d (lingkaran %d simpul ditambah 1 simpul pusat) adalah?", n, n), options: buildOptions(correct, []float64{float64(n), float64(n + 1), float64(2*n + 1), float64(n * (n + 1) / 2)}, optionCount, false)}
	default:
		k := between(2, 5)
		correct := float64(k * intPow(2, k-1))
		return question{prompt: fmt.Sprintf("Banyak sisi graf hypercube Q_%d adalah?", k), options: buildOptions(correct, []float64{float64(intPow(2, k)), float64(k * intPow(2, k)), float64(intPow(2, k) - 1), correct + 1}, optionCount, false)}
	}
}

func genKTreeEuler(optionCount int) question {
	switch rand.IntN(3) {
	case 0:
		n := between(5, 50)
		return question{prompt: fmt.Sprintf("Sebuah pohon memiliki %d simpul. Banyak sisinya adalah?", n), options: buildOptions(float64(n-1), []float64{float64(n), float64(n + 1), float64(n * (n - 1) / 2)}, optionCount, false)}
	case 1:
		n, k := between(10, 50), between(2, 6)
		return question{prompt: fmt.Sprintf("Sebuah hutan (forest) memiliki %d simpul dan %d komponen. Banyak sisinya adalah?", n, k), options: buildOptions(float64(n-k), []float64{float64(n - 1), float64(n + k), float64(n - k + 1), float64(n)}, optionCount, false)}
	default:
		v := between(5, 20)
		e := between(v, 3*v-6)
		correct := float64(e - v + 2)
		return question{prompt: fmt.Sprintf("Graf planar terhubung memiliki %d simpul dan %d sisi. Banyak muka (face) termasuk muka luar adalah?", v, e), options: buildOptions(correct, []float64{float64(e - v), float64(v - e + 2), float64(e + v - 2), correct + 1}, optionCount, true)}
	}
}

func genKChromatic(optionCount int) question {
	type item struct {
		desc  string
		value int
	}
	n := between(4, 11)
	cycle := 2
	if n%2 == 1 {
		cycle = 3
	}
	wheel := 3
	if n%2 == 1 {
		wheel = 4
	}
	it := pickOne([]item{
		{fmt.Sprintf("graf lingkaran C_%d", n), cycle},
		{fmt.Sprintf("graf lengkap K_%d", n), n},
		{fmt.Sprintf("graf bipartit lengkap K_%d,%d", between(2, 6), between(2, 6)), 2},
		{fmt.Sprintf("graf roda W_%d (lingkaran %d simpul + 1 pusat)", n, n), wheel},
		{fmt.Sprintf("sebuah pohon dengan %d simpul", n), 2},
		{"graf Petersen", 3},
	})
	return question{
		prompt:  fmt.Sprintf("Bilangan kromatik %s adalah?", it.desc),
		options: buildOptions(float64(it.value), []float64{2, 3, 4, float64(n), float64(n - 1)}, optionCount, false),
	}
}

func genKCayley(optionCount int) question {
	n := between(3, 7)
	correct := float64(intPow(n, n-2))
	return question{
		prompt:  fmt.Sprintf("Banyak pohon merentang (spanning tree) dari graf lengkap K_%d adalah?", n),
		options: buildOptions(correct, []float64{float64(intPow(n, n-1)), float64(factorial(n)), float64(n * (n - 1) / 2), float64(intPow(n-1, n-2))}, optionCount, false),
	}
}

func genKShortestPath5(optionCount int) question {
	n := 5
	edges := randomWeightedGraph(n, between(2, 4))
	d := shortestDistances(n, edges)
	correct := float64(d[0][n-1])
	total := 0
	for _, e := range edges {
		total += e.w
	}
	return question{
		prompt:  fmt.Sprintf("Graf berbobot tak berarah dengan sisi (bobot): %s. Panjang lintasan terpendek dari A ke E adalah?", formatEdges(edges)),
		options: buildOptions(correct, []float64{correct + 1, correct + 2, correct - 1, float64(mstWeight(n, edges)), float64(total)}, optionCount, false),
	}
}

// ---------- Teori Bilangan ----------

func genKGcdLcm(optionCount int) question {
	a, b := between(12, 300), between(12, 300)
	g, l := gcd(a, b), lcm(a, b)
	if rand.IntN(2) == 0 {
		return question{prompt: fmt.Sprintf("FPB dari %d dan %d adalah?", a, b), options: buildOptions(float64(g), []float64{float64(l), float64(2 * g), float64(g + 1), float64(absInt(a - b))}, optionCount, false)}
	}
	return question{prompt: fmt.Sprintf("KPK dari %d dan %d adalah?", a, b), options: buildOptions(float64(l), []float64{float64(a * b), float64(g), float64(2 * l), float64(l + a)}, optionCount, false)}
}

func genKModPow(optionCount int) question {
	a, b, m := between(2, 20), between(10, 200), between(5, 50)
	correct := modPow(a, b, m)
	return question{
		prompt:  fmt.Sprintf("Sisa pembagian %d^%d oleh %d adalah?", a, b, m),
		options: buildOptions(float64(correct), []float64{float64(a * b % m), float64(a % m), float64((correct + 1) % m), float64(m - 1 - correct)}, optionCount, false),
	}
}

func genKModInverseGeneral(optionCount int) question {
	m := between(7, 40)
	a := between(2, m-1)
	for gcd(a, m) != 1 {
		a = between(2, m-1)
	}
	inv := modInverse(a, m)
	return question{
		prompt:  fmt.Sprintf("Invers dari %d modulo %d (di antara 0 sampai %d) adalah?", a, m, m-1),
		options: buildOptions(float64(inv), []float64{float64(m - inv), float64(a), float64(m - a), float64((inv + 1) % m)}, optionCount, false),
	}
}

func genKCRT2(optionCount int) question {
	moduli := []int{3, 4, 5, 7, 8, 9, 11, 13}
	m := pickOne(moduli)
	n := pickOne(moduli)
	for n == m || gcd(m, n) != 1 {
		n = pickOne(moduli)
	}
	a, b := between(0, m-1), between(0, n-1)
	x := 0
	for x = 1; x <= m*n; x++ {
		if x%m == a && x%n == b {
			break
		}
	}
	return question{
		prompt:  fmt.Sprintf("Bilangan bulat positif terkecil x yang memenuhi x ≡ %d (mod %d) dan x ≡ %d (mod %d) adalah?", a, m, b, n),
		options: buildOptions(float64(x), []float64{float64(x + m*n), float64(a + b), float64(m * n), float64(x + 1)}, optionCount, false),
	}
}

func genKDivisorFunctions(optionCount int) question {
	n := between(12, 400)
	d, s, phi := divisorCount(n), divisorSum(n), eulerPhi(n)
	cands := []float64{float64(d), float64(s), float64(phi), float64(n - 1), float64(s - n)}
	switch rand.IntN(3) {
	case 0:
		return question{prompt: fmt.Sprintf("Banyak pembagi positif dari %d adalah?", n), options: buildOptions(float64(d), cands, optionCount, false)}
	case 1:
		return question{prompt: fmt.Sprintf("Jumlah semua pembagi positif dari %d adalah?", n), options: buildOptions(float64(s), cands, optionCount, false)}
	default:
		return question{prompt: fmt.Sprintf("Nilai fungsi phi Euler φ(%d) adalah?", n), options: buildOptions(float64(phi), cands, optionCount, false)}
	}
}

func genKLastDigit(optionCount int) question {
	a, b := between(2, 99), between(10, 2000)
	correct := modPow(a, b, 10)
	return question{
		prompt:  fmt.Sprintf("Angka satuan dari %d^%d adalah?", a, b),
		options: buildOptions(float64(correct), []float64{float64(a % 10), float64(b % 10), float64((correct + 5) % 10), float64(a * b % 10)}, optionCount, false),
	}
}

// ---------- Probabilitas ----------

func genKBinomialMeanVar(optionCount int) question {
	for {
		n := between(5, 100)
		p := pickOne([]float64{0.1, 0.2, 0.25, 0.4, 0.5})
		mean := cleanFloat(float64(n) * p)
		variance := cleanFloat(float64(n) * p * (1 - p))
		if !isExact2(mean) || !isExact2(variance) {
			continue
		}
		if rand.IntN(2) == 0 {
			return question{prompt: fmt.Sprintf("X ~ Binomial(n = %d, p = %s). Nilai E(X) adalah?", n, fmtNum(p)), options: buildOptions(mean, []float64{variance, round2(mean * p), mean + 1, float64(n)}, optionCount, false)}
		}
		return question{prompt: fmt.Sprintf("X ~ Binomial(n = %d, p = %s). Nilai Var(X) adalah?", n, fmtNum(p)), options: buildOptions(variance, []float64{mean, round2(mean * mean), round2(math.Sqrt(variance)), variance + 1}, optionCount, false)}
	}
}

func genKDiscreteRV(optionCount int) question {
	k := between(3, 4)
	values := rand.Perm(9)[:k]
	for i := range values {
		values[i] -= 2 // -2..6
	}
	sort.Ints(values)
	// peluang kelipatan 0,1 (jumlah 10 bagian)
	parts := make([]int, k)
	remaining := 10
	for i := range k - 1 {
		parts[i] = between(1, remaining-(k-1-i))
		remaining -= parts[i]
	}
	parts[k-1] = remaining
	ex10, ex2 := 0, 0 // E·10, E(X²)·10
	probs := make([]string, k)
	for i := range k {
		ex10 += values[i] * parts[i]
		ex2 += values[i] * values[i] * parts[i]
		probs[i] = fmtNum(float64(parts[i]) / 10)
	}
	mean := float64(ex10) / 10
	variance := float64(10*ex2-ex10*ex10) / 100
	if rand.IntN(2) == 0 {
		return question{
			prompt:  fmt.Sprintf("Variabel acak X bernilai %s dengan peluang berturut-turut %s. Nilai E(X) adalah?", joinInts(values), strings.Join(probs, ", ")),
			options: buildOptions(round2(mean), []float64{round2(variance), round2(float64(ex2) / 10), round2(mean + 1), round2(-mean)}, optionCount, true),
		}
	}
	return question{
		prompt:  fmt.Sprintf("Variabel acak X bernilai %s dengan peluang berturut-turut %s. Nilai Var(X) adalah?", joinInts(values), strings.Join(probs, ", ")),
		options: buildOptions(round2(variance), []float64{round2(mean), round2(float64(ex2) / 10), round2(variance + 1), round2(math.Abs(mean))}, optionCount, true),
	}
}

func genKLinearTransformRV(optionCount int) question {
	mu, sigma2 := between(-5, 20), between(1, 25)
	a, b := nonZeroBetween(-4, 5), between(-10, 10)
	y := formatTerms([]int{a, b}, []string{"X", ""})
	if rand.IntN(2) == 0 {
		correct := float64(a*mu + b)
		return question{
			prompt:  fmt.Sprintf("E(X) = %d dan Var(X) = %d. Nilai E(Y) untuk Y = %s adalah?", mu, sigma2, y),
			options: buildOptions(correct, []float64{float64(a * mu), float64(a*sigma2 + b), float64(mu + b), correct + 1}, optionCount, true),
		}
	}
	correct := float64(a * a * sigma2)
	return question{
		prompt:  fmt.Sprintf("E(X) = %d dan Var(X) = %d. Nilai Var(Y) untuk Y = %s adalah?", mu, sigma2, y),
		options: buildOptions(correct, []float64{float64(a * sigma2), float64(a*a*sigma2 + b), float64(sigma2), correct + float64(b*b)}, optionCount, true),
	}
}

func genKPdfConstant(optionCount int) question {
	for {
		m := between(0, 3)
		a := pickOne([]int{1, 2, 4, 5, 10})
		k := float64(m+1) / float64(intPow(a, m+1))
		if !isExact2(k) {
			continue
		}
		f := "k"
		switch m {
		case 1:
			f = "kx"
		case 2, 3:
			f = fmt.Sprintf("kx^%d", m)
		}
		return question{
			prompt:  fmt.Sprintf("f(x) = %s untuk 0 ≤ x ≤ %d (dan 0 di luar selang itu) adalah fungsi kepadatan peluang. Nilai k = ?", f, a),
			options: buildOptions(round2(k), []float64{round2(1 / float64(intPow(a, m+1))), round2(float64(m+1) / float64(a)), float64(intPow(a, m+1)), round2(k * 2)}, optionCount, false),
		}
	}
}

func genKUniformContinuous(optionCount int) question {
	for {
		a := between(0, 10)
		w := pickOne([]int{2, 4, 5, 6, 10, 12, 20})
		b := a + w
		switch rand.IntN(3) {
		case 0:
			mean := float64(a+b) / 2
			return question{prompt: fmt.Sprintf("X ~ Uniform(%d, %d). Nilai E(X) adalah?", a, b), options: buildOptions(mean, []float64{float64(w) / 2, float64(b), float64(w*w) / 12, mean + 1}, optionCount, false)}
		case 1:
			variance := float64(w*w) / 12
			if !isExact2(variance) {
				continue
			}
			return question{prompt: fmt.Sprintf("X ~ Uniform(%d, %d). Nilai Var(X) adalah?", a, b), options: buildOptions(round2(variance), []float64{float64(a+b) / 2, round2(float64(w*w) / 4), float64(w), round2(variance + 1)}, optionCount, false)}
		default:
			c := between(a, b-1)
			d := between(c+1, b)
			p := float64(d-c) / float64(w)
			if !isExact2(p) {
				continue
			}
			return question{prompt: fmt.Sprintf("X ~ Uniform(%d, %d). Nilai P(%d < X < %d) adalah?", a, b, c, d), options: buildOptions(round2(p), []float64{round2(1 - p), round2(float64(d-c) / float64(b)), float64(d - c), round2(p / 2)}, optionCount, false)}
		}
	}
}

func genKBayes(optionCount int) question {
	for {
		p := pickOne([]float64{0.1, 0.2, 0.3, 0.4, 0.5})
		q := pickOne([]float64{0.5, 0.6, 0.7, 0.8, 0.9})
		r := pickOne([]float64{0.1, 0.2, 0.3, 0.4, 0.5})
		pb := p*q + (1-p)*r
		val := cleanFloat(p * q / pb)
		if !isExact2(val) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Sebuah tes penyakit: P(sakit) = %s, P(positif | sakit) = %s, P(positif | tidak sakit) = %s. Nilai P(sakit | positif) adalah?", fmtNum(p), fmtNum(q), fmtNum(r)),
			options: buildOptions(round2(val), []float64{q, round2(p * q), round2(pb), round2(1 - val)}, optionCount, false),
		}
	}
}

func genKPoissonGeometric(optionCount int) question {
	for {
		if rand.IntN(2) == 0 {
			l := between(1, 12)
			correct := float64(l + l*l)
			return question{prompt: fmt.Sprintf("X ~ Poisson(λ = %d). Nilai E(X²) adalah?", l), options: buildOptions(correct, []float64{float64(l), float64(l * l), float64(2 * l), correct + 1}, optionCount, false)}
		}
		p := pickOne([]float64{0.05, 0.1, 0.2, 0.25, 0.5})
		if rand.IntN(2) == 0 {
			return question{prompt: fmt.Sprintf("X menyatakan banyak percobaan sampai sukses pertama dengan peluang sukses %s (distribusi geometrik). Nilai E(X) adalah?", fmtNum(p)), options: buildOptions(round2(1/p), []float64{p, round2((1 - p) / p), round2(1/p + 1), round2(1 / (1 - p))}, optionCount, false)}
		}
		variance := cleanFloat((1 - p) / (p * p))
		if !isExact2(variance) {
			continue
		}
		return question{prompt: fmt.Sprintf("X menyatakan banyak percobaan sampai sukses pertama dengan peluang sukses %s (distribusi geometrik). Nilai Var(X) adalah?", fmtNum(p)), options: buildOptions(round2(variance), []float64{round2(1 / p), round2((1 - p) / p), round2(1 / (p * p)), round2(variance + 1)}, optionCount, false)}
	}
}

// ---------- Statistika Matematika ----------

func squareSampleSizes() []int { return []int{4, 9, 16, 25, 36, 49, 64, 100} }

func genKSampleVariance(optionCount int) question {
	n := pickOne([]int{4, 5, 6})
	mean := between(10, 80)
	devs := randomDeviations(n, 6, func(sumSq int) bool { return isExact2(float64(sumSq) / float64(n-1)) })
	data := make([]int, n)
	sumSq := 0
	for i, d := range devs {
		data[i] = mean + d
		sumSq += d * d
	}
	correct := round2(float64(sumSq) / float64(n-1))
	return question{
		prompt:  fmt.Sprintf("Variansi sampel (pembagi n - 1) dari data %s adalah?", joinInts(data)),
		options: buildOptions(correct, []float64{round2(float64(sumSq) / float64(n)), float64(sumSq), round2(math.Sqrt(correct)), round2(correct + 1)}, optionCount, false),
	}
}

func genKStandardError(optionCount int) question {
	for {
		sigma, n := between(2, 40), pickOne(squareSampleSizes())
		root := int(math.Sqrt(float64(n)))
		se := float64(sigma) / float64(root)
		if !isExact2(se) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Simpangan baku populasi σ = %d dan ukuran sampel n = %d. Galat baku (standard error) rata-rata sampel adalah?", sigma, n),
			options: buildOptions(round2(se), []float64{round2(float64(sigma) / float64(n)), float64(sigma), float64(sigma * root), round2(se + 1)}, optionCount, false),
		}
	}
}

func genKZStatistic(optionCount int) question {
	for {
		mu0, sigma, n := between(40, 100), between(2, 30), pickOne(squareSampleSizes())
		d := nonZeroBetween(-10, 10)
		root := int(math.Sqrt(float64(n)))
		z := float64(d*root) / float64(sigma)
		if !isExact2(z) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Uji H₀: μ = %d dengan σ = %d, n = %d, dan x̄ = %d. Nilai statistik uji z adalah?", mu0, sigma, n, mu0+d),
			options: buildOptions(round2(z), []float64{round2(float64(d) / float64(sigma)), -round2(z), round2(float64(d*n) / float64(sigma)), round2(z + 1)}, optionCount, true),
		}
	}
}

func genKConfidenceInterval(optionCount int) question {
	for {
		sigma, n := between(2, 30), pickOne(squareSampleSizes())
		xbar := between(20, 200)
		se := float64(sigma) / math.Sqrt(float64(n))
		half := cleanFloat(1.96 * se)
		if !isExact2(half) {
			continue
		}
		upper, lower := round2(float64(xbar)+half), round2(float64(xbar)-half)
		if rand.IntN(2) == 0 {
			return question{
				prompt:  fmt.Sprintf("Selang kepercayaan 95%% (z = 1.96) untuk μ dengan x̄ = %d, σ = %d, dan n = %d. Batas atas selang tersebut adalah?", xbar, sigma, n),
				options: buildOptions(upper, []float64{lower, round2(float64(xbar) + se), round2(float64(xbar) + 1.96*float64(sigma)), round2(upper + 1)}, optionCount, false),
			}
		}
		return question{
			prompt:  fmt.Sprintf("Selang kepercayaan 95%% (z = 1.96) untuk μ dengan x̄ = %d, σ = %d, dan n = %d. Batas bawah selang tersebut adalah?", xbar, sigma, n),
			options: buildOptions(lower, []float64{upper, round2(float64(xbar) - se), round2(float64(xbar) - 1.96*float64(sigma)), round2(lower - 1)}, optionCount, true),
		}
	}
}

func genKMLE(optionCount int) question {
	for {
		switch rand.IntN(3) {
		case 0:
			n := pickOne([]int{20, 25, 40, 50, 100})
			k := between(1, n-1)
			p := float64(k) / float64(n)
			if !isExact2(p) {
				continue
			}
			return question{prompt: fmt.Sprintf("Dari %d percobaan Bernoulli terdapat %d sukses. Penduga kemungkinan maksimum (MLE) untuk p adalah?", n, k), options: buildOptions(round2(p), []float64{round2(1 - p), float64(k), round2(float64(k) / float64(n-1)), round2(p / 2)}, optionCount, false)}
		case 1:
			data := make([]int, pickOne([]int{4, 5}))
			sum := 0
			for i := range data {
				data[i] = between(0, 9)
				sum += data[i]
			}
			mean := float64(sum) / float64(len(data))
			return question{prompt: fmt.Sprintf("Data %s berasal dari distribusi Poisson(λ). Penduga MLE untuk λ adalah?", joinInts(data)), options: buildOptions(round2(mean), []float64{float64(sum), round2(1 / math.Max(mean, 0.5)), round2(mean + 1), float64(len(data))}, optionCount, false)}
		default:
			xbar := pickOne([]float64{0.5, 2, 4, 5, 10, 20, 25})
			return question{prompt: fmt.Sprintf("Sampel dari distribusi eksponensial dengan parameter laju λ memiliki rata-rata %s. Penduga MLE untuk λ adalah?", fmtNum(xbar)), options: buildOptions(round2(1/xbar), []float64{xbar, round2(xbar * xbar), round2(2 / xbar), round2(1/xbar + 1)}, optionCount, false)}
		}
	}
}

func genKTStatistic(optionCount int) question {
	for {
		n, s := pickOne([]int{4, 9, 16, 25, 36}), between(2, 20)
		d := nonZeroBetween(-8, 8)
		mu0 := between(20, 100)
		root := int(math.Sqrt(float64(n)))
		t := float64(d*root) / float64(s)
		if !isExact2(t) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Sampel berukuran n = %d memiliki rata-rata %d dan simpangan baku s = %d. Nilai statistik uji t untuk H₀: μ = %d adalah?", n, mu0+d, s, mu0),
			options: buildOptions(round2(t), []float64{round2(float64(d) / float64(s)), -round2(t), round2(float64(d*n) / float64(s)), round2(t + 1)}, optionCount, true),
		}
	}
}

// ---------- Statistika Terapan & Regresi ----------

func genKRegressionSlope(optionCount int) question {
	ys := make([]int, 5)
	sumY, sxy := 0, 0
	pairs := make([]string, 5)
	for i := range ys {
		ys[i] = between(0, 30)
		sumY += ys[i]
		sxy += (i - 2) * ys[i] // x = 1..5, x̄ = 3
		pairs[i] = fmt.Sprintf("(%d, %d)", i+1, ys[i])
	}
	b := float64(sxy) / 10 // Sxx = 10
	ybar := float64(sumY) / 5
	a := cleanFloat(ybar - 3*b)
	data := strings.Join(pairs, ", ")
	switch rand.IntN(3) {
	case 0:
		return question{prompt: fmt.Sprintf("Data (x, y): %s. Dengan metode kuadrat terkecil y = a + bx, nilai b adalah?", data), options: buildOptions(round2(b), []float64{round2(a), float64(sxy), round2(ybar), round2(b + 1)}, optionCount, true)}
	case 1:
		return question{prompt: fmt.Sprintf("Data (x, y): %s. Dengan metode kuadrat terkecil y = a + bx, nilai a adalah?", data), options: buildOptions(round2(a), []float64{round2(b), round2(ybar), round2(ybar + 3*b), round2(a + 1)}, optionCount, true)}
	default:
		pred := cleanFloat(a + 6*b)
		return question{prompt: fmt.Sprintf("Data (x, y): %s. Dengan garis regresi kuadrat terkecil, prediksi y untuk x = 6 adalah?", data), options: buildOptions(round2(pred), []float64{round2(a + 5*b), round2(ybar), round2(6 * b), round2(pred + 1)}, optionCount, true)}
	}
}

func genKResidual(optionCount int) question {
	a := between(-5, 10)
	b := pickOne([]float64{0.5, 1, 1.5, 2, 2.5, 3})
	x := between(1, 10)
	yhat := float64(a) + b*float64(x)
	e := float64(nonZeroBetween(-10, 10)) / 2
	y := yhat + e
	return question{
		prompt:  fmt.Sprintf("Model regresi ŷ = %d + %sx. Untuk data (x, y) = (%d, %s), nilai residual e = y - ŷ adalah?", a, fmtNum(b), x, fmtNum(y)),
		options: buildOptions(e, []float64{-e, yhat, y, e + 1}, optionCount, true),
	}
}

func genKRSquared(optionCount int) question {
	r2 := pickOne([]float64{0.25, 0.4, 0.5, 0.6, 0.75, 0.8, 0.9, 0.95})
	sst := 20 * between(3, 30)
	sse := int(math.Round(float64(sst) * (1 - r2)))
	return question{
		prompt:  fmt.Sprintf("Model regresi memiliki SST = %d dan SSE = %d. Nilai koefisien determinasi R² adalah?", sst, sse),
		options: buildOptions(r2, []float64{round2(1 - r2), round2(math.Sqrt(r2)), round2(float64(sse) / float64(sst-sse)), round2(r2 - 0.1)}, optionCount, false),
	}
}

func genKCovariance(optionCount int) question {
	for {
		n := pickOne([]int{4, 5})
		xs, ys := make([]int, n), make([]int, n)
		sx, sy, sxy := 0, 0, 0
		pairs := make([]string, n)
		for i := range n {
			xs[i], ys[i] = between(0, 10), between(0, 10)
			sx += xs[i]
			sy += ys[i]
			sxy += xs[i] * ys[i]
			pairs[i] = fmt.Sprintf("(%d, %d)", xs[i], ys[i])
		}
		num := float64(n*sxy-sx*sy) / float64(n) // Σ(x - x̄)(y - ȳ)
		cov := cleanFloat(num / float64(n-1))
		if !isExact2(cov) || cov == 0 {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Kovariansi sampel (pembagi n - 1) dari data (x, y): %s adalah?", strings.Join(pairs, ", ")),
			options: buildOptions(round2(cov), []float64{round2(num / float64(n)), -round2(cov), float64(sxy), round2(cov + 1)}, optionCount, true),
		}
	}
}

func genKCorrelation(optionCount int) question {
	for {
		u, v := between(2, 10), between(2, 10)
		r := pickOne([]float64{-0.8, -0.6, -0.5, -0.25, 0.25, 0.4, 0.5, 0.6, 0.75, 0.8, 0.9})
		sxy := cleanFloat(r * float64(u*v))
		if !isExact2(sxy) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Diketahui Sxx = %d, Syy = %d, dan Sxy = %s. Nilai koefisien korelasi r adalah?", u*u, v*v, fmtNum(sxy)),
			options: buildOptions(r, []float64{round2(r * r), -r, round2(sxy / float64(u*u)), round2(sxy / float64(u*u*v*v))}, optionCount, true),
		}
	}
}

// ---------- Geometri Analitik ----------

func randomQuadrupleNormal() ([]int, int) {
	q := pickOne(pyQuadruples)
	n := []int{q.a * randSign(), q.b * randSign(), q.c * randSign()}
	rand.Shuffle(3, func(i, j int) { n[i], n[j] = n[j], n[i] })
	return n, q.d
}

func genKPointPlaneDistance(optionCount int) question {
	n, length := randomQuadrupleNormal()
	p := []int{between(-5, 5), between(-5, 5), between(-5, 5)}
	r := between(1, 6)
	d := randSign()*r*length - (n[0]*p[0] + n[1]*p[1] + n[2]*p[2])
	correct := float64(r)
	return question{
		prompt:  fmt.Sprintf("Jarak titik (%s) ke bidang %s = 0 adalah?", joinInts(p), formatTerms([]int{n[0], n[1], n[2], d}, []string{"x", "y", "z", ""})),
		options: buildOptions(correct, []float64{float64(r * length), float64(absInt(d)), correct + 1, round2(float64(r*length) / float64(absInt(n[0])+absInt(n[1])+absInt(n[2])))}, optionCount, false),
	}
}

func genKSphere(optionCount int) question {
	a, b, c := between(-5, 5), between(-5, 5), between(-5, 5)
	r := between(1, 9)
	eq := fmt.Sprintf("x² + y² + z²%s%s%s%s = 0", linTerm(-2*a, "x"), linTerm(-2*b, "y"), linTerm(-2*c, "z"), linTerm(a*a+b*b+c*c-r*r, ""))
	if rand.IntN(2) == 0 {
		return question{prompt: fmt.Sprintf("Jari-jari bola %s adalah?", eq), options: buildOptions(float64(r), []float64{float64(r * r), float64(r + 1), float64(absInt(a*a + b*b + c*c - r*r)), float64(r - 1)}, optionCount, false)}
	}
	return question{prompt: fmt.Sprintf("Koordinat x titik pusat bola %s adalah?", eq), options: buildOptions(float64(a), []float64{float64(-a), float64(2 * a), float64(-2 * a), float64(a + 1)}, optionCount, true)}
}

func genKParallelepiped(optionCount int) question {
	for {
		m := make([][]int, 3)
		for i := range m {
			m[i] = []int{between(-3, 4), between(-3, 4), between(-3, 4)}
		}
		det := det3(m)
		if det == 0 {
			continue
		}
		vecs := fmt.Sprintf("u = (%s), v = (%s), dan w = (%s)", joinInts(m[0]), joinInts(m[1]), joinInts(m[2]))
		if rand.IntN(2) == 0 {
			return question{prompt: fmt.Sprintf("Volume paralelepipedum yang dibentuk oleh vektor %s adalah?", vecs), options: buildOptions(float64(absInt(det)), []float64{round2(float64(absInt(det)) / 6), float64(absInt(det) * 2), float64(absInt(det) + 1), round2(float64(absInt(det)) / 2)}, optionCount, false)}
		}
		vol := float64(absInt(det)) / 6
		if !isExact2(vol) {
			continue
		}
		return question{prompt: fmt.Sprintf("Volume bidang empat (tetrahedron) yang dibentuk oleh vektor %s dari titik asal adalah?", vecs), options: buildOptions(round2(vol), []float64{float64(absInt(det)), round2(float64(absInt(det)) / 3), round2(float64(absInt(det)) / 2), round2(vol + 1)}, optionCount, false)}
	}
}

func genKParallelogramArea(optionCount int) question {
	for {
		x1, y1, x2, y2 := between(-6, 6), between(-6, 6), between(-6, 6), between(-6, 6)
		det := absInt(x1*y2 - x2*y1)
		if det == 0 {
			continue
		}
		if rand.IntN(2) == 0 {
			return question{prompt: fmt.Sprintf("Luas jajargenjang yang dibentuk oleh vektor u = (%d, %d) dan v = (%d, %d) adalah?", x1, y1, x2, y2), options: buildOptions(float64(det), []float64{float64(det) / 2, float64(absInt(x1*y2 + x2*y1)), float64(det + 1), float64(absInt(x1*x2 + y1*y2))}, optionCount, false)}
		}
		ox, oy := between(-5, 5), between(-5, 5)
		return question{prompt: fmt.Sprintf("Luas segitiga dengan titik sudut (%d, %d), (%d, %d), dan (%d, %d) adalah?", ox, oy, ox+x1, oy+y1, ox+x2, oy+y2), options: buildOptions(float64(det)/2, []float64{float64(det), float64(det)/2 + 1, float64(det) / 4, float64(absInt(x1*y2+x2*y1)) / 2}, optionCount, false)}
	}
}

func genKAngleBetweenPlanes(optionCount int) question {
	u, v := transformedAnglePair()
	angle := vectorAngleDeg(u, v)
	if angle > 90 {
		angle = 180 - angle
	}
	names := []string{"x", "y", "z"}
	return question{
		prompt: fmt.Sprintf(
			"Besar sudut lancip antara bidang %s = %d dan bidang %s = %d adalah? (derajat)",
			formatTerms(u, names), between(-9, 9), formatTerms(v, names), between(-9, 9),
		),
		options: buildOptions(angle, []float64{0, 30, 45, 60, 90}, optionCount, false),
	}
}

// transformedAnglePair: pasangan vektor 3D bersudut "cantik" (45/60/90/120/135°).
func transformedAnglePair() ([]int, []int) {
	base := pickOne([][2][3]int{
		{{1, 0, 0}, {1, 1, 0}}, {{1, 1, 0}, {0, 1, 1}}, {{1, 1, 0}, {1, -1, 0}}, {{1, 0, 0}, {-1, 1, 0}},
		{{1, 1, 0}, {-1, 0, -1}}, {{1, 2, 2}, {2, 1, -2}}, {{0, 1, 1}, {1, 0, 1}},
	})
	perm := rand.Perm(3)
	signs := []int{randSign(), randSign(), randSign()}
	k1, k2 := between(1, 3), between(1, 3)
	u, v := make([]int, 3), make([]int, 3)
	for i := range 3 {
		u[i] = base[0][perm[i]] * signs[i] * k1
		v[i] = base[1][perm[i]] * signs[i] * k2
	}
	return u, v
}

func vectorAngleDeg(u, v []int) float64 {
	dot, nu, nv := 0.0, 0.0, 0.0
	for i := range u {
		dot += float64(u[i] * v[i])
		nu += float64(u[i] * u[i])
		nv += float64(v[i] * v[i])
	}
	return math.Round(math.Acos(dot/math.Sqrt(nu*nv)) * 180 / math.Pi)
}

func genKLinePlaneIntersection(optionCount int) question {
	for {
		p0 := []int{between(-5, 5), between(-5, 5), between(-5, 5)}
		dir := []int{between(-3, 3), between(-3, 3), between(-3, 3)}
		n := []int{between(-3, 3), between(-3, 3), between(-3, 3)}
		nd := n[0]*dir[0] + n[1]*dir[1] + n[2]*dir[2]
		if nd == 0 {
			continue
		}
		t0 := nonZeroBetween(-3, 4)
		p := []int{p0[0] + t0*dir[0], p0[1] + t0*dir[1], p0[2] + t0*dir[2]}
		s := n[0]*p[0] + n[1]*p[1] + n[2]*p[2]
		i := rand.IntN(3)
		axis := []string{"x", "y", "z"}[i]
		correct := float64(p[i])
		return question{
			prompt: fmt.Sprintf(
				"Garis (x, y, z) = (%s) + t(%s) memotong bidang %s = %d. Koordinat %s titik potongnya adalah?",
				joinInts(p0), joinInts(dir), formatTerms(n, []string{"x", "y", "z"}), s, axis,
			),
			options: buildOptions(correct, []float64{float64(p0[i]), float64(t0), float64(p[(i+1)%3]), correct + float64(dir[i]), correct + 1}, optionCount, true),
		}
	}
}

// ---------- Geometri Transformasi ----------

func genKRotationComposition(optionCount int) question {
	if rand.IntN(2) == 0 {
		angles := []int{30, 45, 60, 90, 120, 135, 150, 180, 210, 270}
		t1, t2 := pickOne(angles), pickOne(angles)
		correct := float64((t1 + t2) % 360)
		return question{
			prompt:  fmt.Sprintf("Rotasi %d° terhadap titik asal dilanjutkan rotasi %d° (keduanya berlawanan arah jarum jam) setara dengan satu rotasi sebesar? (derajat, 0 ≤ θ < 360)", t1, t2),
			options: buildOptions(correct, []float64{float64(absInt(t1 - t2)), float64((360 - (t1+t2)%360) % 360), float64((t1 + t2 + 180) % 360), float64(t1 + t2)}, optionCount, false),
		}
	}
	a := 15 * between(0, 11)
	b := 15 * between(0, 11)
	for b == a {
		b = 15 * between(0, 11)
	}
	correct := float64(((2*(b-a))%360 + 360) % 360)
	return question{
		prompt:  fmt.Sprintf("Pencerminan terhadap garis melalui titik asal bersudut %d° (terhadap sumbu x positif), dilanjutkan pencerminan terhadap garis bersudut %d°, setara dengan rotasi sebesar? (derajat, 0 ≤ θ < 360)", a, b),
		options: buildOptions(correct, []float64{float64(((b-a)%360 + 360) % 360), float64(((2*(a-b))%360 + 360) % 360), float64((int(correct) + 180) % 360), float64(a + b)}, optionCount, false),
	}
}

// reflectionLine: garis y = mx dengan cos 2α & sin 2α rasional (hasil pencerminan desimal pas).
type reflectionLine struct {
	text     string
	cos, sin float64
}

var reflectionLines = []reflectionLine{
	{"y = 2x", -0.6, 0.8}, {"y = x/2", 0.6, 0.8}, {"y = 3x", -0.8, 0.6}, {"y = x/3", 0.8, 0.6},
	{"y = -2x", -0.6, -0.8}, {"y = -x/2", 0.6, -0.8},
}

func reflectPoint(l reflectionLine, x, y int) (float64, float64) {
	return cleanFloat(l.cos*float64(x) + l.sin*float64(y)), cleanFloat(l.sin*float64(x) - l.cos*float64(y))
}

func genKReflectLineRational(optionCount int) question {
	l := pickOne(reflectionLines)
	x, y := nonZeroBetween(-9, 9), nonZeroBetween(-9, 9)
	rx, ry := reflectPoint(l, x, y)
	axis, correct, other, swapped := "x", rx, ry, float64(y)
	if rand.IntN(2) == 0 {
		axis, correct, other, swapped = "y", ry, rx, float64(x)
	}
	return question{
		prompt:  fmt.Sprintf("Titik P(%d, %d) dicerminkan terhadap garis %s. Koordinat %s bayangan titik P adalah?", x, y, l.text, axis),
		options: buildOptions(round2(correct), []float64{round2(other), swapped, -round2(correct), round2(correct + 1)}, optionCount, true),
	}
}

func genKAreaUnderMatrix(optionCount int) question {
	for {
		pts := [][2]int{{between(-4, 4), between(-4, 4)}, {between(-4, 4), between(-4, 4)}, {between(-4, 4), between(-4, 4)}}
		twice := absInt((pts[1][0]-pts[0][0])*(pts[2][1]-pts[0][1]) - (pts[2][0]-pts[0][0])*(pts[1][1]-pts[0][1]))
		m := [][]int{{between(-3, 3), between(-3, 3)}, {between(-3, 3), between(-3, 3)}}
		det := m[0][0]*m[1][1] - m[0][1]*m[1][0]
		if twice == 0 || det == 0 {
			continue
		}
		area := float64(twice) / 2
		correct := float64(absInt(det)) * area
		return question{
			prompt: fmt.Sprintf(
				"Segitiga dengan titik sudut (%d, %d), (%d, %d), dan (%d, %d) ditransformasi oleh matriks %s (baris dipisah titik koma). Luas bayangannya adalah?",
				pts[0][0], pts[0][1], pts[1][0], pts[1][1], pts[2][0], pts[2][1], formatMatrix(m),
			),
			options: buildOptions(correct, []float64{area, float64(det*det) * area, correct + float64(absInt(det)), float64(twice * absInt(det))}, optionCount, false),
		}
	}
}

func genKDilationCenter(optionCount int) question {
	a, b := between(-5, 5), between(-5, 5)
	x, y := between(-8, 8), between(-8, 8)
	k := pickOne([]float64{-2, -1, 2, 3, 0.5, -0.5, 1.5})
	rx, ry := float64(a)+k*float64(x-a), float64(b)+k*float64(y-b)
	axis, correct, other, originGuess := "x", rx, ry, k*float64(x)
	if rand.IntN(2) == 0 {
		axis, correct, other, originGuess = "y", ry, rx, k*float64(y)
	}
	return question{
		prompt:  fmt.Sprintf("Titik P(%d, %d) didilatasi dengan pusat (%d, %d) dan faktor skala %s. Koordinat %s bayangannya adalah?", x, y, a, b, fmtNum(k), axis),
		options: buildOptions(round2(correct), []float64{round2(other), round2(originGuess), -round2(correct), round2(correct + 1)}, optionCount, true),
	}
}

// ---------- Topologi Dasar ----------

func genKEulerCharacteristic(optionCount int) question {
	switch rand.IntN(3) {
	case 0:
		g := between(0, 6)
		correct := float64(2 - 2*g)
		return question{prompt: fmt.Sprintf("Karakteristik Euler permukaan tertutup terorientasi bergenus %d adalah?", g), options: buildOptions(correct, []float64{float64(2 - g), float64(2 * g), float64(2 + 2*g), correct - 1}, optionCount, true)}
	case 1:
		k := between(1, 5)
		correct := float64(2 - k)
		return question{prompt: fmt.Sprintf("Karakteristik Euler permukaan tertutup tak terorientasi dengan %d crosscap adalah?", k), options: buildOptions(correct, []float64{float64(2 - 2*k), float64(k), float64(2 + k), correct + 1}, optionCount, true)}
	default:
		type poly struct {
			name    string
			v, e, f int
		}
		n := between(3, 8)
		p := pickOne([]poly{
			{"tetrahedron", 4, 6, 4}, {"kubus", 8, 12, 6}, {"oktahedron", 6, 12, 8}, {"dodekahedron", 20, 30, 12}, {"ikosahedron", 12, 30, 20},
			{fmt.Sprintf("prisma segi-%d", n), 2 * n, 3 * n, n + 2}, {fmt.Sprintf("limas segi-%d", n), n + 1, 2 * n, n + 1},
		})
		return question{prompt: fmt.Sprintf("Sebuah %s memiliki %d titik sudut dan %d rusuk. Menurut rumus Euler, banyak sisinya (muka) adalah?", p.name, p.v, p.e), options: buildOptions(float64(p.f), []float64{float64(p.e - p.v), float64(p.f + 2), float64(p.v), float64(p.f - 1)}, optionCount, false)}
	}
}

func genKIntervalSet(optionCount int) question {
	pos := between(-5, 0)
	var parts []string
	intervals, points, length := 0, 0, 0
	for range between(3, 5) {
		if rand.IntN(3) == 0 {
			parts = append(parts, fmt.Sprintf("{%d}", pos))
			points++
			pos += between(2, 3)
			continue
		}
		w := between(1, 4)
		left, right := pickOne([]string{"[", "("}), pickOne([]string{"]", ")"})
		parts = append(parts, fmt.Sprintf("%s%d, %d%s", left, pos, pos+w, right))
		intervals++
		length += w
		pos += w + between(1, 3)
	}
	set := strings.Join(parts, " ∪ ")
	components := intervals + points
	boundary := 2*intervals + points
	switch rand.IntN(3) {
	case 0:
		return question{prompt: fmt.Sprintf("A = %s ⊂ ℝ (topologi biasa). Banyak komponen terhubung A adalah?", set), options: buildOptions(float64(components), []float64{float64(intervals), float64(boundary), float64(components + 1), float64(points)}, optionCount, false)}
	case 1:
		return question{prompt: fmt.Sprintf("A = %s ⊂ ℝ (topologi biasa). Banyak titik batas (boundary) A adalah?", set), options: buildOptions(float64(boundary), []float64{float64(components), float64(2 * components), float64(boundary - points), float64(boundary + 1)}, optionCount, false)}
	default:
		return question{prompt: fmt.Sprintf("A = %s ⊂ ℝ (topologi biasa). Panjang total interior A adalah?", set), options: buildOptions(float64(length), []float64{float64(length + points), float64(length - 1), float64(length + 1), float64(intervals)}, optionCount, false)}
	}
}

func genKFiniteTopology(optionCount int) question {
	switch rand.IntN(3) {
	case 0:
		n := between(1, 6)
		correct := float64(intPow(2, n))
		return question{prompt: fmt.Sprintf("Banyak himpunan terbuka pada topologi diskrit di himpunan dengan %d anggota adalah?", n), options: buildOptions(correct, []float64{2, float64(n), float64(n + 1), correct * 2}, optionCount, false)}
	case 1:
		n := between(2, 6)
		return question{prompt: fmt.Sprintf("Banyak himpunan terbuka pada topologi trivial (indiskrit) di himpunan dengan %d anggota adalah?", n), options: buildOptions(2, []float64{1, float64(n), float64(intPow(2, n)), float64(n + 1)}, optionCount, false)}
	default:
		n := between(1, 4)
		count := []int{0, 1, 4, 29, 355}[n]
		return question{prompt: fmt.Sprintf("Banyak topologi berbeda yang dapat dibentuk pada himpunan dengan %d anggota adalah?", n), options: buildOptions(float64(count), []float64{float64(intPow(2, n)), float64(intPow(2, intPow(2, n))), float64(factorial(n)), float64(count + 1)}, optionCount, false)}
	}
}

func genKBettiGenus(optionCount int) question {
	if rand.IntN(2) == 0 {
		n := between(1, 8)
		return question{prompt: fmt.Sprintf("Grup fundamental dari buket (wedge) %d lingkaran adalah grup bebas dengan rank?", n), options: buildOptions(float64(n), []float64{float64(n + 1), float64(2 * n), 1, float64(n - 1)}, optionCount, false)}
	}
	g := between(1, 6)
	return question{prompt: fmt.Sprintf("Bilangan Betti pertama b₁ dari permukaan tertutup terorientasi bergenus %d adalah?", g), options: buildOptions(float64(2*g), []float64{float64(g), float64(2 - 2*g), float64(2*g + 1), float64(g + 1)}, optionCount, true)}
}

// ---------- Metode Matematika Teknik & Fisika ----------

func genKFourierCoeff(optionCount int) question {
	for {
		if rand.IntN(2) == 0 {
			// f(x) = cx pada (-π, π): b_n = 2c(-1)^(n+1)/n
			c, n := nonZeroBetween(-4, 5), between(1, 6)
			sign := 1
			if n%2 == 0 {
				sign = -1
			}
			val := float64(2*c*sign) / float64(n)
			if !isExact2(val) {
				continue
			}
			return question{prompt: fmt.Sprintf("Deret Fourier f(x) = %s pada (-π, π) memiliki koefisien sinus b_n. Nilai b_%d = ?", formatPoly(c, 0), n), options: buildOptions(round2(val), []float64{-round2(val), round2(float64(c) / float64(n)), float64(2 * c), 0}, optionCount, true)}
		}
		// gelombang kotak amplitudo A: b_n = 4A/(nπ) untuk n ganjil
		a, n := between(1, 5), pickOne([]int{1, 3, 5})
		val := float64(4*a) / float64(n)
		if !isExact2(val) {
			continue
		}
		return question{prompt: fmt.Sprintf("Gelombang kotak f(x) = %d untuk 0 < x < π dan f(x) = -%d untuk -π < x < 0 memiliki koefisien sinus b_%d = k/π. Nilai k = ?", a, a, n), options: buildOptions(round2(val), []float64{round2(float64(2*a) / float64(n)), float64(4 * a), 0, round2(val + 1)}, optionCount, false)}
	}
}

func genKGammaBeta(optionCount int) question {
	for {
		switch rand.IntN(3) {
		case 0:
			n := between(2, 8)
			correct := float64(factorial(n - 1))
			return question{prompt: fmt.Sprintf("Nilai fungsi gamma Γ(%d) adalah?", n), options: buildOptions(correct, []float64{float64(factorial(n)), float64(factorial(n - 2)), float64(n - 1), correct + 1}, optionCount, false)}
		case 1:
			m, n := between(1, 5), between(1, 5)
			val := float64(factorial(m-1)*factorial(n-1)) / float64(factorial(m+n-1))
			if !isExact2(val) {
				continue
			}
			return question{prompt: fmt.Sprintf("Nilai fungsi beta B(%d, %d) adalah?", m, n), options: buildOptions(round2(val), []float64{round2(float64(factorial(m)*factorial(n)) / float64(factorial(m+n))), float64(factorial(m - 1)), round2(val * 2), 1}, optionCount, false)}
		default:
			if rand.IntN(2) == 0 {
				return question{prompt: "Γ(3/2) = k√π. Nilai k adalah?", options: buildOptions(0.5, []float64{1, 0.75, 1.5, 0.25}, optionCount, false)}
			}
			return question{prompt: "Γ(5/2) = k√π. Nilai k adalah?", options: buildOptions(0.75, []float64{0.5, 1.5, 2.5, 1.25}, optionCount, false)}
		}
	}
}

func genKDiracDelta(optionCount int) question {
	a, b, c, d := between(-3, 3), between(-5, 5), between(-6, 6), between(-9, 9)
	for a == 0 && b == 0 {
		b = between(-5, 5)
	}
	k := nonZeroBetween(-3, 3)
	p := func(x int) int { return a*x*x*x + b*x*x + c*x + d }
	correct := float64(p(k))
	return question{
		prompt:  fmt.Sprintf("∫ dari -∞ sampai ∞ (%s)·δ(%s) dx = ?", formatPoly(a, b, c, d), formatPoly(1, -k)),
		options: buildOptions(correct, []float64{float64(p(-k)), float64(d), correct + 1, -correct}, optionCount, true),
	}
}

func genKJacobian(optionCount int) question {
	switch rand.IntN(3) {
	case 0:
		var a, b, c, d, det int
		for det == 0 {
			a, b, c, d = between(-5, 5), between(-5, 5), between(-5, 5), between(-5, 5)
			det = a*d - b*c
		}
		return question{prompt: fmt.Sprintf("Transformasi x = %s, y = %s. Nilai |∂(x, y)/∂(u, v)| adalah?", formatTerms([]int{a, b}, []string{"u", "v"}), formatTerms([]int{c, d}, []string{"u", "v"})), options: buildOptions(float64(absInt(det)), []float64{float64(absInt(a*d + b*c)), float64(absInt(a + d)), float64(absInt(det) + 1), float64(absInt(a*b - c*d))}, optionCount, false)}
	case 1:
		r := between(1, 12)
		return question{prompt: fmt.Sprintf("Jacobian transformasi koordinat polar (x = r cos θ, y = r sin θ) di r = %d adalah?", r), options: buildOptions(float64(r), []float64{float64(r * r), 1, float64(2 * r), float64(r + 1)}, optionCount, false)}
	default:
		rho := between(1, 9)
		return question{prompt: fmt.Sprintf("Jacobian koordinat bola ρ² sin φ di ρ = %d dan φ = 90° adalah?", rho), options: buildOptions(float64(rho*rho), []float64{float64(rho), 0, float64(2 * rho), float64(rho*rho + 1)}, optionCount, false)}
	}
}

func genKTaylorCoeff(optionCount int) question {
	for {
		k := between(2, 4)
		switch rand.IntN(3) {
		case 0:
			a, n := nonZeroBetween(-3, 3), between(k, 8)
			correct := float64(combination(n, k) * intPow(a, k))
			return question{prompt: fmt.Sprintf("Koefisien x^%d pada ekspansi (%s)^%d adalah?", k, formatPoly(a, 1), n), options: buildOptions(correct, []float64{float64(combination(n, k)), float64(intPow(a, k)), float64(combination(n, k-1) * intPow(a, k-1)), -correct}, optionCount, true)}
		case 1:
			a := nonZeroBetween(-4, 4)
			val := float64(intPow(a, k)) / float64(factorial(k))
			if !isExact2(val) {
				continue
			}
			return question{prompt: fmt.Sprintf("Koefisien x^%d pada deret Maclaurin e^(%s) adalah?", k, formatPoly(a, 0)), options: buildOptions(round2(val), []float64{float64(intPow(a, k)), round2(float64(a) / float64(factorial(k))), round2(val * 2), 1}, optionCount, true)}
		default:
			a := nonZeroBetween(-4, 4)
			correct := float64(intPow(a, k))
			return question{prompt: fmt.Sprintf("Koefisien x^%d pada deret Maclaurin 1/(%s) adalah?", k, formatPoly(-a, 1)), options: buildOptions(correct, []float64{float64(-intPow(a, k)), float64(a * k), float64(intPow(a, k+1)), correct + 1}, optionCount, true)}
		}
	}
}

func genKLaplaceInverse(optionCount int) question {
	for {
		a, b := nonZeroBetween(-5, 5), nonZeroBetween(-5, 5)
		p, q := between(-6, 6), between(-9, 9)
		if a == b || (p == 0 && q == 0) {
			continue
		}
		A := float64(p*a+q) / float64(a-b)
		B := float64(p*b+q) / float64(b-a)
		if !isExact2(A) || !isExact2(B) {
			continue
		}
		return question{
			prompt: fmt.Sprintf(
				"L⁻¹{(%s) / ((%s)(%s))} = Ae^(%s) + Be^(%s). Nilai A = ?",
				formatPolyVar("s", p, q), formatPolyVar("s", 1, -a), formatPolyVar("s", 1, -b), formatPolyVar("t", a, 0), formatPolyVar("t", b, 0),
			),
			options: buildOptions(round2(A), []float64{round2(B), -round2(A), float64(p), round2(A + 1)}, optionCount, true),
		}
	}
}

// ---------- Riset Operasi ----------

func genKLPBox(optionCount int) question {
	a := between(6, 20)
	b := between(a/2+1, a-1)
	c := between(a-b+1, a-1)
	p, q := between(2, 9), between(2, 9)
	cons := []lpConstraint{{1, 1, a}, {1, 0, b}, {0, 1, c}}
	values := lpObjectiveValues(cons, false, p, q)
	best := values[len(values)-1]
	return question{
		prompt:  fmt.Sprintf("Maksimumkan Z = %dx + %dy dengan kendala x + y ≤ %d, x ≤ %d, y ≤ %d, x ≥ 0, y ≥ 0. Nilai Z maksimum adalah?", p, q, a, b, c),
		options: buildOptions(best, append(values[:len(values)-1], float64(p*b+q*c), best+1), optionCount, false),
	}
}

func genKTransportation2x2(optionCount int) question {
	s1, s2 := between(10, 60), between(10, 60)
	d1 := between(5, s1+s2-5)
	d2 := s1 + s2 - d1
	c := []int{between(1, 20), between(1, 20), between(1, 20), between(1, 20)}
	cost := func(x11 int) int {
		x12, x21 := s1-x11, d1-x11
		x22 := s2 - x21
		return c[0]*x11 + c[1]*x12 + c[2]*x21 + c[3]*x22
	}
	lo, hi := max(0, d1-s2), min(s1, d1)
	best, worst := cost(lo), cost(lo)
	for x := lo; x <= hi; x++ {
		best = min(best, cost(x))
		worst = max(worst, cost(x))
	}
	northwest := cost(min(s1, d1))
	return question{
		prompt: fmt.Sprintf(
			"Masalah transportasi: pabrik P1 (pasokan %d) dan P2 (pasokan %d) ke gudang G1 (permintaan %d) dan G2 (permintaan %d). Biaya per unit: P1→G1 = %d, P1→G2 = %d, P2→G1 = %d, P2→G2 = %d. Biaya total minimum adalah?",
			s1, s2, d1, d2, c[0], c[1], c[2], c[3],
		),
		options: buildOptions(float64(best), []float64{float64(worst), float64(northwest), float64(best + 1), float64(best + c[0])}, optionCount, false),
	}
}

func permutations(n int) [][]int {
	if n == 1 {
		return [][]int{{0}}
	}
	var out [][]int
	for _, p := range permutations(n - 1) {
		for pos := 0; pos <= len(p); pos++ {
			q := append(append(append([]int{}, p[:pos]...), n-1), p[pos:]...)
			out = append(out, q)
		}
	}
	return out
}

// assignmentCosts: biaya minimum, maksimum, dan diagonal utama (penugasan i -> i).
func assignmentCosts(c [][]int) (best, worst, diagonal int) {
	n := len(c)
	best, worst = 1<<30, -1
	for _, p := range permutations(n) {
		total := 0
		for i, j := range p {
			total += c[i][j]
		}
		best, worst = min(best, total), max(worst, total)
	}
	for i := range n {
		diagonal += c[i][i]
	}
	return
}

func genKAssignment3(optionCount int) question {
	c := make([][]int, 3)
	rows := make([]string, 3)
	for i := range c {
		c[i] = []int{between(1, 20), between(1, 20), between(1, 20)}
		rows[i] = fmt.Sprintf("%c = [%s]", 'A'+i, joinInts(c[i]))
	}
	best, worst, diag := assignmentCosts(c)
	return question{
		prompt:  fmt.Sprintf("Tiga pekerja ditugaskan ke tiga pekerjaan (1, 2, 3), satu pekerja satu pekerjaan, dengan biaya: %s. Biaya penugasan total minimum adalah?", strings.Join(rows, ", ")),
		options: buildOptions(float64(best), []float64{float64(worst), float64(diag), float64(best + 1), float64(best + 2)}, optionCount, false),
	}
}

func genKQueueMM1(optionCount int) question {
	for {
		mu := between(3, 20)
		lambda := between(1, mu-1)
		base := fmt.Sprintf("Antrian M/M/1 dengan laju kedatangan λ = %d pelanggan/jam dan laju pelayanan μ = %d pelanggan/jam.", lambda, mu)
		switch rand.IntN(4) {
		case 0:
			rho := float64(lambda) / float64(mu)
			if !isExact2(rho) {
				continue
			}
			return question{prompt: base + " Tingkat utilisasi ρ adalah?", options: buildOptions(round2(rho), []float64{round2(1 - rho), round2(float64(mu) / float64(lambda)), round2(rho / 2), round2(rho * rho)}, optionCount, false)}
		case 1:
			l := float64(lambda) / float64(mu-lambda)
			if !isExact2(l) {
				continue
			}
			return question{prompt: base + " Rata-rata banyak pelanggan dalam sistem (L) adalah?", options: buildOptions(round2(l), []float64{round2(float64(lambda*lambda) / float64(mu*(mu-lambda))), round2(float64(lambda) / float64(mu)), round2(l + 1), round2(1 / float64(mu-lambda))}, optionCount, false)}
		case 2:
			lq := float64(lambda*lambda) / float64(mu*(mu-lambda))
			if !isExact2(lq) {
				continue
			}
			return question{prompt: base + " Rata-rata panjang antrian (Lq) adalah?", options: buildOptions(round2(lq), []float64{round2(float64(lambda) / float64(mu-lambda)), round2(float64(lambda) / float64(mu)), round2(lq + 1), round2(lq * 2)}, optionCount, false)}
		default:
			w := 60 / float64(mu-lambda)
			if !isExact2(w) {
				continue
			}
			return question{prompt: base + " Rata-rata waktu pelanggan dalam sistem (W) adalah? (menit)", options: buildOptions(round2(w), []float64{round2(60 * float64(lambda) / float64(mu*(mu-lambda))), round2(w / 60), round2(60 / float64(mu)), round2(w + 1)}, optionCount, false)}
		}
	}
}

func genKEOQ(optionCount int) question {
	for {
		q := between(2, 30) * 10
		h, s := between(1, 20), between(10, 200)
		num := q * q * h
		if num%(2*s) != 0 {
			continue
		}
		d := num / (2 * s)
		if d < 100 || d > 100000 {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Permintaan tahunan %d unit, biaya pesan Rp%d ribu per pesanan, dan biaya simpan Rp%d ribu per unit per tahun. Jumlah pesanan ekonomis (EOQ) adalah? (unit)", d, s, h),
			options: buildOptions(float64(q), []float64{float64(2 * d * s / h), float64(2 * q), round2(math.Sqrt(float64(d*s) / float64(h))), float64(q + 10)}, optionCount, false),
		}
	}
}

// ---------- Matematika Keuangan Lanjut ----------

func genKNPV(optionCount int) question {
	r := pickOne([]int{10, 20})
	a, b := between(1, 20), between(1, 20)
	cf1 := a * (100 + r) * 10             // PV = 1000a
	cf2 := b * (100 + r) * (100 + r) / 10 // PV = 1000b
	i0 := between(1, a+b+5) * 1000
	correct := float64(1000*(a+b) - i0)
	return question{
		prompt: fmt.Sprintf(
			"Investasi awal Rp%d ribu menghasilkan arus kas Rp%d ribu di akhir tahun 1 dan Rp%d ribu di akhir tahun 2. Dengan tingkat diskonto %d%% per tahun, NPV investasi tersebut adalah? (ribu rupiah)",
			i0, cf1, cf2, r,
		),
		options: buildOptions(correct, []float64{float64(cf1 + cf2 - i0), -correct, float64(1000 * (a + b)), correct + 100}, optionCount, true),
	}
}

func genKFVAnnuity(optionCount int) question {
	p := between(1, 50) * 100
	i := pickOne([]int{10, 20})
	n := between(2, 3)
	fv := p * (intPow(100+i, n) - intPow(100, n)) / (i * intPow(100, n-1))
	return question{
		prompt:  fmt.Sprintf("Setiap akhir tahun ditabung Rp%d ribu dengan bunga majemuk %d%% per tahun. Nilai tabungan tepat setelah setoran ke-%d adalah? (ribu rupiah)", p, i, n),
		options: buildOptions(float64(fv), []float64{float64(p * n), float64(p * n * (100 + i) / 100), float64(fv + p), float64(fv - p*i/100)}, optionCount, false),
	}
}

func genKPerpetuity(optionCount int) question {
	for {
		p := between(1, 100) * 100
		i := pickOne([]float64{4, 5, 8, 10, 12.5, 20, 25})
		pv := float64(p) / (i / 100)
		if !isExact2(cleanFloat(pv)) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Nilai sekarang perpetuitas (anuitas abadi) dengan pembayaran Rp%d ribu per tahun dan suku bunga %s%% per tahun adalah? (ribu rupiah)", p, fmtNum(i)),
			options: buildOptions(round2(pv), []float64{round2(float64(p) * i / 100), round2(float64(p) / (1 + i/100)), round2(pv * 2), round2(pv + float64(p))}, optionCount, false),
		}
	}
}

func genKEffectiveRate(optionCount int) question {
	for {
		r := pickOne([]int{4, 6, 8, 10, 12, 20})
		m := pickOne([]int{2, 4, 12})
		eff := cleanFloat((math.Pow(1+float64(r)/float64(100*m), float64(m)) - 1) * 100)
		if !isExact2(eff) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Suku bunga nominal %d%% per tahun dimajemukkan %d kali setahun. Suku bunga efektif tahunannya adalah? (%%)", r, m),
			options: buildOptions(round2(eff), []float64{float64(r), round2(float64(r) / float64(m)), round2(eff + 0.5), float64(r * m)}, optionCount, false),
		}
	}
}

func genKPayback(optionCount int) question {
	for {
		i0 := between(10, 200) * 100
		cf := between(5, 100) * 100
		val := float64(i0) / float64(cf)
		if !isExact2(val) || val < 1 {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Investasi Rp%d ribu menghasilkan arus kas bersih konstan Rp%d ribu per tahun. Periode pengembalian (payback period) adalah? (tahun)", i0, cf),
			options: buildOptions(round2(val), []float64{round2(float64(cf) / float64(i0)), round2(val + 1), round2(val / 2), float64(i0-cf) / 100}, optionCount, false),
		}
	}
}

// ---------- Aljabar Linear Terapan (Data Science) ----------

func vectorWithNorm() ([]int, int) {
	if rand.IntN(2) == 0 {
		t := pickOne(pyTriples)
		v := []int{t.a * randSign(), t.b * randSign(), 0}
		rand.Shuffle(3, func(i, j int) { v[i], v[j] = v[j], v[i] })
		return v, t.c
	}
	q := pickOne(pyQuadruples)
	v := []int{q.a * randSign(), q.b * randSign(), q.c * randSign()}
	rand.Shuffle(3, func(i, j int) { v[i], v[j] = v[j], v[i] })
	return v, q.d
}

func genKCosineSimilarity(optionCount int) question {
	for {
		u, nu := vectorWithNorm()
		v, nv := vectorWithNorm()
		dot := u[0]*v[0] + u[1]*v[1] + u[2]*v[2]
		val := float64(dot) / float64(nu*nv)
		if dot == 0 || !isExact2(val) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Cosine similarity antara vektor u = (%s) dan v = (%s) adalah?", joinInts(u), joinInts(v)),
			options: buildOptions(round2(val), []float64{float64(dot), -round2(val), round2(float64(dot) / float64(nu*nu)), 1}, optionCount, true),
		}
	}
}

func genKVectorNorms(optionCount int) question {
	v, l2 := vectorWithNorm()
	l1, linf := 0, 0
	for _, x := range v {
		l1 += absInt(x)
		linf = max(linf, absInt(x))
	}
	cands := []float64{float64(l1), float64(l2), float64(linf), float64(l2 * l2)}
	switch rand.IntN(3) {
	case 0:
		return question{prompt: fmt.Sprintf("Norma L1 (‖x‖₁) dari vektor x = (%s) adalah?", joinInts(v)), options: buildOptions(float64(l1), cands, optionCount, false)}
	case 1:
		return question{prompt: fmt.Sprintf("Norma L2 (‖x‖₂) dari vektor x = (%s) adalah?", joinInts(v)), options: buildOptions(float64(l2), cands, optionCount, false)}
	default:
		return question{prompt: fmt.Sprintf("Norma L∞ (‖x‖∞) dari vektor x = (%s) adalah?", joinInts(v)), options: buildOptions(float64(linf), cands, optionCount, false)}
	}
}

func genKMatrixChain(optionCount int) question {
	d := []int{between(2, 20), between(2, 20), between(2, 20), between(2, 20)}
	left := d[0]*d[1]*d[2] + d[0]*d[2]*d[3]  // (AB)C
	right := d[1]*d[2]*d[3] + d[0]*d[1]*d[3] // A(BC)
	best := min(left, right)
	return question{
		prompt:  fmt.Sprintf("Matriks A (%d×%d), B (%d×%d), dan C (%d×%d). Banyak perkalian skalar minimum untuk menghitung ABC adalah?", d[0], d[1], d[1], d[2], d[2], d[3]),
		options: buildOptions(float64(best), []float64{float64(max(left, right)), float64(d[0] * d[1] * d[2] * d[3]), float64(d[0] * d[3]), float64(best + d[0])}, optionCount, false),
	}
}

func genKPCAVariance(optionCount int) question {
	for {
		k := between(4, 5)
		eig := make([]int, k)
		total := 0
		for i := range eig {
			eig[i] = between(1, 30)
			total += eig[i]
		}
		sort.Sort(sort.Reverse(sort.IntSlice(eig)))
		top := between(1, 2)
		sum := eig[0]
		if top == 2 {
			sum += eig[1]
		}
		pct := float64(sum) / float64(total) * 100
		if !isExact2(cleanFloat(pct)) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Nilai eigen matriks kovariansi pada PCA adalah %s. Persentase variansi yang dijelaskan oleh %d komponen utama pertama adalah? (%%)", joinInts(eig), top),
			options: buildOptions(round2(pct), []float64{round2(float64(eig[0]) / float64(total) * 100), float64(sum), round2(100 - pct), round2(pct + 5)}, optionCount, false),
		}
	}
}

func genKLeastSquaresOrigin(optionCount int) question {
	for {
		n := between(3, 4)
		pairs := make([]string, n)
		sxy, sxx, sy, sx := 0, 0, 0, 0
		for i := range n {
			x, y := between(1, 6), between(-5, 20)
			pairs[i] = fmt.Sprintf("(%d, %d)", x, y)
			sxy += x * y
			sxx += x * x
			sx += x
			sy += y
		}
		w := float64(sxy) / float64(sxx)
		if !isExact2(w) {
			continue
		}
		return question{
			prompt:  fmt.Sprintf("Model y = wx (tanpa intersep) dicocokkan ke data %s dengan metode kuadrat terkecil. Nilai w adalah?", strings.Join(pairs, ", ")),
			options: buildOptions(round2(w), []float64{round2(float64(sy) / float64(sx)), float64(sxy), round2(w + 1), -round2(w)}, optionCount, true),
		}
	}
}

func genKMSE(optionCount int) question {
	n := between(4, 5)
	actual, pred := make([]int, n), make([]int, n)
	sse, sae := 0, 0
	for i := range n {
		actual[i] = between(0, 20)
		pred[i] = actual[i] + between(-4, 4)
		e := actual[i] - pred[i]
		sse += e * e
		sae += absInt(e)
	}
	mse := float64(sse) / float64(n)
	return question{
		prompt:  fmt.Sprintf("Nilai aktual %s dan prediksi model %s. Mean squared error (MSE) adalah?", joinInts(actual), joinInts(pred)),
		options: buildOptions(mse, []float64{float64(sae) / float64(n), float64(sse), round2(math.Sqrt(mse)), mse + 1}, optionCount, false),
	}
}

func genKFrobenius(optionCount int) question {
	rows, cols := 2, between(2, 3)
	m := make([][]int, rows)
	sq, abs, trace := 0, 0, 0
	for i := range m {
		m[i] = make([]int, cols)
		for j := range m[i] {
			m[i][j] = between(-5, 5)
			sq += m[i][j] * m[i][j]
			abs += absInt(m[i][j])
		}
		trace += m[i][i]
	}
	return question{
		prompt:  fmt.Sprintf("Nilai ‖A‖_F² (kuadrat norma Frobenius) dari A = %s (baris dipisah titik koma) adalah?", formatMatrix(m)),
		options: buildOptions(float64(sq), []float64{float64(abs), float64(trace * trace), round2(math.Sqrt(float64(sq))), float64(sq + 1)}, optionCount, false),
	}
}
