# project-group-61
project-group-61 created by GitHub Classroom.
This program is designed to control n elevators across m floors. It is based on a peer-to-peer architecture, and uses several modules to achieve its functionality.

## Packages
The module elevatorproject includes the following main packages
### worldview
The worldview package is responsible for constructing a understanding of all current states and requests of all peers on the network. It uses information from the peerview package to enable this functionallity. 
### peerview //Kanskje rename til requestdistributor?
The peerview package is responsible for updating a peer's understanding of all other peers hall and cab requests. The package represents what this peer knows about all other peers, and uses this information to create a mutual understanding between all peers of all requests placed on the network. This module distributes all requests placed on the network using cyclic counters.
### requestassigner
The requestassigner package uses infromation from the worldview package to calculate which elevator is best suited to handle which requests. All requests placed on the system is reassigned every time a new request enters the system.

### singleelevtor
The singleelevator package is responsible for controlling the physical elevator.
### lamps
The lamps package is responsible for turning on and off all lamps on this elevator. 










