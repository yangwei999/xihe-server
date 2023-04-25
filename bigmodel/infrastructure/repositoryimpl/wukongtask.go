package repositoryimpl

import (
	"github.com/opensourceways/xihe-server/bigmodel/domain/repository"
	commondomain "github.com/opensourceways/xihe-server/common/domain"
	commonrepo "github.com/opensourceways/xihe-server/common/domain/repository"
	"github.com/opensourceways/xihe-server/common/infrastructure/pgsql"
	types "github.com/opensourceways/xihe-server/domain"
)

func NewWuKongAsyncRepo(cfg *Config) repository.WuKongAsyncTask {
	return &wukongAsyncRepoImpl{
		cli: pgsql.NewDBTable(cfg.Table.WukongRequest),
	}
}

type wukongAsyncRepoImpl struct {
	cli pgsqlClient
}

func (impl *wukongAsyncRepoImpl) GetWaitingTaskRank(user types.Account, t commondomain.Time) (r int, err error) {
	var twukong []TWukongTask

	// 1. get all task before t
	err = impl.cli.DB().
		Where("created_at > ? and status IN ?", t.Time(), []string{"waiting", "running"}).
		Find(&twukong).Error
	if err != nil {
		if impl.cli.IsRowNotFound(err) {
			err = commonrepo.NewErrorResourceNotExists(err)

			return
		}

		return
	}

	// 2. is user in task
	f1 := func(v []TWukongTask) (ok bool, t TWukongTask) {
		for i := range twukong {
			if twukong[i].User == user.Account() {
				return true, twukong[i]
			}
		}

		return
	}

	ok, task := f1(twukong)
	if !ok {
		return 0, nil
	}

	// 2. caculate rank
	f2 := func(v []TWukongTask, task TWukongTask) int {
		i := 1

		for j := range v {
			if v[j].CreatedAt < task.CreatedAt {
				i++
			}
		}

		return i
	}

	return f2(twukong, task), nil
}

func (impl *wukongAsyncRepoImpl) GetLastFinishedTask(user types.Account) (resp repository.WuKongTaskResp, err error) {
	var twukong TWukongTask

	filter := map[string]interface{}{
		fieldUserName: user.Account(),
	}

	order := "created_at DESC"

	if err = impl.cli.GetOrderOneRecord(filter, order, &twukong); err != nil {
		if impl.cli.IsRowNotFound(err) {
			err = commonrepo.NewErrorResourceNotExists(err)

			return
		}

		return
	}

	err = twukong.toWuKongTaskResp(&resp)

	return
}
