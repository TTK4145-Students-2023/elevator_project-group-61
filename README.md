# project-group-61
project-group-61 created by GitHub Classroom.

This program is designed to control n elevators across m floors. It is based on a peer-to-peer architecture, and uses several modules to achieve its functionality.

## Usage

To run the program, you can either run the executable file or run the `main.go` file directly. The program requires two command line arguments:

- `id`: The ID of the elevator you want to control (an integer value).
- `port`: The port number to use for the network connection (an integer value).

To run the program using the executable file, use the following command:
./main -id=<elevatorID> -port=<portNumber>
 
To run the program using the `main.go` file, use the following command:
go run main.go -id=<elevatorID> -port=<portNumber>
Replace `<elevatorID>` and `<portNumber>` with the values you want to use for the elevator ID and port number, respectively.

## Program architecture
This image is a illustartion of how the modules in the program interact.
![alt text](https://github.com/TTK4145-Students-2023/project-group-61/blob/code-quality/ProgramArchitecture.jpeg)

## Packages
The module elevatorproject includes the following main packages
### worldview
The worldview package is responsible for constructing a understanding of all current states and requests of all peers on the network. It uses information from the peerview package to enable this functionality. 
### peerview
The peerview package is a module responsible for managing and synchronizing elevator requests and states among multiple elevators in the peer-to-peer network. This package communicates with other peers in the network to ensure that accurate and updated information about elevator requests and states is available to all peers. The package maintains a view of the local and remote elevator requests. Central to this package is the MyPeerView struct, which represents the local peer's view of states, hall requests and cab requests for all connected peers in the network.

### requestassigner
The requestassigner package uses information from the worldview package to calculate which elevator is best suited to handle which requests. All requests placed on the system are reassigned every time a there is a change of state in the system.
### singleelevtor
The singleelevator package is responsible for controlling the physical elevator, executing the assigned orders and handling logic for updating what orders are new, what orders are finished, and what state the elevator is in.
### lamps
The lamps package is responsible for turning on and off all lamps on this elevator. 


