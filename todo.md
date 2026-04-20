# Solution: mountContext recovery on driver startup

The core issue is that mountContext is stored only in memory (nodeserver.go), so when the driver restarts, it loses track of existing mounts. 
While the driver has some recovery logic in prepareTargetDirectory(nodeserver.go) , it only handles corrupted mounts, not driver restarts.

## TODO
- create a mechanism to persist mount metadata to survive driver restarts (save & load state functions on the NodeServer struct)
- then implement startup recovery that scans for existing mounts - ns.recoverExistingMounts function
- call ns.recoverExistingMounts during driver initialization in main.go
- finally update NodePublishVolume and NodeUnpublishVolume in nodeserver.go to persist/cleanup state