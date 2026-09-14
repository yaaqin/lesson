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
			batchID string
		)
		if err := challengeRows.Scan(
			&c.ID, &batchID, &c.Name, &c.IsExam, &c.TimeLimitSeconds,
			&c.QuestionCountRequired, &c.PassThresholdPercent, &c.OptionCount, &c.QuestionBankSize,
		); err != nil {
			return nil, err
		}
		pos, ok := batchIndex[batchID]
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

// AdminUpdateChallengeTiming: satu-satunya field yang bisa diedit dari dashboard
// admin di iterasi ini — durasi waktu (FSD.md belum punya endpoint ini, request
// eksplisit dari user di sesi ini: "time yang per soal itu bisa diedit di dashboard").
func (s *Service) AdminUpdateChallengeTiming(ctx context.Context, challengeID string, timeLimitSeconds int) error {
	_, err := s.db.Exec(ctx, `
		UPDATE challenges SET time_limit_seconds = $1, updated_at = now() WHERE id = $2
	`, timeLimitSeconds, challengeID)
	return err
}
