# project-group-61
project-group-61 created by GitHub Classroom.
This program is designed to control n elevators across m floors. It is based on a peer-to-peer architecture, and uses several modules to achieve its functionality.

## Packages
The module elevatorproject includes the following main packages
### worldview
The worldview package is responsible for constructing a understanding of all current states and requests of all peers on the network. It uses information from the peerview package to enable this functionality. 
### peerview
The peerview package is a module responsible for managing and synchronizing elevator requests and states among multiple elevators in the peer-to-peer network. This package communicates with other peers in the network to ensure that accurate and updated information about elevator requests and states is available to all peers. The package maintains a view of the local and remote elevator requests. Central to this package is the MyPeerView struct, which represents the local peer's view of states, hall requests and cab requests for all connected peers in the network.
<!--
The peerview package is responsible for updating a peer's understanding of all other peer’s hall and cab requests. The package represents what this peer knows about all other peers and uses this information to create a mutual understanding between all peers of all requests placed on the network. This module distributes all requests placed on the network using cyclic counters.
-->
### requestassigner
The requestassigner package uses information from the worldview package to calculate which elevator is best suited to handle which requests. All requests placed on the system is reassigned every time a new request enters the system.
### singleelevtor
The singleelevator package is responsible for controlling the physical elevator.
### lamps
The lamps package is responsible for turning on and off all lamps on this elevator. 










