package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"gitlab.com/Sh00ty/hootydb/internal"
	"gitlab.com/Sh00ty/hootydb/internal/perf"
	"gitlab.com/Sh00ty/hootydb/internal/utils/logger"
)

type ammoGen struct {
	keys []int
	log  internal.Logger
}

func (a *ammoGen) GetAmmo() perf.Ammo {
	keyPtr := rand.Int31n(int32(len(a.keys)))
	val := int(time.Now().Second())

	body := strings.NewReader(fmt.Sprintf("{\"key\":\"%d\", \"value\":\"%d\"}", a.keys[keyPtr], val))
	return perf.Ammo{
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body:   body,
		Path:   "db/set/",
		Method: "POST",
	}
}

func main() {
	ctx := context.Background()
	log := logger.NewLogger("perf", "set_perf")
	ag := &ammoGen{
		log:  log,
		keys: []int{1, 2, 3},
	}
	perf.NewPerfTesting(ctx, ag, log)
}
