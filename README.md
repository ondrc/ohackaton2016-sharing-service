O Hackaton 2016 / Sharing Service
===============================

Sharing service is a service that allows users to:

- offer items (e.g. beamer, car, drum set, ...) for sharing or rental to other
  users
    - an item is offered for a given period, at given location
- query available itms
    - by location, time and category  
- book an offered item
    - booked item shall be withdrawn by the user at given location and returned
      back at the same location in time

The solution attempts to use Event Sourcing & CQRS architectural pattern.
It is separated into 3 micro-services which communicate via Google
Cloud PubSub service (publish subscribe message bus / event sourcing):

- registration & booking service (RegistrationService.go) - command part
    - accepts item registration (offers from users)
    - accepts item booking
    - publishes appropriate messages to "events " topic on the PubSub service
- query service (QueryServiceMain.go) - query part
    - keeps set of active offerings for querying
    - subscribes to "events" topic on the PubSub service
    - updates the state (set of offerings) based on received events
        - the state is maintained in memory only, could be offloaded to some database eventually
    - on startup, it reads the persisted events from data store to reconstruct
      initial state (not implemented yet)
- data storage service (DataStorageService.go) - persistence
    - subscribes to "events" and stores them in Google Cloud Data Store
    - events from the Data Store can be replayed to reconstruct state (not implemented)

The following image illustrates the topology:

![topology](doc/topology.jpg "Topology")

Not implemented parts:
- persistence
   - we planned to persist events into Google Cloud Data Store but we are facing
     last minute issues with that
- frontent
   - we planned to implement some basic web front end, but we did not managed to
     do so
   - the service is only available as set of externally available web services
       - curl

Testing commands in curl:
-------------------------

TODO



      

