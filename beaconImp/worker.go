package beaconImp

import "beacon/log"

func (B *Beacon) Start() {
	err := B.r.Run(":8000")
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Beacon started on port 8000")
}
