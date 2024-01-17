package exchange

import (
	"context"
	"time"

	"github.com/alcionai/clues"
	"github.com/microsoftgraph/msgraph-sdk-go/models"

	"github.com/alcionai/corso/src/pkg/fault"
	"github.com/alcionai/corso/src/pkg/logger"
	"github.com/alcionai/corso/src/pkg/path"
	"github.com/alcionai/corso/src/pkg/services/m365/api"
	"github.com/alcionai/corso/src/pkg/services/m365/api/graph"
)

var (
	_ graph.ContainerResolver = &mailContainerCache{}
	_ containerRefresher      = &mailRefresher{}
)

type mailRefresher struct {
	getter containerGetter
	userID string
}

func (r *mailRefresher) refreshContainer(
	ctx context.Context,
	id string,
) (graph.CachedContainer, error) {
	c, err := r.getter.GetContainerByID(ctx, r.userID, id)
	if err != nil {
		return nil, clues.Stack(err)
	}

	f := graph.NewCacheFolder(c, nil, nil)

	return &f, nil
}

// mailContainerCache struct used to improve lookup of directories within exchange.Mail
// cache map of cachedContainers where the  key =  M365ID
// nameLookup map: Key: DisplayName Value: ID
type mailContainerCache struct {
	*containerResolver
	enumer containersEnumerator[models.MailFolderable]
	getter containerGetter
	userID string
}

// init ensures that the structure's fields are initialized.
// Fields Initialized when cache == nil:
// [mc.cache]
func (mc *mailContainerCache) init(
	ctx context.Context,
) error {
	if mc.containerResolver == nil {
		mc.containerResolver = newContainerResolver(&mailRefresher{
			userID: mc.userID,
			getter: mc.getter,
		})
	}

	return mc.populateMailRoot(ctx)
}

// populateMailRoot manually fetches directories that are not returned during Graph for msgraph-sdk-go v. 40+
// rootFolderAlias is the top-level directory for exchange.Mail.
// Action ensures that cache will stop at appropriate level.
// @error iff the struct is not properly instantiated
func (mc *mailContainerCache) populateMailRoot(ctx context.Context) error {
	f, err := mc.getter.GetContainerByID(ctx, mc.userID, api.MsgFolderRoot)
	if err != nil {
		return clues.Wrap(err, "fetching root folder")
	}

	temp := graph.NewCacheFolder(
		f,
		// Root folder doesn't store any mail messages so it isn't given any paths.
		// Giving it a path/location would cause us to add extra path elements that
		// the user doesn't see in the regular UI for Exchange.
		path.Builder{}.Append(), // path of IDs
		path.Builder{}.Append()) // display location
	if err := mc.addFolder(&temp); err != nil {
		return clues.WrapWC(ctx, err, "adding resolver dir")
	}

	return nil
}

// Populate utility function for populating the mailFolderCache.
// Number of Graph Queries: 1.
// @param baseID: M365ID of the base of the exchange.Mail.Folder
// @param baseContainerPath: the set of folder elements that make up the path
// for the base container in the cache.
func (mc *mailContainerCache) Populate(
	ctx context.Context,
	errs *fault.Bus,
	baseID string,
	baseContainerPath ...string,
) error {
	start := time.Now()

	logger.Ctx(ctx).Info("populating container cache")

	if err := mc.init(ctx); err != nil {
		return clues.Wrap(err, "initializing")
	}

	el := errs.Local()

	containers, err := mc.enumer.EnumerateContainers(
		ctx,
		mc.userID,
		"")
	ctx = clues.Add(ctx, "num_enumerated_containers", len(containers))

	if err != nil {
		return clues.WrapWC(ctx, err, "enumerating containers")
	}

	for _, c := range containers {
		if el.Failure() != nil {
			return el.Failure()
		}

		cacheFolder := graph.NewCacheFolder(c, nil, nil)

		err := mc.addFolder(&cacheFolder)
		if err != nil {
			errs.AddRecoverable(
				ctx,
				graph.Stack(ctx, err).Label(fault.LabelForceNoBackupCreation))
		}
	}

	if err := mc.populatePaths(ctx, errs); err != nil {
		return clues.Wrap(err, "populating paths")
	}

	logger.Ctx(ctx).Infow(
		"done populating container cache",
		"duration", time.Since(start))

	return el.Failure()
}
