# project-group-61
project-group-61 created by GitHub Classroom.
This program is designed to control n elevators across m floors. It is based on a pure peer-to-peer architecture, and uses several modules to achieve its functionality.

## Packages
The module elevatorproject includes the following main packages
### worldview
The worldview package is responsible for constructing a understanding of all current states and requests of all peers on the network. It uses information from the peerview package to enable this functionallity. 

### peerview //Kanskje rename til requestdistributor?
The peerview package is responsible for updating a peer's understanding of all other peers hall and cab requests. The package represents what this peer knows about all other peers, and uses this information to create a mutual understanding between all peers of all requests placed on the network. This module distributes all requests placed on the network using cyclic counters.
### requestassigner
The requestassigner package uses infromation from the worldview package to calculate which elevator is best suited to handle which requests. All requests on the system is reassigned every time a new request enters the system.

### singleelevtor
The singleelevator package is responsible for controlling the physical elevator.
### lamps



## Overview

This program uses several modules/packages to achieve its functionality. Here is a brief summary of each module and what it does:

- [`worldview`](#worldview)
- [`peerview`](#peerview): Responsible for handling requests from the elevators and assigning them to a specific elevator.
- [`requestassigner`](#requestassigner): Responsible for assigning requests to the appropriate elevator based on its current location and availability.
- [`singleelevator`](#singleelevator): Responsible for controlling the behavior of a single elevator.
- [`lamps`](#lamps): Responsible for controlling the lamps in the elevator to indicate the current floor and direction of travel.




## `worldview`

The `worldview` module is responsible for maintaining the state of the elevators and their current locations.

## `peerview`

The `peerview` module is responsible for handling requests from the elevators and assigning them to a specific elevator.

## `requestassigner`

The `requestassigner` module is responsible for assigning requests to the appropriate elevator based on its current location and availability.

## `singleelevator`

The `singleelevator` module is responsible for controlling the behavior of a single elevator.

## `lamps`

The `lamps` module is responsible for controlling the lamps in the elevator to indicate the current floor and direction of travel.

## Contributors

This program was written by [Your Name]. If you have any questions or comments, please contact me at [Your Email].










