package launchpad

import (
	"context"
	"fmt"
	"time"

	"github.com/MichaelMure/git-bug/bridge/core"
	"github.com/MichaelMure/git-bug/bug"
	"github.com/MichaelMure/git-bug/cache"
	"github.com/MichaelMure/git-bug/identity"
)

type launchpadImporter struct {
	conf core.Configuration
}

func (li *launchpadImporter) Init(conf core.Configuration) error {
	li.conf = conf
	return nil
}

const keyLaunchpadID = "launchpad-id"
const keyLaunchpadLogin = "launchpad-login"

func (li *launchpadImporter) ensurePerson(repo *cache.RepoCache, owner LPPerson) (*cache.IdentityCache, error) {
	// Look first in the cache
	i, err := repo.ResolveIdentityImmutableMetadata(keyLaunchpadLogin, owner.Login)
	if err == nil {
		return i, nil
	}
	if _, ok := err.(identity.ErrMultipleMatch); ok {
		return nil, err
	}

	return repo.NewIdentityRaw(
		owner.Name,
		"",
		owner.Login,
		"",
		map[string]string{
			keyLaunchpadLogin: owner.Login,
		},
	)
}

func (li *launchpadImporter) ImportAll(ctx context.Context, repo *cache.RepoCache, since time.Time) (<-chan core.ImportResult, error) {
	out := make(chan core.ImportResult)
	lpAPI := new(launchpadAPI)

	err := lpAPI.Init()
	if err != nil {
		return nil, err
	}

	lpBugs, err := lpAPI.SearchTasks(li.conf["project"])
	if err != nil {
		return nil, err
	}

	go func() {
		for _, lpBug := range lpBugs {
			var b *cache.BugCache
			var err error

			select {
			case <-ctx.Done():
				return
			default:
				lpBugID := fmt.Sprintf("%d", lpBug.ID)
				b, err = repo.ResolveBugCreateMetadata(keyLaunchpadID, lpBugID)
				if err != nil && err != bug.ErrBugNotExist {
					out <- core.NewImportError(err, lpBugID)
					return
				}

				owner, err := li.ensurePerson(repo, lpBug.Owner)
				if err != nil {
					out <- core.NewImportError(err, lpBugID)
					return
				}

				if err == bug.ErrBugNotExist {
					createdAt, _ := time.Parse(time.RFC3339, lpBug.CreatedAt)
					b, _, err = repo.NewBugRaw(
						owner,
						createdAt.Unix(),
						lpBug.Title,
						lpBug.Description,
						nil,
						map[string]string{
							keyLaunchpadID: lpBugID,
						},
					)
					if err != nil {
						reason := fmt.Sprintf("failed to add bug id #%s", lpBugID)
						out <- core.NewImportError(err, reason)
						return
					}

					out <- core.NewImportBug(b.Id())

				} else {
					/* TODO: Update bug */
					fmt.Println("TODO: Update bug")
				}

				/* Handle messages */
				if len(lpBug.Messages) == 0 {
					reason := fmt.Sprintf("failed to fetch comments for bug #%s", lpBugID)
					out <- core.NewImportError(err, reason)
					return
				}

				// The Launchpad API returns the bug description as the first
				// comment, so skip it.
				for _, lpMessage := range lpBug.Messages[1:] {
					_, err := b.ResolveOperationWithMetadata(keyLaunchpadID, lpMessage.ID)
					if err != nil && err != cache.ErrNoMatchingOp {
						reason := fmt.Sprintf("failed to fetch comments for bug #%s", lpBugID)
						out <- core.NewImportError(err, reason)
						return
					}

					// If this comment already exists, we are probably
					// updating an existing bug. We do not want to duplicate
					// the comments, so let us just skip this one.
					// TODO: Can Launchpad comments be edited?
					if err == nil {
						continue
					}

					owner, err := li.ensurePerson(repo, lpMessage.Owner)
					if err != nil {
						out <- core.NewImportError(err, "")
						return
					}

					// This is a new comment, we can add it.
					createdAt, _ := time.Parse(time.RFC3339, lpMessage.CreatedAt)
					op, err := b.AddCommentRaw(
						owner,
						createdAt.Unix(),
						lpMessage.Content,
						nil,
						map[string]string{
							keyLaunchpadID: lpMessage.ID,
						})
					if err != nil {
						reason := fmt.Sprintf("failed to add comment to bug #%s", lpBugID)
						out <- core.NewImportError(err, reason)
						return
					}

					hash, err := op.Hash()
					if err != nil {
						out <- core.NewImportError(err, "")
						return
					}

					out <- core.NewImportComment(hash.String())
				}

				err = b.CommitAsNeeded()
				if err != nil {
					out <- core.NewImportError(err, "")
					return
				}
			}
		}
	}()

	return out, nil
}
