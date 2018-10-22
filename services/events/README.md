# Stellar Event Service
Stellar provides an internal distributed event system for services to communicate.
Stellar uses the [NATS](https://nats.io/) scalable event service.  It is simple and
straightforward making an incredibly reliable messaging system.  One example
of how this is used is with the Application service.  It publishes a message whenever
an application is updated (created, removed, etc).  The proxy service subscribes to
the application subject and reloads the proxy service whenever there is an update.

Stellar uses the [typeurl](https://github.com/containerd/typeurl) package from containerd
to marshal an `Any` type used in the Stellar event message.

Here is an example of using the Stellar Event service:

```go
stream, err := c.EventsService().Subscribe(context.Background(), &eventsapi.SubscribeRequest{
	Subject: "stellar.>",
})
if err != nil {
	logrus.WithError(err).Error("error subscribing to application events")
	return
}

for {
	evt, err := stream.Recv()
	if err != nil {
		logrus.WithError(err).Error("error subscribing to application events")
		return
	}

	logrus.Infof("event %+v received", evt)
}
```

To publish an event, there is a helper function, `PublishEvent`:

```go
c, err := s.client()
if err != nil {
	return err
}
defer c.Close()

if err := stellar.PublishEvent(c, "subject", v); err != nil {
	return err
}
```
