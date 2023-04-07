// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"net/http"

	"github.com/ava-labs/blobvm/chain"
)

type RecentActivityReply struct {
	Activity []*chain.Activity `serialize:"true" json:"activity"`
}

func (svc *PublicService) RecentActivity(_ *http.Request, _ *struct{}, reply *RecentActivityReply) error {
	cs := uint64(svc.vm.config.ActivityCacheSize)
	if cs == 0 {
		return nil
	}

	// Sort results from newest to oldest
	start := svc.vm.activityCacheCursor
	i := start
	activity := []*chain.Activity{}
	for i > 0 && start-i < cs {
		i--
		item := svc.vm.activityCache[i%cs]
		if item == nil {
			break
		}
		activity = append(activity, item)
	}
	reply.Activity = activity
	return nil
}
