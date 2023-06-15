package mock

import (
	"context"

	"github.com/alcionai/corso/src/internal/common/idname"
	"github.com/alcionai/corso/src/internal/common/prefixmatcher"
	"github.com/alcionai/corso/src/internal/data"
	"github.com/alcionai/corso/src/internal/operations/inject"
	"github.com/alcionai/corso/src/pkg/backup/details"
	"github.com/alcionai/corso/src/pkg/control"
	"github.com/alcionai/corso/src/pkg/fault"
	"github.com/alcionai/corso/src/pkg/path"
	"github.com/alcionai/corso/src/pkg/selectors"
)

var _ inject.BackupProducer = &Controller{}

type Controller struct {
	Collections []data.BackupCollection
	Exclude     *prefixmatcher.StringSetMatcher

	Deets *details.Details

	Err error

	Stats data.CollectionStats
}

func (ctrl Controller) ProduceBackupCollections(
	_ context.Context,
	_ idname.Provider,
	_ selectors.Selector,
	_ []data.RestoreCollection,
	_ int,
	_ control.Options,
	_ *fault.Bus,
) (
	[]data.BackupCollection,
	prefixmatcher.StringSetReader,
	bool,
	error,
) {
	return ctrl.Collections, ctrl.Exclude, ctrl.Err == nil, ctrl.Err
}

func (ctrl Controller) IsBackupRunnable(
	_ context.Context,
	_ path.ServiceType,
	_ string,
) (bool, error) {
	return true, ctrl.Err
}

func (ctrl Controller) Wait() *data.CollectionStats {
	return &ctrl.Stats
}

func (ctrl Controller) ConsumeRestoreCollections(
	_ context.Context,
	_ int,
	_ selectors.Selector,
	_ control.RestoreConfig,
	_ control.Options,
	_ []data.RestoreCollection,
	_ *fault.Bus,
) (*details.Details, error) {
	return ctrl.Deets, ctrl.Err
}