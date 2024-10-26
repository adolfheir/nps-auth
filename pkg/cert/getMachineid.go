package cert

import (
	"sync"

	"github.com/denisbrodbeck/machineid"
)

var (
	MachineID string
	once      sync.Once
)

func GetMachineID() string {
	once.Do(func() {
		id, err := machineid.ProtectedID("nps-auth")
		if err != nil {
			panic(err)
		}
		MachineID = id
	})
	return MachineID
}
