GLARe
-----

**G**rpc **L**ocal **A**gent with **Re**dis protocol
=====

Usage:

1. create your server (see <example/greeter.go>)
2. connect with redis client you like, and fire

```
redis> GRPC REQUEST <ClientMapIndex> <GrpcMethodName> <ProtobufPayload>
```
