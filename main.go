package main

import "beacon/beaconImp"

func main() {
	beacon := beaconImp.Beacon{}
	beacon.Init()
	beacon.Start()
}
