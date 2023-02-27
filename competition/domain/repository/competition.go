package repository

import (
	"github.com/opensourceways/xihe-server/competition/domain"
	types "github.com/opensourceways/xihe-server/domain"
)

type CompetitionListOption struct {
	CompetitionIds []string
	Status         domain.CompetitionStatus
}

type Competition interface {
	FindCompetition(cid string) (domain.Competition, error)
	FindCompetitions(*CompetitionListOption) ([]domain.CompetitionSummary, error)

	FindScoreOrder(cid string) (domain.CompetitionScoreOrder, error)
}

type PlayerVersion struct {
	Player  *domain.Player
	Version int
}

type Player interface {
	SaveTeamName(*domain.Player, int) error

	AddPlayer(*domain.Player) error

	DeletePlayer(p *domain.Player, version int) error

	AddMember(team PlayerVersion, member PlayerVersion) error

	CompetitorsCount(cid string) (int, error)

	FindPlayer(cid string, a types.Account) (domain.Player, int, error)

	FindCompetitionsUserApplied(types.Account) ([]string, error)

	SavePlayer(p *domain.Player, version int) error

	ResumePlayer(cid string, a types.Account) (err error)
}

type Work interface {
	SaveWork(*domain.Work) error
	SaveRepo(*domain.Work, int) error
	AddSubmission(*domain.Work, *domain.PhaseSubmission, int) error
	SaveSubmission(*domain.Work, *domain.PhaseSubmission) error

	FindWork(domain.WorkIndex, domain.CompetitionPhase) (domain.Work, int, error)
	FindWorks(cid string) ([]domain.Work, error)
}
