# Stellar Nameserver Service

The Stellar Nameserver service provides service discovery via DNS.  Every node runs a local resolver so there is a single hop to lookup a record.
This also makes it more fault tolerant.  The nameserver service uses the same core library that CoreDNS uses ([miekg/dns](https://github.com/miekg/dns)).
Upon application deployment, records will automatically be added:

```
demo.web00.stellar    A                   172.16.0.4
demo.web00.stellar    TXT                 node=ctr-00; updated=2018-09-09T18:01:31-04:00
```

If an endpoint is specified in the application, an `SRV` record will also be added:

```
demo.web00.stellar    SRV   demo.web00.stellar      service=web proto=http priority=0 weight=0 port=8080
```

To be more friendly to ops, you can also add custom records:

```
NAME:
   sctl nameserver create - create nameserver record

USAGE:
   sctl nameserver create [command options] NAME VALUE

OPTIONS:
   --type value, -t value  resource record type (A, CNAME, TXT, SRV, MX) (default: "A")
```
