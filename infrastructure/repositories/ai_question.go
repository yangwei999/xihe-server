package repositories

import (
	"github.com/opensourceways/xihe-server/domain"
	"github.com/opensourceways/xihe-server/domain/repository"
)

type AIQuestionMapper interface {
	GetCompetitorAndSubmission(string, string) (
		bool, int, QuestionSubmissionDO, error,
	)

	SaveCompetitor(string, *CompetitorInfoDO) error

	InsertSubmission(string, *QuestionSubmissionDO) (string, error)
	UpdateSubmission(string, *QuestionSubmissionDO) error
	GetSubmission(qid, competitor, date string) (QuestionSubmissionDO, error)

	GetQuestions(pool string, choice, completion []int) (
		[]ChoiceQuestionDO,
		[]CompletionQuestionDO, error,
	)

	GetResult(string) ([]QuestionSubmissionInfoDO, error)
}

func NewAIQuestionRepository(mapper AIQuestionMapper) repository.AIQuestion {
	return aiquestion{mapper}
}

type aiquestion struct {
	mapper AIQuestionMapper
}

func (impl aiquestion) GetCompetitorAndSubmission(qid string, competitor domain.Account) (
	isCompetitor bool, score int,
	submission domain.QuestionSubmission,
	err error,
) {
	isCompetitor, score, v, err := impl.mapper.GetCompetitorAndSubmission(
		qid, competitor.Account(),
	)
	if err != nil {
		err = convertError(err)

		return
	}

	if v.Id != "" {
		err = v.toQuestionSubmission(&submission)
	}

	return
}

func (impl aiquestion) SaveCompetitor(qid string, competitor *domain.CompetitorInfo) error {
	do := new(CompetitorInfoDO)
	toCompetitorInfoDO(competitor, do)

	if err := impl.mapper.SaveCompetitor(qid, do); err != nil {
		return convertError(err)
	}

	return nil
}

func (impl aiquestion) GetQuestions(pool string, choice, completion []int) (
	[]domain.ChoiceQuestion, []domain.CompletionQuestion, error,
) {
	return impl.mapper.GetQuestions(pool, choice, completion)
}

func (impl aiquestion) SaveSubmission(qid string, v *domain.QuestionSubmission) (string, error) {
	do := new(QuestionSubmissionDO)
	impl.toQuestionSubmissionDO(v, do)

	if v.Id == "" {
		return impl.mapper.InsertSubmission(qid, do)
	}

	err := impl.mapper.UpdateSubmission(qid, do)

	return v.Id, err
}

func (impl aiquestion) GetSubmission(qid string, user domain.Account, date string) (
	submission domain.QuestionSubmission, err error,
) {
	v, err := impl.mapper.GetSubmission(qid, user.Account(), date)
	if err != nil {
		err = convertError(err)

		return
	}

	err = v.toQuestionSubmission(&submission)

	return
}

func (impl aiquestion) GetResult(qid string) ([]domain.QuestionSubmissionInfo, error) {
	v, err := impl.mapper.GetResult(qid)
	if err != nil {
		return nil, err
	}

	r := make([]domain.QuestionSubmissionInfo, len(v))
	for i := range v {
		if err = v[i].toQuestionSubmissionInfo(&r[i]); err != nil {
			return nil, err
		}
	}

	return r, nil
}
