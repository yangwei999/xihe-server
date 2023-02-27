package repositoryimpl

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/xihe-server/competition/domain"
	"github.com/opensourceways/xihe-server/competition/domain/repository"
	types "github.com/opensourceways/xihe-server/domain"
	repoerr "github.com/opensourceways/xihe-server/domain/repository"
)

func NewPlayerRepo(m mongodbClient) repository.Player {
	return playerRepoImpl{m}
}

type playerRepoImpl struct {
	cli mongodbClient
}

func (impl playerRepoImpl) docFilter(cid string) bson.M {
	return bson.M{
		fieldCid:     cid,
		fieldEnabled: true,
	}
}

func (impl playerRepoImpl) playerFilter(p *domain.Player) (bson.M, error) {
	v, err := impl.cli.ObjectIdFilter(p.Id)
	if err != nil {
		return nil, err
	}

	filter := impl.docFilter(p.CompetitionId)
	for k := range v {
		filter[k] = v[k]
	}

	return filter, nil
}

func (impl playerRepoImpl) docFilterByUser(cid string, a types.Account) bson.M {
	filter := impl.docFilter(cid)

	impl.cli.AppendElemMatchToFilter(
		fieldCompetitors, true,
		bson.M{fieldAccount: a.Account()}, filter,
	)

	return filter
}

// AddPlayer
func (impl playerRepoImpl) AddPlayer(p *domain.Player) error {
	return impl.insertPlayer(p)
}

func (impl playerRepoImpl) DeletePlayer(p *domain.Player, version int) error {
	return impl.updateEnabledOfPlayer(p, false, version)
}

func (repo playerRepoImpl) genPlayerDoc(p *domain.Player) (bson.M, error) {
	cs := make([]dCompetitor, p.CompetitorsCount())
	for i, m := range p.Members() {
		cs[i+1] = toCompetitorDoc(&m)
	}
	cs[0] = toCompetitorDoc(&p.Leader)

	obj := dPlayer{
		CompetitionId: p.CompetitionId,
		Competitors:   cs,
		Leader:        p.Leader.Account.Account(),
		Enabled:       true,
	}
	if p.IsATeam() {
		obj.TeamName = p.Team.Name.TeamName()
	}

	doc, err := genDoc(&obj)

	return doc, err
}

func (impl playerRepoImpl) insertPlayer(p *domain.Player) error {
	doc, err := impl.genPlayerDoc(p)
	if err != nil {
		return err
	}
	doc[fieldVersion] = 0

	f := func(ctx context.Context) error {
		filter := impl.docFilterByUser(p.CompetitionId, p.Leader.Account)

		if p.IsATeam() {
			v := impl.docFilter(p.CompetitionId)
			v[fieldTeamName] = p.Team.Name.TeamName()

			filter = bson.M{"$or": bson.A{filter, v}}
		}

		_, err := impl.cli.NewDocIfNotExist(ctx, filter, doc)

		return err
	}

	if err = withContext(f); err != nil {
		if impl.cli.IsDocExists(err) {
			err = repoerr.NewErrorDuplicateCreating(err)
		}
	}

	return err
}

func (impl playerRepoImpl) updateEnabledOfPlayer(p *domain.Player, enable bool, version int) error {
	return impl.update(p, bson.M{fieldEnabled: enable}, version)
}

func (impl playerRepoImpl) update(p *domain.Player, doc bson.M, version int) error {
	filter, err := impl.playerFilter(p)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) error {
		return impl.cli.UpdateDoc(ctx, filter, doc, mongoCmdSet, version)
	}

	if err = withContext(f); err != nil {
		if impl.cli.IsDocNotExists(err) {
			err = repoerr.NewErrorConcurrentUpdating(err)
		}
	}

	return err
}

// SaveTeamName
func (impl playerRepoImpl) SaveTeamName(p *domain.Player, version int) error {
	return impl.update(p, bson.M{fieldTeamName: p.Team.Name.TeamName()}, version)
}

// FindPlayer
func (impl playerRepoImpl) FindPlayer(cid string, a types.Account) (
	p domain.Player, version int, err error,
) {
	var v dPlayer

	f := func(ctx context.Context) error {
		return impl.cli.GetDoc(ctx, impl.docFilterByUser(cid, a), nil, &v)
	}

	if err = withContext(f); err != nil {
		if impl.cli.IsDocNotExists(err) {
			err = repoerr.NewErrorResourceNotExists(err)
		}
	} else {
		if err = v.toPlayer(&p); err == nil {
			p.SetCurrentUser(a)

			version = v.Version
		}
	}

	return
}

// FindCompetitionsUserApplied
func (impl playerRepoImpl) FindCompetitionsUserApplied(a types.Account) (
	r []string, err error,
) {
	var v []dPlayer

	f := func(ctx context.Context) error {
		filter := impl.docFilterByUser("", a)
		delete(filter, fieldCid)

		return impl.cli.GetDocs(ctx, filter, bson.M{fieldCid: 1}, &v)
	}

	if err = withContext(f); err != nil || len(v) == 0 {
		return
	}

	r = make([]string, len(v))
	for i := range v {
		r[i] = v[i].Id.Hex()
	}

	return
}

// CompetitorsCount
func (impl playerRepoImpl) CompetitorsCount(cid string) (int, error) {
	var v []struct {
		Total int `bson:"toal"`
	}

	f := func(ctx context.Context) error {
		key := "$" + fieldCompetitors

		fields := bson.M{
			"num": bson.M{
				"$cond": bson.M{
					"if":   bson.M{"$isArray": key},
					"then": bson.M{"$size": key},
					"else": 0,
				},
			},
		}

		pipeline := bson.A{
			bson.M{"$match": impl.docFilter(cid)},
			bson.M{"$addFields": fields},
			bson.M{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": "$num"}}},
		}

		cursor, err := impl.cli.Collection().Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}

		return cursor.All(ctx, &v)
	}

	if err := withContext(f); err != nil || len(v) == 0 {
		return 0, err
	}

	return v[0].Total, nil
}

// AddMember
func (impl playerRepoImpl) AddMember(
	team repository.PlayerVersion,
	member repository.PlayerVersion,
) error {
	err := impl.updateEnabledOfPlayer(member.Player, false, member.Version)
	if err != nil {
		return err
	}

	return impl.addMember(team, member.Player)
}

func (impl playerRepoImpl) addMember(
	team repository.PlayerVersion, member *domain.Player,
) error {
	filter, err := impl.playerFilter(team.Player)
	if err != nil {
		return err
	}

	c := toCompetitorDoc(&member.Leader)
	doc, err := genDoc(&c)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) error {
		return impl.cli.UpdateDoc(
			ctx, filter,
			bson.M{fieldCompetitors: doc}, mongoCmdPush, team.Version,
		)
	}

	if err = withContext(f); err != nil {
		if impl.cli.IsDocNotExists(err) {
			err = repoerr.NewErrorConcurrentUpdating(err)
		}
	}

	return err
}

// SavePlayer
func (impl playerRepoImpl) SavePlayer(p *domain.Player, version int) error {
	filter, err := impl.playerFilter(p)
	if err != nil {
		return err
	}

	doc, err := impl.genPlayerDoc(p)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) error {
		return impl.cli.UpdateDoc(ctx, filter, doc, mongoCmdSet, version)
	}

	if err = withContext(f); err != nil {
		if impl.cli.IsDocNotExists(err) {
			err = repoerr.NewErrorConcurrentUpdating(err)
		}
	}

	return err
}

func (impl playerRepoImpl) disabledPlayerFilter(cid string, a types.Account) bson.M {
	filter := bson.M{
		fieldCid:     cid,
		fieldEnabled: false,
	}

	impl.cli.AppendElemMatchToFilter(
		fieldCompetitors, true,
		bson.M{fieldAccount: a.Account()}, filter,
	)

	return filter
}

func (impl playerRepoImpl) getVersion(cid string, a types.Account) (version int, err error) {
	filter := impl.disabledPlayerFilter(cid, a)

	var v dPlayer
	f := func(ctx context.Context) error {
		return impl.cli.GetDoc(ctx, filter, nil, &v)
	}

	if err = withContext(f); err != nil {
		return
	}

	version = v.Version

	return
}

func (impl playerRepoImpl) ResumePlayer(cid string, a types.Account) (err error) {
	version, err := impl.getVersion(cid, a)
	if err != nil {
		return
	}

	filter := impl.disabledPlayerFilter(cid, a)

	f := func(ctx context.Context) error {
		return impl.cli.UpdateDoc(ctx, filter, bson.M{fieldEnabled: true}, mongoCmdSet, version)
	}

	if err = withContext(f); err != nil {
		if impl.cli.IsDocNotExists(err) {
			err = repoerr.NewErrorConcurrentUpdating(err)
		}
	}

	return
}
