// Copyright 2023 lukemakhanu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package liveScore

import (
	"context"

	"github.com/lukemakhanu/veimu_apps/internal/domain/liveScoreStatuses"
)

// LiveScoreFetcher : returns live scores from kiron.
type LiveScoreFetcher interface {
	GetLiveScores(ctx context.Context) ([]liveScoreStatuses.ScoreLine, error)
}
