package coursesvc

import "testing"

func TestApplyReviewStripsAnswersOnResultOnly(t *testing.T) {
	newRes := func() *SubmitResult {
		return &SubmitResult{Results: []QuestionResult{{Correct: false, Expected: "let", Explanation: "x", Feedback: "y"}}}
	}

	res := newRes()
	applyReview(res, ReviewSetting{Mode: ReviewResultOnly, Seconds: 5})
	if r := res.Results[0]; r.Expected != "" || r.Explanation != "" || r.Feedback != "" {
		t.Fatalf("result_only masih bocorin kunci: %+v", r)
	}

	res = newRes()
	applyReview(res, ReviewSetting{Mode: ReviewTimed, Seconds: 3})
	if res.Results[0].Expected != "let" || res.Review.Seconds != 3 {
		t.Fatalf("timed harus tetep ngirim pembahasan: %+v", res)
	}
}

func TestReviewSettingValid(t *testing.T) {
	for _, r := range []ReviewSetting{{"full", 5}, {"timed", 1}, {"result_only", 120}} {
		if !r.valid() {
			t.Errorf("%+v harusnya valid", r)
		}
	}
	for _, r := range []ReviewSetting{{"hidden", 5}, {"timed", 0}, {"full", 121}} {
		if r.valid() {
			t.Errorf("%+v harusnya gak valid", r)
		}
	}
}
