package curriculumsvc

import "context"

// AdminListCurriculum: seluruh pohon Tier -> Batch -> Challenge, buat dashboard
// admin platform nampilin & ngedit waktu tiap challenge (FSD.md bagian 5, admin).
func (s *Service) AdminListCurriculum(ctx context.Context) ([]AdminTier, error) {
	tierRows, err := s.db.Query(ctx, `
		SELECT id, code, name FROM tiers ORDER BY order_index
	`)
	if err != nil {
		return nil, err
	}
	tiers := []AdminTier{}
	tierIndex := map[string]int{}
	for tierRows.Next() {
		var t AdminTier
		if err := tierRows.Scan(&t.ID, &t.Code, &t.Name); err != nil {
			tierRows.Close()
			return nil, err
		}
		t.Batches = []AdminBatch{}
		tierIndex[t.ID] = len(tiers)
		tiers = append(tiers, t)
	}
	tierRows.Close()
	if err := tierRows.Err(); err != nil {
		return nil, err
	}

	batchRows, err := s.db.Query(ctx, `
		SELECT id, tier_id, name FROM batches ORDER BY order_index
	`)
	if err != nil {
		return nil, err
	}
	batchIndex := map[string]struct{ tierPos, batchPos int }{}
	for batchRows.Next() {
		var (
			id, tierID, name string
		)
		if err := batchRows.Scan(&id, &tierID, &name); err != nil {
			batchRows.Close()
			return nil, err
		}
		tierPos, ok := tierIndex[tierID]
		if !ok {
			continue
		}
		b := AdminBatch{ID: id, Name: name, Challenges: []AdminChallenge{}}
		tiers[tierPos].Batches = append(tiers[tierPos].Batches, b)
		batchIndex[id] = struct{ tierPos, batchPos int }{tierPos, len(tiers[tierPos].Batches) - 1}
	}
	batchRows.Close()
	if err := batchRows.Err(); err != nil {
		return nil, err
	}

	challengeRows, err := s.db.Query(ctx, `
		SELECT
			c.id, c.batch_id, c.name, c.is_exam, c.time_limit_seconds,
			c.question_count_required, c.pass_threshold_percent, c.option_count,
			(SELECT count(*) FROM questions q WHERE q.challenge_id = c.id AND q.status = 'published')
		FROM challenges c
		ORDER BY c.order_index
	`)
	if err != nil {
		return nil, err
	}
	defer challengeRows.Close()

	for challengeRows.Next() {
		var (
			c       AdminChallenge
			batchID *string
		)
		if err := challengeRows.Scan(
			&c.ID, &batchID, &c.Name, &c.IsExam, &c.TimeLimitSeconds,
			&c.QuestionCountRequired, &c.PassThresholdPercent, &c.OptionCount, &c.QuestionBankSize,
		); err != nil {
			return nil, err
		}
		if batchID == nil {
			// Challenge tier "umum" nempel ke category, bukan batch (lihat
			// migrations/0005_puzzle_categories.sql) -- pohon Tier->Batch->Challenge
			// di halaman ini belum nampung itu, jadi dilewatin dulu.
			continue
		}
		pos, ok := batchIndex[*batchID]
		if !ok {
			continue
		}
		tiers[pos.tierPos].Batches[pos.batchPos].Challenges = append(
			tiers[pos.tierPos].Batches[pos.batchPos].Challenges, c,
		)
	}
	if err := challengeRows.Err(); err != nil {
		return nil, err
	}

	return tiers, nil
}

// AdminListCategoriesByTier: pohon Category -> Challenge buat tier yang gak
// pake batch (mis. "umum" -- lihat migrations/0005_puzzle_categories.sql).
// Dipisah dari AdminListCurriculum karena bentuk pohonnya beda (2 level, bukan
// 3), tapi AdminChallenge-nya dipakai bareng sama field yang sama persis.
func (s *Service) AdminListCategoriesByTier(ctx context.Context, tierCode string) ([]AdminCategory, error) {
	catRows, err := s.db.Query(ctx, `
		SELECT c.id, c.code, c.name
		FROM categories c
		JOIN tiers t ON t.id = c.tier_id
		WHERE t.code = $1
		ORDER BY c.order_index
	`, tierCode)
	if err != nil {
		return nil, err
	}
	categories := []AdminCategory{}
	catIndex := map[string]int{}
	for catRows.Next() {
		var c AdminCategory
		if err := catRows.Scan(&c.ID, &c.Code, &c.Name); err != nil {
			catRows.Close()
			return nil, err
		}
		c.Challenges = []AdminChallenge{}
		catIndex[c.ID] = len(categories)
		categories = append(categories, c)
	}
	catRows.Close()
	if err := catRows.Err(); err != nil {
		return nil, err
	}

	challengeRows, err := s.db.Query(ctx, `
		SELECT
			ch.id, ch.category_id, ch.name, ch.is_exam, ch.time_limit_seconds,
			ch.question_count_required, ch.pass_threshold_percent, ch.option_count,
			(SELECT count(*) FROM questions q WHERE q.challenge_id = ch.id AND q.status = 'published')
		FROM challenges ch
		JOIN categories c ON c.id = ch.category_id
		JOIN tiers t ON t.id = c.tier_id
		WHERE t.code = $1
		ORDER BY ch.order_index
	`, tierCode)
	if err != nil {
		return nil, err
	}
	defer challengeRows.Close()

	for challengeRows.Next() {
		var (
			ch         AdminChallenge
			categoryID string
		)
		if err := challengeRows.Scan(
			&ch.ID, &categoryID, &ch.Name, &ch.IsExam, &ch.TimeLimitSeconds,
			&ch.QuestionCountRequired, &ch.PassThresholdPercent, &ch.OptionCount, &ch.QuestionBankSize,
		); err != nil {
			return nil, err
		}
		pos, ok := catIndex[categoryID]
		if !ok {
			continue
		}
		categories[pos].Challenges = append(categories[pos].Challenges, ch)
	}
	if err := challengeRows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

// AdminUpdateChallengeTiming: satu-satunya field yang bisa diedit dari dashboard
// admin di iterasi ini — durasi waktu (FSD.md belum punya endpoint ini, request
// eksplisit dari user di sesi ini: "time yang per soal itu bisa diedit di dashboard").
func (s *Service) AdminUpdateChallengeTiming(ctx context.Context, challengeID string, timeLimitSeconds int) error {
	_, err := s.db.Exec(ctx, `
		UPDATE challenges SET time_limit_seconds = $1, updated_at = now() WHERE id = $2
	`, timeLimitSeconds, challengeID)
	return err
}
